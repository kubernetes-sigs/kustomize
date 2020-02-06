package github

// GitHub only returns at most 1000 results per search query,
// this is problematic if you want to retrieve all the results for a given
// search query. However, GitHub allows you to specify as much as you want per
// query to make things more specific. Specifically for files, GitHub allows
// you to specify their sizes with range queries. This is very convenient
// since it allows us to split the search into disjoint sets/shards of results
// from the different file size ranges.
//
// Some important factors to consider:
//
// -  These queries are rate limited by the API to roughly once query every two
//    seconds.
//
// -  The search space for file sizes is in bytes, from 0B to < 512KiB (this is
//    a huge search space that cannot be probed linearly in a timely manner if
//    granularity is to be expected).
//
// -  If you have K files there will likely be ~K/1000 sets that you have find
//    from this search space in order to get all of the results.
//
// -  If you have O(K) sets it is unlikely that they are all of the same size,
//    since (most files are power law distributed). That means that the range
//    might be significantly smaller for 1000 small files, than it is for
//    1000 large files.
//
// -  This method is a best effort approach. There are some limitations to what
//    it can and can't do, so please note the following:
//
//    +  There may very well be a filesize that has more than 1000 results.
//       this method cannot help in this case. However, requerying over time
//       (days/weeks/months) while sorting by last indexed values may be
//       sufficient to eventually get all of the results.
//
//    +  It's possible that the github API returns inconsistent counts. This
//       is problematic in most cases, since it can cause many issues if the
//       case is not handled properly. For instance, if you requested the
//       number of files of an interval from size:0..64 and get that there
//       are 900 results, you may query at size:0..96 and get that there
//       are 800 results. To guarantee that this approach completes and does
//       not get into a query loop over the same intervals, it will retry a few
//       times and take the largest of the results or the largest previously
//       queried value from another range (in this case, the implementation
//       could decide that size:0..96 must have 900) results. This makes the
//       approach best effort even if there are no single file sizes of over
//       1000 results.
//
//
// The approach that was taken to solve this problem is the following:
//
// 1. Determine the total number of results by querying from the lower bound
//    to the upper bound (size:0..max). If there are less than 1000 files,
//    return a single range of values (size:0..max) since all results can be
//    retrieved.
//
// 2. Otherwise, set a target number of files to be 1000.
//
// 3. Binary search for the range from 0..r that provides a file count that is
//    less than or equal to the target. Once this value is found, store the
//    upper bound of range (r). If r is the same as the previous value, (or 0)
//    increase r by one (this guarantees progress, but will miss out on some
//    results).
//
// 4. Increase the target by 1000.
//
// 5. Repeat steps 3 and 4 until the target is at or exceeds the total number
//    of files.
//
//
// In general there are other ways to get all of the files from GitHub. In
// some cases it would be sufficient to just get the files that are being
// updated/indexed by github periodically to update the corpus, so this
// complicated approach does not have to be run every time. However, for
// some searches, there may be too many results on a time interval to do
// this simple update search limited to only 1000 results.
//
// There is also a more sophisticated approach that may yield better
// performance:
// -  Perform this search once and create a prior distribution of file sizes.
//    Each time you want to retrieve the results of the query, scale the
//    prior of expected ranges to the current number of files. From each
//    expected range of 1000 files, perform a exponential search to find the
//    lower bound of the range. This would likely reduce the total number
//    of queries by a significant amount since it would only have to search
//    for a small set of values around each likely range boundary.
//
// However, actually retrieving the files will be the bottleneck operation
// since the number of queries to find the ranges will be close to:
//   log2(maxFileSize) * totalResults / 1000 ~= totalResults / 50
// whereas the number of queries to actually get all of the search results
// are close to:
//   apiCallsPerResult * 10(pages) * 100(resultsPerPage) * totalResults / 1000
//   = apiCallsPerResult * totalResults.
//
// So it could very well take apiCallsPerResult * 50 times longer to actually
// fetch the results (assuming the quotas for the API calls are the same as the
// search API), than it does to perform these range searches.

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// Files cannot be more than 2^19 bytes, according to
// https://help.github.com/en/articles/searching-code#considerations-for-code-search
const (
	githubMaxFileSize        = uint64(1 << 19)
	githubMaxResultsPerQuery = uint64(1000)
)

// Interface instead of struct for testing purposes.
// Not expecting to have multiple implementations.
type cachedSearch interface {
	CountResults(uint64, uint64) (uint64, error)
	RequestString(filesize rangeFormatter) string
}

// cachedSearch is a simple data structure that maps the upper bound (r) of a
// range from 0 to r to the number of files that have between 0 and r files
// (inclusive). It also guarantees that the counts are monotonically increasing
// (not strict) as the value for r increases, by looking at the maximal
// previous file count for the value that precedes r in the cache.
//
// It uses a bit trick to be more efficient in detecting
// inconsistencies in the returned data from the Github API.
// Therefore, the cache expects a search to always start at 0, and
// it expects the max file size to be a power of 2. If this is to be changed
// there are a few considerations to keep in mind:
//
// 1. The cache is only efficient if the queries can be reused, so if
//    the first chunk of files lives in the range 0..x, continuing the
//    search for the next chunk from x+1..max (while asymptotically sane)
//    may actually be less efficient since the cache is essentially reset
//    at every interval. This leads to a larger number of requests in
//    practice, and requests are what's expensive (rate limits).
//
// 2. The github API is not perfectly monotonic.. (this is somewhat
//    problematic). The current cache implementation looks at the
//    predecessor entry to find out if the current value is monotonic.
//    This is where the bit trick is used, since each step in the binary
//    search is adding or omitting to add a decreasing power of 2 to the query
//    value, we can remove the least significant set bit to find the
//    predecessor in constant time. Ultimately since the search is rate
//    limited, we could also easily afford to compute this in linear time
//    by iterating over cached values. So this trick is not crucial to the
//    cache's performance.
type githubCachedSearch struct {
	cache       map[uint64]uint64
	gcl         GhClient
	baseRequest request
}

func newCache(client GhClient, query Query) githubCachedSearch {
	return githubCachedSearch{
		cache: map[uint64]uint64{
			0: 0,
		},
		gcl:         client,
		baseRequest: client.CodeSearchRequestWith(query),
	}
}

func (c githubCachedSearch) CountResults(lowerBound, upperBound uint64) (uint64, error) {
	count, cached := c.cache[upperBound]
	if cached {
		return count, nil
	}

	sizeRange := RangeWithin{lowerBound, upperBound}
	rangeRequest := c.RequestString(sizeRange)

	result := c.gcl.parseGithubResponseWithRetry(rangeRequest)
	if result.Error != nil {
		return count, result.Error
	}

	// As range search uses powers of 2 for binary search, the previously
	// cached value is easy to find by removing the least significant set
	// bit from the current upperBound, since each step of the search adds
	// least significant set bit.
	//
	// Finding the predecessor could also be implemented by iterating over
	// the map to find the largest key that is smaller than upperBound if
	// this approach deemed too complex.
	trail := bits.TrailingZeros64(upperBound)
	prev := uint64(0)
	if trail != 64 {
		prev = upperBound - (1 << uint64(trail))
	}

	// Sometimes the github API is not monotonically increasing, or ouputs
	// an erroneous value of 0, or 1. This logic makes sure that it was not
	// erroneous, and that the sequence continues to be monotonic by setting
	// the current query count to match the previous value. which at least
	// guarantees that the range search terminates.
	//
	// On the other hand, if files are added, then we way loose out on some
	// files in a reviously completed range, but these files should be there
	// the next time the crawler runs, so this is not really problematic.
	retryMonotonicCount := 4
	for result.Parsed.TotalCount < c.cache[prev] {
		logger.Printf(
			"Retrying query... current lower bound: %d, got: %d\n",
			c.cache[prev], result.Parsed.TotalCount)

		result = c.gcl.parseGithubResponseWithRetry(rangeRequest)
		if result.Error != nil {
			return count, result.Error
		}

		retryMonotonicCount--
		if retryMonotonicCount <= 0 {
			result.Parsed.TotalCount = c.cache[prev]
			logger.Println(
				"Retries for monotonic check exceeded,",
				" setting value to match predecessor")
		}
	}

	count = result.Parsed.TotalCount
	logger.Printf("Caching new query %s, with count %d (incomplete_results: %v)\n",
		sizeRange.RangeString(), count, result.Parsed.IncompleteResults)
	c.cache[upperBound] = count
	return count, nil
}

func (c githubCachedSearch) RequestString(filesize rangeFormatter) string {
	return c.baseRequest.CopyWith(Filesize(filesize)).URL()
}

// Outputs a (possibly incomplete) list of ranges to query to find most search
// results as permissible by the search github search API. Github search only
// allows 1,000 results per query (paginated).
// Source: https://developer.github.com/v3/search/
//
// This leaves the possibility of having file sizes with more than 1000 results,
// This would mean that the search as it is could not find all files. If queries
// are sorted by last indexed, and retrieved on regular intervals, it should be
// sufficient to get most if not all documents.
func FindRangesForRepoSearch(cache cachedSearch, lowerBound, upperBound uint64) ([]string, error) {
	totalFiles, err := cache.CountResults(lowerBound, upperBound)
	if err != nil {
		return nil, err
	}
	logger.Println("total kustomization files: ", totalFiles)

	if githubMaxResultsPerQuery >= totalFiles {
		return []string{
			cache.RequestString(RangeWithin{lowerBound, upperBound}),
		}, nil
	}

	// Find all the ranges of file sizes such that all files are queryable
	// using the Github API. This does not compute an optimal ranges, since
	// the number of queries needed to get the information required to
	// compute an optimal range is expected to be much larger than the
	// number of queries performed this way.
	//
	// The number of ranges is k = (number of files)/1000, and finding a
	// range is logarithmic in the max file size (n = filesize). This means
	// that preprocessing takes O(k * lg n) queries to find the ranges with
	// a binary search over file sizes.
	//
	// My intuition is that this approach is competitive to a perfectly
	// optimal solution, but I didn't actually take the time to do a
	// rigorous proof. Intuitively, since files sizes are typically power
	// law distibuted the binary search will be very skewed towards the
	// smaller file ranges. This means that in practice this approach will
	// make fewer than (#files/1000)*(log(n) = 19) queries for
	// preprocessing, since it reuses a lot of the queries in the denser
	// ranges. Furthermore, because of the distribution, it should be very
	// easy to find ranges that are very close to the upper bound, up to
	// the limiting factor of having no more than 1000 files accessible per
	// range.
	filesAccessible := uint64(0)
	sizes := make([]uint64, 0)
	sizes = append(sizes, lowerBound)
	for filesAccessible < totalFiles {
		target := filesAccessible + githubMaxResultsPerQuery
		if target >= totalFiles {
			break
		}

		logger.Printf("%d accessible files, next target = %d\n",
			filesAccessible, target)

		size, err := FindFileSize(cache, target, lowerBound, upperBound)
		if err != nil {
			return nil, err
		}

		// If there are more than 1000 files in the next bucket, we must
		// advance anyway and lose out on some files :(.
		if l := len(sizes); l > 0 && sizes[l-1] == size {
			size++
		}

		nextAccessible, err := cache.CountResults(lowerBound, size)
		if err != nil {
			return nil, fmt.Errorf(
				"cache should be populated at %d already, got %v",
				size, err)
		}
		if nextAccessible < filesAccessible {
			return nil, fmt.Errorf(
				"number of results dropped from %d to %d within range search",
				filesAccessible, nextAccessible)
		}

		filesAccessible = nextAccessible
		if nextAccessible < totalFiles {
			sizes = append(sizes, size)
		}
	}
	sizes = append(sizes, upperBound)
	return formatFilesizeRanges(cache, sizes), nil
}

// FindFileSize finds the filesize range from [lowerBound, return value] that has
// the largest file count that is smaller than or equal to
// githubMaxResultsPerQuery. It is important to note that this returned value
// could already be in a previous range if the next file size has more than 1000
// results. It is left to the caller to handle this bit of logic and guarantee
// forward progession in this case.
func FindFileSize(
	cache cachedSearch, targetFileCount, lowerBound, upperBound uint64) (uint64, error) {

	// Binary search for file sizes that make up the next <=1000 element
	// chunk.
	cur := lowerBound
	increase := (upperBound - lowerBound) / 2

	for increase > 0 {
		mid := cur + increase

		count, err := cache.CountResults(lowerBound, mid)
		if err != nil {
			return count, err
		}

		if count <= targetFileCount {
			cur = mid
		}

		if count == targetFileCount {
			break
		}

		increase /= 2
	}

	return cur, nil
}

func formatFilesizeRanges(cache cachedSearch, sizes []uint64) []string {
	n := len(sizes)
	if n < 2 {
		return []string{}
	}

	ranges := make([]string, 0, n-1)
	ranges = append(ranges, cache.RequestString(RangeWithin{sizes[0], sizes[1]}))
	for i := 1; i < n-1; i++ {
		ranges = append(ranges, cache.RequestString(RangeWithin{sizes[i] + 1, sizes[i+1]}))
	}
	return ranges
}

func RangeSizes(s string) RangeWithin {
	start := strings.Index(s, "+size:") + len("+size:")
	end := strings.Index(s, "&")
	ranges := strings.Split(s[start:end], "..")
	lowerBound, _ := strconv.ParseUint(ranges[0], 10, 64)
	upperBound, _ := strconv.ParseUint(ranges[1], 10, 64)
	return RangeWithin{lowerBound, upperBound}
}
