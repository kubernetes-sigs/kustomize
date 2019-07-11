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

	"sigs.k8s.io/kustomize/internal/search/doc"
)

var logger = log.New(os.Stdout, "Github Crawler: ",
	log.LstdFlags|log.LUTC|log.Llongfile)

// Implements crawler.Crawler.
type githubCrawler struct {
	rc    RequestConfig
	query Query
}

func NewCrawler(
	accessToken string, retryCount uint64, query Query) githubCrawler {

	return githubCrawler{
		rc: RequestConfig{
			perPage:     githubMaxPageSize,
			retryCount:  retryCount,
			accessToken: accessToken,
		},
		query: query,
	}
}

// Implements crawler.Crawler.
func (gc githubCrawler) Crawl(
	ctx context.Context, output chan<- *doc.KustomizationDocument) error {

	// Since Github returns a max of 1000 results per query, we can use
	// multiple queries that split the search space into chunks of at most
	// 1000 files to get all of the data.
	ranges, err := FindRangesForRepoSearch(newCache(gc.rc, gc.query))
	if err != nil {
		return fmt.Errorf("could not split search into ranges, %v\n",
			err)
	}

	// Query each range for files.
	errs := make(multiError, 0)
	for _, query := range ranges {
		err := processQuery(ctx, gc.rc, query, output)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// processQuery follows all of the pages in a query, and updates/adds the
// documents from the crawl to the datastore/index.
func processQuery(ctx context.Context, rc RequestConfig, query string,
	output chan<- *doc.KustomizationDocument) error {

	queryPages := make(chan GithubResponseInfo)

	go func() {
		// Forward the document metadata to the retrieval channel.
		// This separation allows for concurrent requests for the code
		// search, and the retrieval portions of the API.
		err := ForwardPaginatedQuery(
			ctx, query, rc.retryCount, queryPages)
		if err != nil {
			// TODO(damienr74) handle this error with redis?
			logger.Println(err)
		}
		close(queryPages)
	}()

	errs := make(multiError, 0)
	for page := range queryPages {
		if page.Error != nil {
			errs = append(errs, page.Error)
			continue
		}

		for _, file := range page.Parsed.Items {
			// TODO(damienr74) This is where we'd need to
			// communicate with redis. Currently always doing a full
			// reindex of the documents. Since the documents are in
			// sorted order in each bucket, we can short circuit the
			// search when we find a file that has been seen, or we
			// can choose to selectively update files.

			k, err := kustomizationResultAdapter(rc, file)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			output <- k
		}
	}

	return errs
}

func kustomizationResultAdapter(rc RequestConfig, k GithubFileSpec) (
	*doc.KustomizationDocument, error) {

	data, err := GetFileData(rc, k)
	if err != nil {
		return nil, err
	}

	creationTime, err := GetFileCreationTime(rc, k)
	if err != nil {
		logger.Printf("(Error: %v) initializing to current time.", err)
	}

	doc := doc.KustomizationDocument{
		DocumentData:  string(data),
		FilePath:      doc.Atom(k.Path),
		RepositoryURL: doc.Atom(k.Repository.URL),
		CreationTime:  creationTime,
	}

	return &doc, nil
}

// ForwardPaginatedQuery follows the links to the next pages and performs all of
// the queries for a given search query, relaying the data from each request
// back to an output channel.
func ForwardPaginatedQuery(ctx context.Context, query string, retryCount uint64,
	output chan<- GithubResponseInfo) error {

	response := parseGithubResponse(query, retryCount)
	if response.Error != nil {
		return response.Error
	}

	output <- response

	for response.LastURL != "" && response.NextURL != "" {
		select {
		case <-ctx.Done():
			return nil
		default:
			response = parseGithubResponse(response.NextURL, retryCount)
			if response.Error != nil {
				return response.Error
			}

			output <- response
		}
	}

	return nil
}

// GetFileData gets the bytes from a file.
func GetFileData(rc RequestConfig, k GithubFileSpec) ([]byte, error) {

	url := rc.ContentsRequest(k.Repository.FullName, k.Path)

	logger.Println("content-url ", url)
	resp, err := GetReposData(url, rc.RetryCount())
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type githubContentRawURL struct {
		DownloadURL string `json:"download_url,omitempty"`
	}
	var rawURL githubContentRawURL
	err = json.Unmarshal(data, &rawURL)
	if err != nil {
		return nil, err
	}

	logger.Println("raw-data-url", rawURL.DownloadURL)
	resp, err = GetReposData(rawURL.DownloadURL, rc.RetryCount())
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// GetFileCreationTime gets the earliest date of a file.
func GetFileCreationTime(
	rc RequestConfig, k GithubFileSpec) (time.Time, error) {

	url := rc.CommitsRequest(k.Repository.FullName, k.Path)

	defaultTime := time.Now()

	logger.Println("commits-url", url)
	resp, err := GetReposData(url, rc.RetryCount())
	if err != nil {
		return defaultTime, err
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
		resp, err = GetReposData(lastURL, rc.RetryCount())
		if err != nil {
			return defaultTime, err
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	earliestDate := []DateSpec{}
	err = json.Unmarshal(data, &earliestDate)
	size := len(earliestDate)
	if err != nil || size == 0 {
		return defaultTime, err
	}

	return time.Parse(time.RFC3339, earliestDate[size-1].Commit.Author.Date)
}

// TODO(damienr74) change the tickers to actually check api rate limits, reset
// times, and throttle requests dynamically based off of current utilization,
// instead of hardcoding the documented values, these calls are not quota'd.
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

const (
	accessTokenKeyword = "access_token="
	perPageKeyword     = "per_page="
	contentSearchURL   = "https://api.github.com/repos"
	contentKeyword     = "contents"
)

type multiError []error

func (me multiError) Error() string {
	size := len(me) + 2
	strs := make([]string, size)
	strs[0] = "Errors [\n\t"
	for i, err := range me {
		strs[i+1] = err.Error()
	}
	strs[size-1] = "\n]"
	return strings.Join(strs, "\n\t")
}

type GithubFileSpec struct {
	Path       string `json:"path,omitempty"`
	Repository struct {
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

func parseGithubResponse(getRequest string, retryCount uint64) GithubResponseInfo {
	resp, err := SearchGithubAPI(getRequest, retryCount)
	requestInfo := GithubResponseInfo{
		Response: resp,
		Error:    err,
		Parsed:   nil,
	}

	if err != nil || resp == nil {
		return requestInfo
	}

	var data []byte
	data, requestInfo.Error = ioutil.ReadAll(resp.Body)
	if requestInfo.Error != nil {
		return requestInfo
	}

	if resp.StatusCode != http.StatusOK {
		logger.Println("Query: ", getRequest)
		logger.Println("Status not OK at the source")
		logger.Println("Header Dump", resp.Header)
		logger.Println("Body Dump", string(data))
		requestInfo.Error = fmt.Errorf("Request Rejected, Status '%s'",
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
func SearchGithubAPI(query string, retryCount uint64) (*http.Response, error) {
	throttleSearchAPI()
	return getWithRetry(query, retryCount)
}

// GetReposData performs a search query and handles rate limitting for
// the '/repos' endpoint as well as timed retries in the case of abuse
// prevention.
func GetReposData(query string, retryCount uint64) (*http.Response, error) {
	throttleRepoAPI()
	return getWithRetry(query, retryCount)
}

func getWithRetry(
	query string, retryCount uint64) (resp *http.Response, err error) {

	resp, err = http.Get(query)

	for err == nil &&
		resp.StatusCode == http.StatusForbidden &&
		retryCount > 0 {

		retryTime := resp.Header.Get("Retry-After")
		i, err := strconv.Atoi(retryTime)
		if err != nil {
			return resp, fmt.Errorf("Forbidden without 'Retry-After'")
		}
		logger.Printf(
			"Status Forbidden, retring %d more times\n", retryCount)

		logger.Printf("Waiting %d seconds before retrying\n", i)
		time.Sleep(time.Second * time.Duration(i))
		retryCount--
		resp, err = http.Get(query)
	}

	return resp, err
}
