// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package edit

import (
	"testing"
)

func TestUpstairs(t *testing.T) {
	var testCases = map[string]struct {
		depth    int
		expected string
	}{
		"zero": {
			depth:    0,
			expected: "",
		},
		"one": {
			depth:    1,
			expected: "../",
		},
		"five": {
			depth:    5,
			expected: "../../../../../",
		},
	}
	for n, tc := range testCases {
		if tc.expected != upstairs(tc.depth) {
			t.Fatalf(
				"%s: for depth %d, expected %q, got %q",
				n, tc.depth, tc.expected, upstairs(tc.depth))
		}
	}
}
