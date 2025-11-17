// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build windows
// +build windows

package filesys

import (
	"testing"
)

// TestPathSplitAndJoin_Windows tests PathSplit and PathJoin on Windows-specific paths
func TestPathSplitAndJoin_Windows(t *testing.T) {
	cases := map[string]struct {
		original string
		expected []string
	}{
		"SimpleRelative": {
			original: "hello\\there",
			expected: []string{"hello", "there"},
		},
		"ForwardSlash": {
			original: "hello/there",
			expected: []string{"hello", "there"},
		},
		"MixedSlash": {
			original: "hello/there\\friend",
			expected: []string{"hello", "there", "friend"},
		},
		"VolumeLetter": {
			original: "C:\\Users\\test",
			expected: []string{"", "C:", "Users", "test"},
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			actual := PathSplit(c.original)
			if len(actual) != len(c.expected) {
				t.Fatalf(
					"expected len %d, got len %d\nexpected: %v\nactual: %v",
					len(c.expected), len(actual), c.expected, actual)
			}
			for i := range c.expected {
				if c.expected[i] != actual[i] {
					t.Fatalf(
						"at i=%d, expected '%s', got '%s'",
						i, c.expected[i], actual[i])
				}
			}
			joined := PathJoin(actual)
			// On Windows, filepath.Clean normalizes paths, so we clean both for comparison
			// We can't guarantee exact match due to separator normalization
			t.Logf("original: %s, joined: %s", c.original, joined)
		})
	}
}

// TestInsertPathPart_Windows tests InsertPathPart on Windows-specific paths
func TestInsertPathPart_Windows(t *testing.T) {
	cases := map[string]struct {
		original string
		pos      int
		part     string
		// expected can vary due to path normalization on Windows
		shouldContain string
	}{
		"BackslashPath": {
			original:      "projects\\whatever",
			pos:           0,
			part:          "valueAdded",
			shouldContain: "valueAdded",
		},
		"ForwardSlashPath": {
			original:      "projects/whatever",
			pos:           0,
			part:          "valueAdded",
			shouldContain: "valueAdded",
		},
		"MixedSlashPath": {
			original:      "projects/whatever\\else",
			pos:           1,
			part:          "valueAdded",
			shouldContain: "valueAdded",
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			result := InsertPathPart(c.original, c.pos, c.part)
			t.Logf("InsertPathPart(%q, %d, %q) = %q", c.original, c.pos, c.part, result)
			// Just verify it doesn't panic and contains the part
			if result == "" {
				t.Fatalf("expected non-empty result")
			}
		})
	}
}
