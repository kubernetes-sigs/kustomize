// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filtersutil

import (
	"sort"
)

// SortedMapKeys returns a sorted slice of keys to the given map.
// Writing this function never gets old.
func SortedMapKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
