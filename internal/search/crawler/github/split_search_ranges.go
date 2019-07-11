package github

import (
	"fmt"
	"math/bits"
)

// Files cannot be more than 2^19 bytes, according to
// https://help.github.com/en/articles/searching-code#considerations-for-code-search
const (
	githubMaxFileSize        = uint64(1 << 19)
	githubMaxResultsPerQuery = uint64(1000)
)

// Interface for testing purposes. Not expecting to have multiple
// implementations.
type cachedSearch interface {
	CountResults(uint64) (uint64, error)
	RequestString(filesize rangeFormatter) string
}

// Cache uses bit tricks to be more efficient in detecting
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
//    search is adding or ommiting to add a decreasing of 2 to the query value,
//    we can remove the least significant set bit to find the predecessor in
//    constant time. Ultimately since the search is rate limited, we could also
//    easily afford to compute this in linear time by iterating
//    over cached values.
type githubCachedSearch struct {
	cache       map[uint64]uint64
	retryCount  uint64
	baseRequest request
}

func newCache(rc RequestConfig, query Query) githubCachedSearch {
	return githubCachedSearch{
		cache: map[uint64]uint64{
			0: 0,
		},
		retryCount:  rc.RetryCount(),
		baseRequest: rc.CodeSearchRequestWith(query),
	}
}

func (c githubCachedSearch) CountResults(upperBound uint64) (uint64, error) {
	count, cached := c.cache[upperBound]
	if cached {
		return count, nil
	}

	sizeRange := RangeWithin{0, upperBound}
	rangeRequest := c.RequestString(sizeRange)

	result := parseGithubResponse(rangeRequest, c.retryCount)
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

		result = parseGithubResponse(rangeRequest, c.retryCount)
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
	logger.Printf("Caching new query %s, with count %d\n",
		sizeRange.RangeString(), count)
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
func FindRangesForRepoSearch(cache cachedSearch) ([]string, error) {
	totalFiles, err := cache.CountResults(githubMaxFileSize)
	if err != nil {
		return nil, err
	}
	logger.Println("total files: ", totalFiles)

	if githubMaxResultsPerQuery >= totalFiles {
		return []string{
			cache.RequestString(RangeWithin{0, githubMaxFileSize}),
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
	// rigurous proof. Intuitively, since files sizes are typically power
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
	for filesAccessible < totalFiles {
		target := filesAccessible + githubMaxResultsPerQuery
		if target >= totalFiles {
			break
		}

		logger.Printf("%d accessible files, next target = %d\n",
			filesAccessible, target)

		cur, err := lowerBoundFileCount(cache, target)
		if err != nil {
			return nil, err
		}

		// If there are more than 1000 files in the next bucket, we must
		// advance anyway and lose out on some files :(.
		if l := len(sizes); l > 0 && sizes[l-1] == cur {
			cur++
		}

		nextAccessible, err := cache.CountResults(cur)
		if err != nil {
			return nil, fmt.Errorf(
				"cache should be populated at %d already, got %v",
				cur, err)
		}
		if nextAccessible < filesAccessible {
			return nil, fmt.Errorf(
				"Number of results dropped from %d to %d within range search",
				filesAccessible, nextAccessible)
		}

		filesAccessible = nextAccessible
		if nextAccessible < totalFiles {
			sizes = append(sizes, cur)
		}
	}

	return formatFilesizeRanges(cache, sizes), nil
}

// lowerBoundFileCount finds the filesize range from [0, return value] that has
// the largest file count that is smaller than or equal to
// githubMaxResultsPerQuery. It is important to note that this returned value
// could already be in a previous range if the next file size has more than 1000
// results. It is left to the caller to handle this bit of logic and guarantee
// forward progession in this case.
func lowerBoundFileCount(
	cache cachedSearch, targetFileCount uint64) (uint64, error) {

	// Binary search for file sizes that make up the next <=1000 element
	// chunk.
	cur := uint64(0)
	increase := githubMaxFileSize / 2

	for increase > 0 {
		mid := cur + increase

		count, err := cache.CountResults(mid)
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
	ranges := make([]string, 0, len(sizes)+1)

	if len(sizes) > 0 {
		ranges = append(ranges, cache.RequestString(
			RangeLessThan{sizes[0] + 1},
		))
	}

	for i := 0; i < len(sizes)-1; i += 1 {
		ranges = append(ranges, cache.RequestString(
			RangeWithin{sizes[i] + 1, sizes[i+1]},
		))

		if i != len(sizes)-2 {
			continue
		}
		ranges = append(ranges, cache.RequestString(
			RangeGreaterThan{sizes[i+1]},
		))
	}

	return ranges
}
