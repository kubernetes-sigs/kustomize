// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
)

func TestParseGitRepositoryPath(t *testing.T) {
	var testCases = map[string]struct {
		urlString string
		expected  string
	}{
		"ssh format": {
			urlString: "git@github.com:path/dummy.git",
			expected:  "github.com/path/dummy",
		},
		"https format": {
			urlString: "https://github.com/path/dummy.git",
			expected:  "github.com/path/dummy",
		},
		"unknown format": {
			urlString: "file:///path/to/repo.git/",
			expected:  "",
		},
	}

	for n, tc := range testCases {
		actual := ParseGitRepositoryPath(tc.urlString)
		t.Log(actual)
		if actual != tc.expected {
			t.Errorf("%s: expected %s, got %s", n, tc.expected, actual)
		}
	}
}
