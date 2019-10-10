// Package github implements the crawler.Crawler interface, getting data
// from the Github search API.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/internal/tools/crawler"
	"sigs.k8s.io/kustomize/internal/tools/doc"
	"sigs.k8s.io/kustomize/internal/tools/httpclient"
	"sigs.k8s.io/kustomize/v3/pkg/git"
	"sigs.k8s.io/kustomize/v3/pkg/pgmconfig"
)

var logger = log.New(os.Stdout, "Github Crawler: ",
	log.LstdFlags|log.LUTC|log.Llongfile)

// Implements crawler.Crawler.
type githubCrawler struct {
	client GitHubClient
	query  Query
}

type GitHubClient struct {
	RequestConfig
	retryCount uint64
	client     *http.Client
}

func NewClient(accessToken string, retryCount uint64, client *http.Client) GitHubClient {
	return GitHubClient{
		retryCount: retryCount,
		client:     client,
		RequestConfig: RequestConfig{
			perPage:     githubMaxPageSize,
			accessToken: accessToken,
		},
	}
}

func NewCrawler(accessToken string, retryCount uint64, client *http.Client,
	query Query) githubCrawler {

	return githubCrawler{
		client: GitHubClient{
			retryCount: retryCount,
			client:     client,
			RequestConfig: RequestConfig{
				perPage:     githubMaxPageSize,
				accessToken: accessToken,
			},
		},
		query: query,
	}
}

// Implements crawler.Crawler.
func (gc githubCrawler) Crawl(
	ctx context.Context, output chan<- crawler.CrawlerDocument) error {

	noETagClient := GitHubClient{
		RequestConfig: gc.client.RequestConfig,
		client:        &http.Client{Timeout: gc.client.client.Timeout},
		retryCount:    gc.client.retryCount,
	}

	// Since Github returns a max of 1000 results per query, we can use
	// multiple queries that split the search space into chunks of at most
	// 1000 files to get all of the data.
	ranges, err := FindRangesForRepoSearch(newCache(noETagClient, gc.query))
	if err != nil {
		return fmt.Errorf("could not split %v into ranges, %v\n",
			gc.query, err)
	}

	logger.Println("ranges: ", ranges)

	// Query each range for files.
	errs := make(multiError, 0)
	for _, query := range ranges {
		err := processQuery(ctx, gc.client, query, output)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (gc githubCrawler) FetchDocument(ctx context.Context, d *doc.Document) error {
	repoURL := d.RepositoryURL + "/" + d.FilePath + "?ref=" + d.DefaultBranch
	repoSpec, err := git.NewRepoSpecFromUrl(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repospec: %v", err)
	}

	url := "https://raw.githubusercontent.com/" + repoSpec.OrgRepo +
		"/" + repoSpec.Ref + "/" + repoSpec.Path

	handle := func(resp *http.Response, err error, path string) error {
		if err == nil && resp.StatusCode == http.StatusOK {
			d.IsSame = httpclient.FromCache(resp.Header)
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			d.DocumentData = string(data)
			d.FilePath = d.FilePath + path
			return nil
		}
		return err
	}
	resp, err := gc.client.GetRawUserContent(url)
	if err := handle(resp, err, ""); err == nil {
		return nil
	}

	for _, file := range pgmconfig.KustomizationFileNames {
		resp, err = gc.client.GetRawUserContent(url + "/" + file)
		err := handle(resp, err, "/"+file)
		if err != nil {
			continue
		}
	}
	return fmt.Errorf("File Not Found: %s", url)
}

func (gc githubCrawler) SetCreated(ctx context.Context, d *doc.Document) error {
	fs := GithubFileSpec{}
	fs.Repository.FullName = d.RepositoryURL + "/" + d.FilePath
	creationTime, err := gc.client.GetFileCreationTime(fs)
	if err != nil {
		return err
	}
	d.CreationTime = &creationTime
	return nil
}

func (gc githubCrawler) Match(d *doc.Document) bool {
	url := d.RepositoryURL + "/" + d.FilePath + "?ref=" + "/" +
		d.DefaultBranch
	repoSpec, err := git.NewRepoSpecFromUrl(url)
	if err != nil {
		return false
	}

	return strings.Contains(repoSpec.Host, "github.com")
}

// processQuery follows all of the pages in a query, and updates/adds the
// documents from the crawl to the datastore/index.
func processQuery(ctx context.Context, gcl GitHubClient, query string,
	output chan<- crawler.CrawlerDocument) error {

	queryPages := make(chan GithubResponseInfo)

	go func() {
		// Forward the document metadata to the retrieval channel.
		// This separation allows for concurrent requests for the code
		// search, and the retrieval portions of the API.
		err := gcl.ForwardPaginatedQuery(ctx, query, queryPages)
		if err != nil {
			// TODO(damienr74) handle this error with redis?
			logger.Println(err)
		}
		close(queryPages)
	}()

	errs := make(multiError, 0)
	errorCnt := 0
	totalCnt := 0
	for page := range queryPages {
		if page.Error != nil {
			errs = append(errs, page.Error)
			continue
		}

		for _, file := range page.Parsed.Items {
			k, err := kustomizationResultAdapter(gcl, file)
			if err != nil {
				errs = append(errs, err)
				errorCnt++
				continue
			}
			output <- k
			totalCnt++
		}

		logger.Printf("got %d files out of %d from API. %d of %d had errors\n",
			totalCnt, page.Parsed.TotalCount, errorCnt, totalCnt)
	}

	return errs
}

func kustomizationResultAdapter(gcl GitHubClient, k GithubFileSpec) (
	crawler.CrawlerDocument, error) {

	data, err := gcl.GetFileData(k)
	if err != nil {
		return nil, err
	}

	if err != nil {
		logger.Printf(
			"(error: %v) initializing to current time.\n", err)
	}

	url := gcl.ReposRequest(k.Repository.FullName)
	defaultBranch, err := gcl.GetDefaultBranch(url)
	if err != nil {
		logger.Printf(
			"(error: %v) setting default_branch to master\n", err)
		defaultBranch = "master"
	}

	doc := doc.KustomizationDocument{
		Document: doc.Document{
			DocumentData:  string(data),
			FilePath:      k.Path,
			DefaultBranch: defaultBranch,
			RepositoryURL: k.Repository.URL,
		},
	}

	return &doc, nil
}

// ForwardPaginatedQuery follows the links to the next pages and performs all of
// the queries for a given search query, relaying the data from each request
// back to an output channel.
func (gcl GitHubClient) ForwardPaginatedQuery(ctx context.Context, query string,
	output chan<- GithubResponseInfo) error {

	logger.Println("querying: ", query)
	response := gcl.parseGithubResponse(query)

	if response.Error != nil {
		return response.Error
	}

	output <- response

	for response.LastURL != "" && response.NextURL != "" {
		select {
		case <-ctx.Done():
			return nil
		default:
			response = gcl.parseGithubResponse(response.NextURL)
			if response.Error != nil {
				return response.Error
			}

			output <- response
		}
	}

	return nil
}

// GetFileData gets the bytes from a file.
func (gcl GitHubClient) GetFileData(k GithubFileSpec) ([]byte, error) {

	url := gcl.ContentsRequest(k.Repository.FullName, k.Path)

	resp, err := gcl.GetReposData(url)
	if err != nil {
		return nil, fmt.Errorf("%+v: could not get '%s' metadata: %v",
			k, url, err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%+v: could not read '%s' metadata: %v",
			k, url, err)
	}
	resp.Body.Close()

	type githubContentRawURL struct {
		DownloadURL string `json:"download_url,omitempty"`
	}
	var rawURL githubContentRawURL
	err = json.Unmarshal(data, &rawURL)
	if err != nil {
		return nil, fmt.Errorf(
			"%+v: could not get 'download_url' from '%s' response: %v",
			k, data, err)
	}

	resp, err = gcl.GetRawUserContent(rawURL.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("%+v: could not fetch file raw data '%s': %v",
			k, rawURL.DownloadURL, err)
	}

	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	return data, err
}

func (gcl GitHubClient) GetDefaultBranch(url string) (string, error) {
	resp, err := gcl.GetReposData(url)
	if err != nil {
		return "", fmt.Errorf(
			"'%s' could not get default_branch: %v", url, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(
			"could not read default_branch: %v", err)
	}

	type defaultBranch struct {
		DefaultBranch string `json:"default_branch,omitempty"`
	}
	var branch defaultBranch
	err = json.Unmarshal(data, &branch)
	if err != nil {
		return "", fmt.Errorf(
			"default_branch json malformed: %v", err)
	}

	return branch.DefaultBranch, nil
}

// GetFileCreationTime gets the earliest date of a file.
func (gcl GitHubClient) GetFileCreationTime(
	k GithubFileSpec) (time.Time, error) {

	url := gcl.CommitsRequest(k.Repository.FullName, k.Path)

	defaultTime := time.Now()

	resp, err := gcl.GetReposData(url)
	if err != nil {
		return defaultTime, fmt.Errorf(
			"%+v: '%s' could not get metadata: %v", k, url, err)
	}

	type DateSpec struct {
		Commit struct {
			Author struct {
				Date string `json:"date,omitempty"`
			} `json:"author,omitempty"`
		} `json:"commit,omitempty"`
	}

	_, lastURL := parseGithubLinkFormat(resp.Header.Get("link"))
	if lastURL != "" {
		resp, err = gcl.GetReposData(lastURL)
		if err != nil {
			return defaultTime, fmt.Errorf(
				"%+v: '%s' could not get metadata: %v",
				k, lastURL, err)
		}
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return defaultTime, fmt.Errorf(
			"%+v: failed to read metadata: %v", k, err)
	}
	earliestDate := []DateSpec{}
	err = json.Unmarshal(data, &earliestDate)
	size := len(earliestDate)
	if err != nil || size == 0 {
		return defaultTime, fmt.Errorf(
			"%+v: server response '%s' not in expected format: %v",
			k, data, err)
	}

	return time.Parse(time.RFC3339, earliestDate[size-1].Commit.Author.Date)
}

// TODO(damienr74) change the tickers to actually check api rate limits, reset
// times, and throttle requests dynamically based off of current utilization,
// instead of hardcoding the documented values, these calls are not quota'd.
// This is now especially important, since caching the API requests will reduce
// API quota use (so we can actually make more requests in the allotted time
// period).
//
// See https://developer.github.com/v3/rate_limit/ for details.
var (
	searchRateTicker  = time.NewTicker(time.Second * 2)
	contentRateTicker = time.NewTicker(time.Second * 1)
)

func throttleSearchAPI() {
	<-searchRateTicker.C
}

func throttleRepoAPI() {
	<-contentRateTicker.C
}

type multiError []error

func (me multiError) Error() string {
	size := len(me) + 2
	strs := make([]string, size)
	strs[0] = "Errors ["
	for i, err := range me {
		strs[i+1] = "\t" + err.Error()
	}
	strs[size-1] = "]"
	return strings.Join(strs, "\n")
}

type GithubFileSpec struct {
	Path       string `json:"path,omitempty"`
	Repository struct {
		API      string `json:"url,omitempty"`
		URL      string `json:"html_url,omitempty"`
		FullName string `json:"full_name,omitempty"`
	} `json:"repository,omitempty"`
}

type githubResponse struct {
	// MaxUint is reserved as a sentinel value.
	// This is the number of files that match the query.
	TotalCount uint64 `json:"total_count,omitempty"`

	// Github representation of a file.
	Items []GithubFileSpec `json:"items,omitempty"`
}

type GithubResponseInfo struct {
	*http.Response
	Parsed  *githubResponse
	Error   error
	NextURL string
	LastURL string
}

func parseGithubLinkFormat(links string) (string, string) {
	const (
		linkNext    = "next"
		linkLast    = "last"
		linkInfoURL = 1
		linkInfoRel = 2
	)

	next, last := "", ""
	linkInfo := regexp.MustCompile(`<(.*)>.*; rel="(last|next)"`)

	for _, link := range strings.Split(links, ",") {
		linkParse := linkInfo.FindStringSubmatch(link)
		if len(linkParse) != 3 {
			continue
		}

		url := linkParse[linkInfoURL]
		switch linkParse[linkInfoRel] {
		case linkNext:
			next = url
		case linkLast:
			last = url
		default:
		}
	}

	return next, last
}

func (gcl GitHubClient) parseGithubResponse(getRequest string) GithubResponseInfo {
	resp, err := gcl.SearchGithubAPI(getRequest)
	requestInfo := GithubResponseInfo{
		Response: resp,
		Error:    err,
		Parsed:   nil,
	}

	if err != nil || resp == nil {
		return requestInfo
	}

	var data []byte
	defer resp.Body.Close()
	data, requestInfo.Error = ioutil.ReadAll(resp.Body)
	if requestInfo.Error != nil {
		return requestInfo
	}

	if resp.StatusCode != http.StatusOK {
		logger.Println("query: ", getRequest)
		logger.Println("status not OK at the source")
		logger.Println("header dump", resp.Header)
		logger.Println("body dump", string(data))
		requestInfo.Error = fmt.Errorf("request rejected, status '%s'",
			resp.Status)
		return requestInfo
	}

	requestInfo.NextURL, requestInfo.LastURL =
		parseGithubLinkFormat(resp.Header.Get("link"))

	resultCount := githubResponse{
		TotalCount: math.MaxUint64,
	}
	requestInfo.Error = json.Unmarshal(data, &resultCount)
	if requestInfo.Error != nil {
		return requestInfo
	}

	requestInfo.Parsed = &resultCount

	return requestInfo

}

// SearchGithubAPI performs a search query and handles rate limitting for
// the 'code/search?' endpoint as well as timed retries in the case of abuse
// prevention.
func (gcl GitHubClient) SearchGithubAPI(query string) (*http.Response, error) {
	throttleSearchAPI()
	return gcl.getWithRetry(query)
}

// GetReposData performs a search query and handles rate limitting for
// the '/repos' endpoint as well as timed retries in the case of abuse
// prevention.
func (gcl GitHubClient) GetReposData(query string) (*http.Response, error) {
	throttleRepoAPI()
	return gcl.getWithRetry(query)
}

// User content (file contents) is not API rate limited, so there's no use in
// throttling this call.
func (gcl GitHubClient) GetRawUserContent(query string) (*http.Response, error) {
	return gcl.getWithRetry(query)
}

func (gcl GitHubClient) getWithRetry(
	query string) (resp *http.Response, err error) {

	resp, err = gcl.client.Get(query)
	retryCount := gcl.retryCount

	for err == nil &&
		resp.StatusCode == http.StatusForbidden &&
		retryCount > 0 {

		retryTime := resp.Header.Get("Retry-After")
		i, err := strconv.Atoi(retryTime)
		if err != nil {
			return resp, fmt.Errorf(
				"query '%s' forbidden without 'Retry-After'", query)
		}
		logger.Printf(
			"status forbidden, retring %d more times\n", retryCount)

		logger.Printf("waiting %d seconds before retrying\n", i)
		time.Sleep(time.Second * time.Duration(i))
		retryCount--
		resp, err = gcl.client.Get(query)
	}

	if err != nil {
		return resp, fmt.Errorf("query '%s' could not be processed, %v",
			query, err)
	}

	return resp, err
}
