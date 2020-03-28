package filtersutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
)

func TestSortedKeys(t *testing.T) {
	testCases := map[string]struct {
		input    map[string]string
		expected []string
	}{
		"empty": {
			input:    map[string]string{},
			expected: []string{}},
		"one": {
			input:    map[string]string{"a": "aaa"},
			expected: []string{"a"}},
		"three": {
			input:    map[string]string{"c": "ccc", "b": "bbb", "a": "aaa"},
			expected: []string{"a", "b", "c"}},
	}
	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			if !assert.Equal(t,
				filtersutil.SortedMapKeys(tc.input),
				tc.expected) {
				t.FailNow()
			}
		})
	}
}
