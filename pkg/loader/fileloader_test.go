/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package loader

import (
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/fs"
)

type testData struct {
	path            string
	expectedContent string
}

var testCases = []testData{
	{
		path:            "foo/project/fileA.yaml",
		expectedContent: "fileA content",
	},
	{
		path:            "foo/project/subdir1/fileB.yaml",
		expectedContent: "fileB content",
	},
	{
		path:            "foo/project/subdir2/fileC.yaml",
		expectedContent: "fileC content",
	},
	{
		path:            "foo/project/fileD.yaml",
		expectedContent: "fileD content",
	},
}

func MakeFakeFs(td []testData) fs.FileSystem {
	fSys := fs.MakeFakeFS()
	for _, x := range td {
		fSys.WriteFile("/"+x.path, []byte(x.expectedContent))
	}
	return fSys
}

func TestLoaderLoad(t *testing.T) {
	l1 := NewFileLoaderAtRoot(MakeFakeFs(testCases))
	if "/" != l1.Root() {
		t.Fatalf("incorrect root: '%s'\n", l1.Root())
	}
	for _, x := range testCases {
		b, err := l1.Load(x.path)
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}
		if !reflect.DeepEqual([]byte(x.expectedContent), b) {
			t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
		}
	}
	l2, err := l1.New("foo/project")
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if "/foo/project" != l2.Root() {
		t.Fatalf("incorrect root: %s\n", l2.Root())
	}
	for _, x := range testCases {
		b, err := l2.Load(strings.TrimPrefix(x.path, "foo/project/"))
		if err != nil {
			t.Fatalf("unexpected load error %v", err)
		}
		if !reflect.DeepEqual([]byte(x.expectedContent), b) {
			t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
		}
	}
	l2, err = l1.New("foo/project/") // Assure trailing slash stripped
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if "/foo/project" != l2.Root() {
		t.Fatalf("incorrect root: %s\n", l2.Root())
	}
}

func TestLoaderNewSubDir(t *testing.T) {
	l1, err := NewFileLoaderAtRoot(MakeFakeFs(testCases)).New("foo/project")
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	l2, err := l1.New("subdir1")
	if err != nil {
		t.Fatalf("unexpected err:  %v\n", err)
	}
	if "/foo/project/subdir1" != l2.Root() {
		t.Fatalf("incorrect root: %s\n", l2.Root())
	}
	x := testCases[1]
	b, err := l2.Load("fileB.yaml")
	if err != nil {
		t.Fatalf("unexpected load error %v", err)
	}
	if !reflect.DeepEqual([]byte(x.expectedContent), b) {
		t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
	}
}

func TestLoaderBadRelative(t *testing.T) {
	l1, err := NewFileLoaderAtRoot(MakeFakeFs(testCases)).New("foo/project/subdir1")
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if "/foo/project/subdir1" != l1.Root() {
		t.Fatalf("incorrect root: %s\n", l1.Root())
	}

	// Cannot cd into a file.
	l2, err := l1.New("fileB.yaml")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l2.Root())
	}

	// It's not okay to stay at the same place.
	l2, err = l1.New(".")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l2.Root())
	}

	// It's not okay to go up and back down into same place.
	l2, err = l1.New("../subdir1")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l2.Root())
	}

	// It's not okay to go up via a relative path.
	l2, err = l1.New("..")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l2.Root())
	}

	// It's not okay to go up via an absolute path.
	l2, err = l1.New("/foo/project")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l2.Root())
	}

	// It's not okay to go to the root.
	l2, err = l1.New("/")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l2.Root())
	}

	// It's okay to go up and down to a sibling.
	l2, err = l1.New("../subdir2")
	if err != nil {
		t.Fatalf("unexpected new error %v", err)
	}
	if "/foo/project/subdir2" != l2.Root() {
		t.Fatalf("incorrect root: %s\n", l2.Root())
	}
	x := testCases[2]
	b, err := l2.Load("fileC.yaml")
	if err != nil {
		t.Fatalf("unexpected load error %v", err)
	}
	if !reflect.DeepEqual([]byte(x.expectedContent), b) {
		t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
	}

	// It's not OK to go over to a previously visited directory.
	// Must disallow going back and forth in a cycle.
	l1, err = l2.New("../subdir1")
	if err == nil {
		t.Fatalf("expected err, but got root %s", l1.Root())
	}
}

func TestLoaderMisc(t *testing.T) {
	l := NewFileLoaderAtRoot(MakeFakeFs(testCases))
	_, err := l.New("")
	if err == nil {
		t.Fatalf("Expected error for empty root location not returned")
	}
	_, err = l.New("https://google.com/project")
	if err == nil {
		t.Fatalf("Expected error")
	}
}
