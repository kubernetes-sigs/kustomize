// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package filesys

import (
	"os"
	"path/filepath"
	"testing"
)

// Confirm behavior of filepath.Split on Windows.
// On Windows, filepath.Split treats volume names (e.g. "C:") specially:
// filepath.Split("C:") returns dir="C:", path="" because "C:" is a volume name
// with no trailing separator. This is relevant to the PathSplit recursion bug.
func TestFilePathSplitWindows(t *testing.T) {
	cases := []struct {
		full string
		dir  string
		file string
	}{
		{
			full: "",
			dir:  "",
			file: "",
		},
		{
			full: SelfDir,
			dir:  "",
			file: SelfDir,
		},
		{
			full: "rabbit.jpg",
			dir:  "",
			file: "rabbit.jpg",
		},
		{
			full: `\`,
			dir:  `\`,
			file: "",
		},
		{
			full: `\beans`,
			dir:  `\`,
			file: "beans",
		},
		{
			full: `C:\`,
			dir:  `C:\`,
			file: "",
		},
		{
			full: `C:`,
			dir:  `C:`,
			file: "",
		},
		{
			full: `C:\Users`,
			dir:  `C:\`,
			file: "Users",
		},
		{
			full: `C:\Users\foo\bar`,
			dir:  `C:\Users\foo\`,
			file: "bar",
		},
		{
			full: `C:\Users\foo\`,
			dir:  `C:\Users\foo\`,
			file: "",
		},
	}
	for _, p := range cases {
		dir, file := filepath.Split(p.full)
		if dir != p.dir || file != p.file {
			t.Fatalf(
				"in '%s',\ngot dir='%s' (expected '%s'),\n got file='%s' (expected '%s').",
				p.full, dir, p.dir, file, p.file)
		}
	}
}

// TestPathSplitAndJoinWindows tests PathSplit and PathJoin with
// relative and backslash-absolute paths on Windows.
func TestPathSplitAndJoinWindows(t *testing.T) {
	cases := map[string]struct {
		original string
		expected []string
	}{
		"Empty": {
			original: "",
			expected: []string{},
		},
		"One": {
			original: "hello",
			expected: []string{"hello"},
		},
		"Two": {
			original: `hello\there`,
			expected: []string{"hello", "there"},
		},
		"Three": {
			original: `hello\my\friend`,
			expected: []string{"hello", "my", "friend"},
		},
	}
	for n, c := range cases {
		f := func(t *testing.T, original string, expected []string) {
			t.Helper()
			actual := PathSplit(original)
			if len(actual) != len(expected) {
				t.Fatalf(
					"expected len %d, got len %d (%v)",
					len(expected), len(actual), actual)
			}
			for i := range expected {
				if expected[i] != actual[i] {
					t.Fatalf(
						"at i=%d, expected '%s', got '%s'",
						i, expected[i], actual[i])
				}
			}
			joined := PathJoin(actual)
			if joined != original {
				t.Fatalf(
					"when rejoining, expected '%s', got '%s'",
					original, joined)
			}
		}
		t.Run("relative"+n, func(t *testing.T) {
			f(t, c.original, c.expected)
		})
		t.Run("absolute"+n, func(t *testing.T) {
			f(t,
				string(os.PathSeparator)+c.original,
				append([]string{""}, c.expected...))
		})
	}
}

// TestPathSplitWindowsVolumePaths tests PathSplit with Windows
// volume paths (drive letters, UNC). Without the fix, these paths
// cause infinite recursion (stack overflow) because filepath.Split
// returns the volume name as dir with no trailing separator to trim,
// creating a fixed-point in the recursion.
func TestPathSplitWindowsVolumePaths(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected []string
	}{
		"DriveLetterOnly": {
			input:    `C:`,
			expected: []string{"C:"},
		},
		"DriveRoot": {
			input:    `C:\`,
			expected: []string{"C:", ""},
		},
		"DriveRootOneDir": {
			input:    `C:\Users`,
			expected: []string{"C:", "Users"},
		},
		"DriveRootTwoDirs": {
			input:    `C:\Users\foo`,
			expected: []string{"C:", "Users", "foo"},
		},
		"DriveRootThreeDirs": {
			input:    `C:\Users\foo\bar`,
			expected: []string{"C:", "Users", "foo", "bar"},
		},
		"DriveRootTrailingBackslash": {
			input:    `C:\foo\`,
			expected: []string{"C:", "foo", ""},
		},
		"RelativeDrivePath": {
			input:    `C:foo`,
			expected: []string{"C:", "foo"},
		},
		"UNCShareOnly": {
			input:    `\\server\share`,
			expected: []string{`\\server\share`},
		},
		"UNCShareWithDir": {
			input:    `\\server\share\foo`,
			expected: []string{`\\server\share`, "foo"},
		},
		"UNCShareWithNestedDirs": {
			input:    `\\server\share\foo\bar`,
			expected: []string{`\\server\share`, "foo", "bar"},
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			actual := PathSplit(c.input)
			if len(actual) != len(c.expected) {
				t.Fatalf(
					"expected len %d, got len %d (%v)",
					len(c.expected), len(actual), actual)
			}
			for i := range c.expected {
				if c.expected[i] != actual[i] {
					t.Fatalf(
						"at i=%d, expected '%s', got '%s'",
						i, c.expected[i], actual[i])
				}
			}
		})
	}
}
