package github

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	perPageArg     = "per_page"
)

const githubMaxPageSize = 100

// Implementation detail, not important to external API.
type queryField struct {
	name  string
	value interface{}
}

// Formats a query field.
func (qf queryField) String() string {
	var value string
	switch v := qf.value.(type) {
	case string:
		value = v
	case rangeFormatter:
		value = v.RangeString()
	default:
		value = fmt.Sprint(v)
	}

	if qf.name == "" {
		return value
	}
	return fmt.Sprint(qf.name, ":", value)
}

// Example of formating a query:
// QueryWith(
//	Filename("kustomization.yaml"),
//	Filesize(RangeWithin{64, 192}),
//	Keyword("copyright"),
//	Keyword("2019"),
// ).String()
//
// Outputs "q=filename:kustomization.yaml+size:64..192+copyright+2018" which
// would search for files that have [64, 192] bytes (inclusive range) and that
// contain the keywords 'copyright' and '2019' somewhere in the file.
type Query []queryField

func QueryWith(qfs ...queryField) Query {
	return qfs
}

func (q Query) String() string {
	strs := make([]string, 0, len(q))
	for _, elem := range q {
		str := elem.String()
		if str == "" {
			continue
		}
		strs = append(strs, str)
	}

	query := strings.Join(strs, "+")
	if query == "" {
		return query
	}
	return "q=" + query
}

// Keyword takes a single word, and formats it according to the Github API.
func Keyword(k string) queryField {
	return queryField{value: k}
}

// Filesize takes a rangeFormatter and formats it according to the Github API.
func Filesize(r rangeFormatter) queryField {
	return queryField{name: "size", value: r}
}

// Filename takes a filename and formats it according to the Github API.
func Filename(f string) queryField {
	return queryField{name: "filename", value: f}
}

// Path takes a filepath and formats it according to the Github API.
func Path(p string) queryField {
	return queryField{name: "path", value: p}
}

// RequestConfig stores common variables that must be present for the queries.
// - CodeSearchRequests: ask Github to check the code indices given a query.
// - ContentsRequests: ask Github where to download a resource given a repo and a
// file path.
// - CommitsRequests: asks Github to list commits made one a file. Useful to
// determine the date of a file.
type RequestConfig struct {
	perPage uint64
}

// CodeSearchRequestWith given a list of query parameters that specify the
// (patial) query, returns a request object with the (parital) query. Must call
// the URL method to get the string value of the URL. See request.CopyWith, to
// understand why the request object is useful.
func (rc RequestConfig) CodeSearchRequestWith(query Query) request {
	req := rc.makeRequest("search/code", query)
	req.vals.Set("sort", "indexed")
	req.vals.Set("order", "desc")
	return req
}

// ContentsRequest given the repo name, and the filepath returns a formatted
// query for the Github API to find the dowload information of this filepath.
func (rc RequestConfig) ContentsRequest(fullRepoName, path string) string {
	uri := fmt.Sprintf("repos/%s/contents/%s", fullRepoName, path)
	return rc.makeRequest(uri, Query{}).URL()
}

func (rc RequestConfig) ReposRequest(fullRepoName string) string {
	uri := fmt.Sprintf("repos/%s", fullRepoName)
	return rc.makeRequest(uri, Query{}).URL()
}

// CommitsRequest given the repo name, and a filepath returns a formatted query
// for the Github API to find the commits that affect this file.
func (rc RequestConfig) CommitsRequest(fullRepoName, path string) string {
	uri := fmt.Sprintf("repos/%s/commits", fullRepoName)
	return rc.makeRequest(uri, Query{Path(path)}).URL()
}

func (rc RequestConfig) makeRequest(path string, query Query) request {
	vals := url.Values{}
	vals.Set(perPageArg, fmt.Sprint(rc.perPage))

	return request{
		url: url.URL{
			Scheme: "https",
			Host:   "api.github.com",
			Path:   path,
		},
		vals:  vals,
		query: query,
	}
}

type request struct {
	url   url.URL
	vals  url.Values
	query Query
}

// CopyWith copies the requests and adds the extra query parameters. Usefull
// for dynamically adding sizes to a filename only query without modifying it.
func (r request) CopyWith(queryParams ...queryField) request {
	cpy := r
	cpy.query = append(cpy.query, queryParams...)
	return cpy
}

// URL encodes the variables and the URL representation into a string.
func (r request) URL() string {
	// Github does not handle URL encoding properly in its API for the
	// q='...', so the query parameter is added without any encoding
	// manually.
	encoded := r.vals.Encode()
	query := r.query.String()
	sep := "&"
	if query == "" {
		sep = ""
	}
	if encoded == "" && query != "" {
		sep = "?"
	}
	r.url.RawQuery = query + sep + encoded
	return r.url.String()
}

// Allows to define a range of numbers and print it in the github range
// query format https://help.github.com/en/articles/understanding-the-search-syntax.
type rangeFormatter interface {
	RangeString() string
}

// RangeLessThan is a range of values strictly less than (<) size.
type RangeLessThan struct {
	size uint64
}

func (r RangeLessThan) RangeString() string {
	return fmt.Sprintf("<%d", r.size)
}

// RangeLessThan is a range of values strictly greater than (>) size.
type RangeGreaterThan struct {
	size uint64
}

func (r RangeGreaterThan) RangeString() string {
	return fmt.Sprintf(">%d", r.size)
}

// RangeWithin is an inclusive range from start to end.
type RangeWithin struct {
	start uint64
	end   uint64
}

func (r RangeWithin) RangeString() string {
	return fmt.Sprintf("%d..%d", r.start, r.end)
}
