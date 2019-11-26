// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const content = `
Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
sed do eiusmod tempor incididunt
ut labore et dolore magna aliqua.
`
const shortContent = "hi"

var topCases = []pathCase{
	{
		what:   "dotdot",
		arg:    ParentDir,
		errStr: "illegal name '..' in file creation",
	},
	{
		what:   "empty",
		arg:    "",
		name:   "",
		errStr: "illegal name '.' in file creation",
	},
	{
		what: "simple",
		arg:  "bob",
		name: "bob",
		path: "bob",
	},
	{
		what: "longer",
		arg:  filepath.Join("longer", "bob"),
		name: "bob",
		path: filepath.Join("longer", "bob"),
	},
	{
		what: "longer yet",
		arg:  filepath.Join("longer", "foo", "bar", "beans", "bob"),
		name: "bob",
		path: filepath.Join("longer", "foo", "bar", "beans", "bob"),
	},
	{
		what: "tricky",
		arg:  filepath.Join("bob", ParentDir, "sally"),
		name: "sally",
		path: "sally",
	},
	{
		what: "trickier",
		arg:  filepath.Join("bob", "sally", ParentDir, ParentDir, "jean"),
		name: "jean",
		path: "jean",
	},
}

func TestMakeEmptyDirInMemory(t *testing.T) {
	n := MakeEmptyDirInMemory()
	if !n.isNodeADir() {
		t.Fatalf("not a directory")
	}
	if n.Size() != 0 {
		t.Fatalf("unexpected size %d", n.Size())
	}
	if n.Name() != "" {
		t.Fatalf("unexpected name '%s'", n.Name())
	}
	if n.Path() != "" {
		t.Fatalf("unexpected path '%s'", n.Path())
	}
	runBasicOperations(
		t, "MakeEmptyDirInMemory", false, topCases, n)
}

func TestMakeFsInMemory(t *testing.T) {
	runBasicOperations(
		t, "MakeFsInMemory", true, topCases, MakeFsInMemory())
}

//nolint:gocyclo
func runBasicOperations(
	t *testing.T, tName string, isFSysRooted bool,
	cases []pathCase, fSys FileSystem) {
	buff := make([]byte, 500)
	for _, c := range cases {
		err := fSys.WriteFile(c.arg, []byte(content))
		if c.errStr != "" {
			if err == nil {
				t.Fatalf("%s; expected error writing to  '%s'!", c.what, c.arg)
			}
			if !strings.Contains(err.Error(), c.errStr) {
				t.Fatalf("%s; expected err containing '%s', got '%v'",
					c.what, c.errStr, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		if !fSys.Exists(c.path) {
			t.Fatalf("%s; expect existence of '%s'", c.what, c.path)
		}
		stuff, err := fSys.ReadFile(c.path)
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		if string(stuff) != content {
			t.Fatalf("%s; unexpected content '%s'", c.what, stuff)
		}
		f, err := fSys.Open(c.arg)
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		fi, err := f.Stat()
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		if fi.Name() != c.name {
			t.Fatalf("%s; expected name '%s', got '%s'", c.what, c.name, fi.Name())
		}
		count, err := f.Read(buff)
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		if string(buff[:count]) != content {
			t.Fatalf("%s; unexpected buff '%s'", c.what, buff)
		}
		count, err = f.Write([]byte(shortContent))
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		if count != len(shortContent) {
			t.Fatalf("%s; unexpected count: %d", c.what, len(shortContent))
		}
		stuff, err = fSys.ReadFile(c.path)
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		if string(stuff) != shortContent {
			t.Fatalf("%s; unexpected content '%s'", c.what, stuff)
		}
	}

	var actualPaths []string
	var err error
	prefix := ""
	{
		root := SelfDir
		if isFSysRooted {
			root = Separator
			prefix = Separator
		}
		err = fSys.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("err '%v' at path %q\n", err, path)
				return nil
			}
			if !info.IsDir() {
				actualPaths = append(actualPaths, path)
			}
			return nil
		})
	}
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	var expectedPaths []string
	for _, c := range cases {
		if c.errStr == "" {
			expectedPaths = append(expectedPaths, prefix+c.path)
		}
	}
	sort.Strings(expectedPaths)
	assertEqualStringSlices(t, expectedPaths, actualPaths, tName)
}

type pathCase struct {
	what   string
	arg    string
	name   string
	path   string
	errStr string
}

func TestAddDir(t *testing.T) {
	cases := []pathCase{
		{
			what:   "dotdot",
			arg:    ParentDir,
			errStr: "cannot add a directory above ''",
		},
		{
			what: "empty",
			arg:  "",
			name: "",
			path: "",
		},
		{
			what: "simple",
			arg:  "bob",
			name: "bob",
			path: "bob",
		},
		{
			what: "longer",
			arg:  filepath.Join("longer", "bob"),
			name: "bob",
			path: filepath.Join("longer", "bob"),
		},
		{
			what: "longer yet",
			arg:  filepath.Join("longer", "foo", "bar", "beans", "bob"),
			name: "bob",
			path: filepath.Join("longer", "foo", "bar", "beans", "bob"),
		},
		{
			what: "tricky",
			arg:  filepath.Join("bob", ParentDir, "sally"),
			name: "sally",
			path: "sally",
		},
		{
			what: "trickier",
			arg:  filepath.Join("bob", "sally", ParentDir, ParentDir, "jean"),
			name: "jean",
			path: "jean",
		},
	}
	for _, c := range cases {
		n := MakeEmptyDirInMemory()
		f, err := n.AddDir(c.arg)
		if c.errStr != "" {
			if err == nil {
				t.Fatalf("%s; expected error!", c.what)
			}
			if !strings.Contains(err.Error(), c.errStr) {
				t.Fatalf(
					"%s; expected error with '%s', got '%v'",
					c.what, c.errStr, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", c.what, err)
		}
		checkNode(t, c.what, f, c.name, 0, true, c.path)
		checkOsStat(t, c.what, f, f.Name(), 0, true)
	}
}

var bagOfCases = []pathCase{
	{
		what:   "empty",
		arg:    "",
		errStr: "illegal name '.' in file creation",
	},
	{
		what: "simple",
		arg:  "bob",
		name: "bob",
		path: "bob",
	},
	{
		what: "longer",
		arg:  filepath.Join("longer", "bob"),
		name: "bob",
		path: filepath.Join("longer", "bob"),
	},
	{
		what: "longer",
		arg:  filepath.Join("longer", "sally"),
		name: "sally",
		path: filepath.Join("longer", "sally"),
	},
	{
		what: "even longer",
		arg:  filepath.Join("longer", "than", "the", "other", "bob"),
		name: "bob",
		path: filepath.Join("longer", "than", "the", "other", "bob"),
	},
	{
		what: "even longer",
		arg:  filepath.Join("even", "much", "longer", "than", "the", "other", "bob"),
		name: "bob",
		path: filepath.Join("even", "much", "longer", "than", "the", "other", "bob"),
	},
}

func TestAddFile(t *testing.T) {
	n := MakeEmptyDirInMemory()
	if n.FileCount() != 0 {
		t.Fatalf("expected no files, got %d", n.FileCount())
	}
	expectedFileCount := 0
	for _, c := range bagOfCases {
		f, err := n.AddFile(c.arg, []byte(content))
		if c.errStr != "" {
			if err == nil {
				t.Fatalf("%s; expected error!", c.what)
			}
			if !strings.Contains(err.Error(), c.errStr) {
				t.Fatalf("%s; expected err containing '%s', got '%v'",
					c.what, c.errStr, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s; unexpected error %v", c.what, err)
		}
		checkNode(t, c.what, f, c.name, len(content), false, c.path)
		checkOsStat(t, c.what, f, f.Name(), len(content), false)

		result, err := n.Find(c.arg)
		if err != nil {
			t.Fatalf("%s; unexpected find error %v", c.what, err)
		}
		if result != f {
			t.Fatalf("%s; unexpected find result %v", c.what, result)
		}

		result, err = n.Find(filepath.Join("longer", "bogus"))
		if err != nil {
			t.Fatalf("%s; unexpected find error %v", c.what, err)
		}
		if result != nil {
			t.Fatalf("%s; unexpected find result %v", c.what, result)
		}

		expectedFileCount++
		fc := n.FileCount()
		if fc != expectedFileCount {
			t.Fatalf("expected file count %d, got %d",
				expectedFileCount, fc)
		}
	}
}

func checkNode(
	t *testing.T, what string, f *fsNode, name string,
	size int, isDir bool, path string) {
	if f.isNodeADir() != isDir {
		t.Fatalf("%s; unexpected isNodeADir = %v", what, f.isNodeADir())
	}
	if f.Size() != int64(size) {
		t.Fatalf("%s; unexpected size %d", what, f.Size())
	}
	if name != f.Name() {
		t.Fatalf("%s; expected name '%s', got '%s'", what, name, f.Name())
	}
	if path != f.Path() {
		t.Fatalf("%s; expected path '%s', got '%s'", what, path, f.Path())
	}
}

func checkOsStat(
	t *testing.T, what string, f File, name string,
	size int, isDir bool) {
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("%s; unexpected stat error %v", what, err)
	}
	if info.IsDir() != isDir {
		t.Fatalf("%s; unexpected info.isNodeADir = %v", what, info.IsDir())
	}
	if info.Size() != int64(size) {
		t.Fatalf("%s; unexpected info.size %d", what, info.Size())
	}
	if info.Name() != name {
		t.Fatalf("%s; expected name '%s', got info.Name '%s'", what, name, info.Name())
	}
}

var bunchOfFiles = []struct {
	path     string
	addAsDir bool
}{
	{
		path: filepath.Join("b", "e", "a", "c", "g"),
	},
	{
		path: filepath.Join("z", "r", "a", "b", "g"),
	},
	{
		path: filepath.Join("b", "q", "a", "c", "g"),
	},
	{
		path:     filepath.Join("b", "a", "a", "m", "g"),
		addAsDir: true,
	},
	{
		path: filepath.Join("b", "w"),
	},
	{
		path: filepath.Join("b", "d", "a", "c", "m"),
	},
	{
		path: filepath.Join("b", "d", "z"),
	},
	{
		path: filepath.Join("b", "d", "y"),
	},
	{
		path: filepath.Join("b", "d", "ignore", "c", "n"),
	},
	{
		path: filepath.Join("b", "d", "x"),
	},
	{
		path: filepath.Join("b", "d", "ignore", "c", "o"),
	},
	{
		path: filepath.Join("b", "d", "ignore", "c", "m"),
	},
	{
		path:     filepath.Join("b", "d", "a", "c", "i"),
		addAsDir: true,
	},
	{
		path: filepath.Join("x"),
	},
	{
		path: filepath.Join("y"),
	},
	{
		path: filepath.Join("b", "d", "a", "c", "i", "beans"),
	},
	{
		path:     filepath.Join("b", "d", "a", "c", "r", "w"),
		addAsDir: true,
	},
	{
		path: filepath.Join("b", "d", "a", "c", "u"),
	},
}

func makeLoadedFileTree(t *testing.T) *fsNode {
	n := MakeEmptyDirInMemory()
	var err error
	expectedFileCount := 0
	for _, item := range bunchOfFiles {
		if item.addAsDir {
			_, err = n.AddDir(item.path)
		} else {
			_, err = n.AddFile(item.path, []byte(content))
			expectedFileCount++
		}
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
	}
	fc := n.FileCount()
	if fc != expectedFileCount {
		t.Fatalf("expected file count %d, got %d",
			expectedFileCount, fc)
	}
	return n
}

func TestWalkMe(t *testing.T) {
	n := makeLoadedFileTree(t)
	var actualPaths []string
	err := n.WalkMe(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("err '%v' at path %q\n", err, path)
			return nil
		}
		if info.IsDir() {
			if info.Name() == "ignore" {
				return filepath.SkipDir
			}
		} else {
			actualPaths = append(actualPaths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	var expectedPaths []string
	for _, c := range bunchOfFiles {
		if !c.addAsDir && !strings.Contains(c.path, "ignore") {
			expectedPaths = append(expectedPaths, c.path)
		}
	}
	sort.Strings(expectedPaths)
	assertEqualStringSlices(t, expectedPaths, actualPaths, "testWalkMe")
}

func TestRemove(t *testing.T) {
	n := makeLoadedFileTree(t)
	orgCount := n.FileCount()

	// Remove the "ignore" directory and everything below it.
	path := filepath.Join("b", "d", "ignore")
	result, err := n.Find(path)
	if err != nil {
		t.Fatalf("%s; unexpected error %v", path, err)
	}
	if result == nil {
		t.Fatalf("%s; expected to find '%s'", path, path)
	}
	if !result.isNodeADir() {
		t.Fatalf("%s; expected to find a directory", path)
	}
	err = result.Remove()
	if err != nil {
		t.Fatalf("%s; unable to remove: %v", path, err)
	}
	result, err = n.Find(path)
	if err != nil {
		// Just because it's gone doesn't mean error.
		t.Fatalf("%s; unexpected error %v", path, err)
	}
	if result != nil {
		t.Fatalf("%s; should not have been able to find '%s'", path, path)
	}

	// There were three files below "ignore".
	orgCount -= 3

	// Now drop one more for a total of four dropped.
	result, _ = n.Find(filepath.Join("y"))
	err = result.Remove()
	if err != nil {
		t.Fatalf("%s; unable to remove: %v", path, err)
	}
	orgCount -= 1

	fc := n.FileCount()
	if fc != orgCount {
		t.Fatalf("expected file count %d, got %d",
			orgCount, fc)
	}
}

func TestExists(t *testing.T) {
	n := makeLoadedFileTree(t)
	path := filepath.Join("b", "d", "a")
	if !n.Exists(path) {
		t.Fatalf("expected existence at %s", path)
	}
	if !n.IsDir(path) {
		t.Fatalf("expected directory at %s", path)
	}
}

func TestRegExpGlob(t *testing.T) {
	n := makeLoadedFileTree(t)
	expected := []string{
		filepath.Join("b", "d", "a", "c", "i", "beans"),
		filepath.Join("b", "d", "a", "c", "m"),
		filepath.Join("b", "d", "a", "c", "u"),
		filepath.Join("b", "d", "ignore", "c", "m"),
		filepath.Join("b", "d", "ignore", "c", "n"),
		filepath.Join("b", "d", "ignore", "c", "o"),
		filepath.Join("b", "d", "x"),
		filepath.Join("b", "d", "y"),
		filepath.Join("b", "d", "z"),
	}
	paths, err := n.RegExpGlob("b/d/*")
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	assertEqualStringSlices(t, expected, paths, "glob test")
}

func TestGlob(t *testing.T) {
	n := makeLoadedFileTree(t)
	expected := []string{
		filepath.Join("b", "d", "x"),
		filepath.Join("b", "d", "y"),
		filepath.Join("b", "d", "z"),
	}
	paths, err := n.Glob("b/d/*")
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	assertEqualStringSlices(t, expected, paths, "glob test")
}

func assertEqualStringSlices(t *testing.T, expected, actual []string, message string) {
	if len(expected) != len(actual) {
		t.Fatalf(
			"%s; unequal sizes; len(expected)=%d, len(actual)=%d\n%+v\n%+v\n",
			message, len(expected), len(actual), expected, actual)
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf(
				"%s; unequal entries; expected=%s, actual=%s",
				message, expected[i], actual[i])
		}
	}
}

func TestFind(t *testing.T) {
	cases := []struct {
		what       string
		arg        string
		expectDir  bool
		expectFile bool
		errStr     string
	}{
		{
			what: "garbage",
			arg:  "///1(*&SA",
		},
		{
			what: "simple",
			arg:  "bob",
		},
		{
			what: "no directory",
			arg:  filepath.Join("b", "rrrrrr"),
		},
		{
			what:      "is a directory",
			arg:       filepath.Join("b", "d", "ignore"),
			expectDir: true,
		},
		{
			what:       "longer, ending in file",
			arg:        filepath.Join("b", "d", "x"),
			expectFile: true,
		},
		{
			what:       "moar longer, ending in file",
			arg:        filepath.Join("b", "d", "a", "c", "u"),
			expectFile: true,
		},
		{
			what:      "directory",
			arg:       filepath.Join("b"),
			expectDir: true,
		},
		{
			// Querying for the empty string could
			// 1) be an error,
			// 2) return no result (and no error) as with
			//    any illegal and therefore non-existent
			//    file name,
			// 3) return the node itself, like running
			//    'ls' with no argument.
			// Going with option 2 (no result, no error),
			// since at this low level it makes more sense
			// if the results for the empty string query
			// differ from the results for the "." query.
			what: "empty name",
			arg:  "",
		},
		{
			what:      "self dir",
			arg:       SelfDir,
			expectDir: true,
		},
		{
			what: "parent dir - doesn't exist",
			arg:  ParentDir,
		},
		{
			what: "many parents - doesn't exist",
			arg:  filepath.Join(ParentDir, ParentDir, ParentDir),
		},
	}

	n := makeLoadedFileTree(t)
	for _, item := range cases {
		result, err := n.Find(item.arg)
		if item.errStr != "" {
			if err == nil {
				t.Fatalf("%s; expected error", item.what)
			}
			if !strings.Contains(err.Error(), item.errStr) {
				t.Fatalf("%s; expected err containing '%s', got '%v'",
					item.what, item.errStr, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", item.what, err)
		}
		if result == nil {
			if item.expectDir {
				t.Fatalf(
					"%s; expected to find directory '%s'", item.what, item.arg)
			}
			if item.expectFile {
				t.Fatalf(
					"%s; expected to find file '%s'", item.what, item.arg)
			}
			continue
		}
		if item.expectDir {
			if !result.isNodeADir() {
				t.Fatalf(
					"%s; expected '%s' to be a directory", item.what, item.arg)
			}
			continue
		}
		if item.expectFile {
			if result.isNodeADir() {
				t.Fatalf("%s; expected '%s' to be a file", item.what, item.arg)
			}
			continue
		}
		t.Fatalf(
			"%s; expected nothing for '%s', but got '%s'",
			item.what, item.arg, result.Path())
	}
}

func TestCleanedAbs(t *testing.T) {
	cases := []struct {
		what   string
		full   string
		cDir   string
		name   string
		errStr string
	}{
		{
			what:   "empty",
			full:   "",
			errStr: "doesn't exist",
		},
		{
			what:   "simple",
			full:   "bob",
			errStr: "'bob' doesn't exist",
		},
		{
			what:   "no directory",
			full:   filepath.Join("b", "rrrrrr"),
			errStr: "'b/rrrrrr' doesn't exist",
		},
		{
			what: "longer, ending in file",
			full: filepath.Join("b", "d", "x"),
			cDir: filepath.Join("b", "d"),
			name: "x",
		},
		{
			what: "moar longer, ending in file",
			full: filepath.Join("b", "d", "a", "c", "u"),
			cDir: filepath.Join("b", "d", "a", "c"),
			name: "u",
		},
		{
			what: "directory",
			full: filepath.Join("b", "d"),
			cDir: filepath.Join("b", "d"),
			name: "",
		},
	}

	n := makeLoadedFileTree(t)
	for _, item := range cases {
		cDir, name, err := n.CleanedAbs(item.full)
		if item.errStr != "" {
			if err == nil {
				t.Fatalf("%s; expected error", item.what)
			}
			if !strings.Contains(err.Error(), item.errStr) {
				t.Fatalf("%s; expected err containing '%s', got '%v'",
					item.what, item.errStr, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s; unexpected error: %v", item.what, err)
		}
		if cDir != ConfirmedDir(item.cDir) {
			t.Fatalf("%s; expected cDir=%s, got '%s'", item.what, item.cDir, cDir)
		}
		if name != item.name {
			t.Fatalf("%s; expected name=%s, got '%s'", item.what, item.name, name)
		}
	}
}
