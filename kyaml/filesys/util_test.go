// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows
// +build !windows

package filesys

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Confirm behavior of filepath.Match
func TestFilePathMatch(t *testing.T) {
	cases := []struct {
		pattern  string
		path     string
		expected bool
	}{
		{
			pattern:  "*e*",
			path:     "hey",
			expected: true,
		},
		{
			pattern:  "*e*",
			path:     "hay",
			expected: false,
		},
		{
			pattern:  "*e*",
			path:     filepath.Join("h", "e", "y"),
			expected: false,
		},
		{
			pattern:  "*/e/*",
			path:     filepath.Join("h", "e", "y"),
			expected: true,
		},
		{
			pattern:  "h/e/*",
			path:     filepath.Join("h", "e", "y"),
			expected: true,
		},
		{
			pattern:  "*/e/y",
			path:     filepath.Join("h", "e", "y"),
			expected: true,
		},
		{
			pattern:  "*/*/*",
			path:     filepath.Join("h", "e", "y"),
			expected: true,
		},
		{
			pattern:  "*/*/*",
			path:     filepath.Join("h", "e", "y", "there"),
			expected: false,
		},
		{
			pattern:  "*/*/*/t*e",
			path:     filepath.Join("h", "e", "y", "there"),
			expected: true,
		},
	}
	for _, item := range cases {
		match, err := filepath.Match(item.pattern, item.path)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if match != item.expected {
			t.Fatalf("'%s' '%s' %v\n", item.pattern, item.path, match)
		}
	}
}

// Confirm behavior of filepath.Split
func TestFilePathSplit(t *testing.T) {
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
			full: "/",
			dir:  "/",
			file: "",
		},
		{
			full: "/beans",
			dir:  "/",
			file: "beans",
		},
		{
			full: "/home/foo/bar",
			dir:  "/home/foo/",
			file: "bar",
		},
		{
			full: "/usr/local/",
			dir:  "/usr/local/",
			file: "",
		},
		{
			full: "/usr//local//go",
			dir:  "/usr//local//",
			file: "go",
		},
	}
	for _, p := range cases {
		dir, file := filepath.Split(p.full)
		if dir != p.dir || file != p.file {
			t.Fatalf(
				"in '%s',\ngot dir='%s' (expected '%s'),\n got file='%s' (expected %s).",
				p.full, dir, p.dir, file, p.file)
		}
	}
}

func TestPathSplitAndJoin(t *testing.T) {
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
			original: "hello/there",
			expected: []string{"hello", "there"},
		},
		"Three": {
			original: "hello/my/friend",
			expected: []string{"hello", "my", "friend"},
		},
	}
	for n, c := range cases {
		f := func(t *testing.T, original string, expected []string) {
			actual := PathSplit(original)
			if len(actual) != len(expected) {
				t.Fatalf(
					"expected len %d, got len %d",
					len(expected), len(actual))
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

func TestInsertPathPart(t *testing.T) {
	cases := map[string]struct {
		original string
		pos      int
		part     string
		expected string
	}{
		"rootOne": {
			original: "/",
			pos:      0,
			part:     "___",
			expected: "/___",
		},
		"rootTwo": {
			original: "/",
			pos:      444,
			part:     "___",
			expected: "/___",
		},
		"rootedFirst": {
			original: "/apple",
			pos:      0,
			part:     "___",
			expected: "/___/apple",
		},
		"rootedSecond": {
			original: "/apple",
			pos:      444,
			part:     "___",
			expected: "/apple/___",
		},
		"rootedThird": {
			original: "/apple/banana",
			pos:      444,
			part:     "___",
			expected: "/apple/banana/___",
		},
		"emptyLow": {
			original: "",
			pos:      -3,
			part:     "___",
			expected: "___",
		},
		"emptyHigh": {
			original: "",
			pos:      444,
			part:     "___",
			expected: "___",
		},
		"peachPie": {
			original: "a/nice/warm/pie",
			pos:      3,
			part:     "PEACH",
			expected: "a/nice/warm/PEACH/pie",
		},
		"rootedPeachPie": {
			original: "/a/nice/warm/pie",
			pos:      3,
			part:     "PEACH",
			expected: "/a/nice/warm/PEACH/pie",
		},
		"longStart": {
			original: "a/b/c/d/e/f",
			pos:      0,
			part:     "___",
			expected: "___/a/b/c/d/e/f",
		},
		"rootedLongStart": {
			original: "/a/b/c/d/e/f",
			pos:      0,
			part:     "___",
			expected: "/___/a/b/c/d/e/f",
		},
		"longMiddle": {
			original: "a/b/c/d/e/f",
			pos:      3,
			part:     "___",
			expected: "a/b/c/___/d/e/f",
		},
		"rootedLongMiddle": {
			original: "/a/b/c/d/e/f",
			pos:      3,
			part:     "___",
			expected: "/a/b/c/___/d/e/f",
		},
		"longEnd": {
			original: "a/b/c/d/e/f",
			pos:      444,
			part:     "___",
			expected: "a/b/c/d/e/f/___",
		},
		"rootedLongEnd": {
			original: "/a/b/c/d/e/f",
			pos:      444,
			part:     "___",
			expected: "/a/b/c/d/e/f/___",
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			actual := InsertPathPart(c.original, c.pos, c.part)
			if actual != c.expected {
				t.Fatalf("expected '%s', got '%s'", c.expected, actual)
			}
		})
	}
}

func TestStripTrailingSeps(t *testing.T) {
	cases := []struct {
		full string
		rem  string
	}{
		{
			full: "foo",
			rem:  "foo",
		},
		{
			full: "",
			rem:  "",
		},
		{
			full: "foo/",
			rem:  "foo",
		},
		{
			full: "foo///bar///",
			rem:  "foo///bar",
		},
		{
			full: "/////",
			rem:  "",
		},
		{
			full: "/",
			rem:  "",
		},
	}
	for _, p := range cases {
		dir := StripTrailingSeps(p.full)
		if dir != p.rem {
			t.Fatalf(
				"in '%s', got dir='%s' (expected '%s')",
				p.full, dir, p.rem)
		}
	}
}

func TestStripLeadingSeps(t *testing.T) {
	cases := []struct {
		full string
		rem  string
	}{
		{
			full: "foo",
			rem:  "foo",
		},
		{
			full: "",
			rem:  "",
		},
		{
			full: "/foo",
			rem:  "foo",
		},
		{
			full: "///foo///bar///",
			rem:  "foo///bar///",
		},
		{
			full: "/////",
			rem:  "",
		},
		{
			full: "/",
			rem:  "",
		},
	}
	for _, p := range cases {
		dir := StripLeadingSeps(p.full)
		if dir != p.rem {
			t.Fatalf(
				"in '%s', got dir='%s' (expected '%s')",
				p.full, dir, p.rem)
		}
	}
}

func TestIsHiddenFilePath(t *testing.T) {
	tests := map[string]struct {
		paths        []string
		expectHidden bool
	}{
		"hiddenGlobs": {
			expectHidden: true,
			paths: []string{
				".*",
				"/.*",
				"dir/.*",
				"dir1/dir2/dir3/.*",
				"../../.*",
				"../../dir/.*",
			},
		},
		"visibleGlobes": {
			expectHidden: false,
			paths: []string{
				"*",
				"/*",
				"dir/*",
				"dir1/dir2/dir3/*",
				"../../*",
				"../../dir/*",
			},
		},
		"hiddenFiles": {
			expectHidden: true,
			paths: []string{
				".root_file.xtn",
				"/.file_1.xtn",
				"dir/.file_2.xtn",
				"dir1/dir2/dir3/.file_3.xtn",
				"../../.file_4.xtn",
				"../../dir/.file_5.xtn",
			},
		},
		"visibleFiles": {
			expectHidden: false,
			paths: []string{
				"root_file.xtn",
				"/file_1.xtn",
				"dir/file_2.xtn",
				"dir1/dir2/dir3/file_3.xtn",
				"../../file_4.xtn",
				"../../dir/file_5.xtn",
			},
		},
	}
	for n, c := range tests {
		t.Run(n, func(t *testing.T) {
			for _, path := range c.paths {
				actual := IsHiddenFilePath(path)
				if actual != c.expectHidden {
					t.Fatalf("For file path %q, expected hidden: %v, got hidden: %v", path, c.expectHidden, actual)
				}
			}
		})
	}
}

func TestRemoveHiddenFiles(t *testing.T) {
	paths := []string{
		"file1.xtn",
		".file2.xtn",
		"dir/fa1",
		"dir/fa2",
		"dir/.fa3",
		"../../.fa4",
		"../../fa5",
		"../../dir/fa6",
		"../../dir/.fa7",
	}
	result := RemoveHiddenFiles(paths)
	expected := []string{
		"file1.xtn",
		"dir/fa1",
		"dir/fa2",
		"../../fa5",
		"../../dir/fa6",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Hidden dirs not correctly removed, expected %v but got %v\n", expected, result)
	}
}
