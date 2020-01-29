package github

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

type testCachedSearch struct {
	cache map[uint64]uint64
}

func (c testCachedSearch) CountResults(lowerBound, upperBound uint64) (uint64, error) {
	log.Printf("CountResults(%05x)\n", upperBound)
	count, ok := c.cache[upperBound]
	if !ok {
		return count, fmt.Errorf("cache not set at %x", upperBound)
	}
	return count, nil
}

func (c testCachedSearch) RequestString(filesize rangeFormatter) string {
	return filesize.RangeString()
}

// TODO(damienr74) make tests easier to write.. I'm thinking I can make the test
// cache take in a list of (filesize, count) pairs and it can populate the cache
// without relying on how the implementation will create queries. This was only
// a quick and dirty test to make sure that modifications are not going to break
// the functionality.
func TestRangeSplitting(t *testing.T) {
	// Keys follow the binary search depending on whether or not the range
	// is too small/large to find close to optimal filesize ranges. This
	// test is heavily tied to the fact that the search is using powers of two
	// to make progress in the search (hence the use of hexadecimal values).
	cache := testCachedSearch{
		map[uint64]uint64{
			0x80000: 5000,
			0x40000: 5000,
			0x20000: 5000,
			0x10000: 5000,
			0x08000: 5000,
			0x04000: 5000,
			0x02000: 5000,
			0x01000: 5000,
			0x00fff: 3950,
			0x00ffe: 3950,
			0x00ffc: 3950,
			0x00ff8: 3950,
			0x00ff0: 3950,
			0x00fe0: 3950,
			0x00fc0: 3950,
			0x00f80: 3950,
			0x00f00: 3950,
			0x00e00: 3950,
			0x00c00: 3950,
			0x00800: 3950,
			0x00400: 3950,
			0x00200: 3688,
			0x00180: 3028,
			0x00100: 2999,
			0x000c0: 2448,
			0x00080: 1999,
			0x00070: 1600,
			0x0006c: 1003,
			0x0006b: 1001,
			0x0006a: 999,
			0x00068: 999,
			0x00060: 999,
			0x00040: 999,
			0x00000: 0,
		},
	}

	requests, err := FindRangesForRepoSearch(cache, 0, 524288)
	if err != nil {
		t.Errorf("Error while finding ranges: %v", err)
	}
	expected := []string{
		"0..106",       // cache.RequestString(RangeWithin{0x00, 0x6a}),
		"107..128",     // cache.RequestString(RangeWithin{0x6b, 0x80}),
		"129..256",     // cache.RequestString(RangeWithin{0x81, 0x100}),
		"257..4095",    // cache.RequestString(RangeWithin{0x101, 0xfff}),
		"4096..524288", // cache.RequestString(RangeWithin{0x1000, 0x80000}),
	}

	if !reflect.DeepEqual(requests, expected) {
		t.Errorf("Expected requests (%v) to equal (%v)", requests, expected)
	}
}

func TestRangeSizes(t *testing.T) {
	s := "https://api.github.com/search/code?q=filename:kustomization.yaml+filename:kustomization.yml" +
		"+filename:kustomization+size:2365..10000&order=desc&per_page=100&sort=indexed"
	returnedResult := RangeSizes(s)
	expectedResult := RangeWithin{uint64(2365), uint64(10000)}
	if !reflect.DeepEqual(returnedResult, expectedResult) {
		t.Errorf("RangeSizes expected (%v), got (%v)",expectedResult, returnedResult)
	}
}
