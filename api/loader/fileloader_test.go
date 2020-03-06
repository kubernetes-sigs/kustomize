/// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/konfig"
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

func MakeFakeFs(td []testData) filesys.FileSystem {
	fSys := filesys.MakeFsInMemory()
	for _, x := range td {
		fSys.WriteFile(x.path, []byte(x.expectedContent))
	}
	return fSys
}

func makeLoader() *fileLoader {
	return NewFileLoaderAtRoot(MakeFakeFs(testCases))

}

func TestLoaderLoad(t *testing.T) {
	l1 := makeLoader()
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
	l1, err := makeLoader().New("foo/project")
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
	l1, err := makeLoader().New("foo/project/subdir1")
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
	l2, err = l1.New(filesys.SelfDir)
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
	l := makeLoader()
	_, err := l.New("")
	if err == nil {
		t.Fatalf("Expected error for empty root location not returned")
	}
	_, err = l.New("https://google.com/project")
	if err == nil {
		t.Fatalf("Expected error")
	}
}

const (
	contentOk           = "hi there, i'm OK data"
	contentExteriorData = "i am data from outside the root"
)

// Create a structure like this
//
//   /tmp/kustomize-test-random
//   ├── base
//   │   ├── okayData
//   │   ├── symLinkToOkayData -> okayData
//   │   └── symLinkToExteriorData -> ../exteriorData
//   └── exteriorData
//
func commonSetupForLoaderRestrictionTest() (string, filesys.FileSystem, error) {
	dir, err := ioutil.TempDir("", "kustomize-test-")
	if err != nil {
		return "", nil, err
	}
	fSys := filesys.MakeFsOnDisk()
	fSys.Mkdir(filepath.Join(dir, "base"))

	fSys.WriteFile(
		filepath.Join(dir, "base", "okayData"), []byte(contentOk))

	fSys.WriteFile(
		filepath.Join(dir, "exteriorData"), []byte(contentExteriorData))

	os.Symlink(
		filepath.Join(dir, "base", "okayData"),
		filepath.Join(dir, "base", "symLinkToOkayData"))
	os.Symlink(
		filepath.Join(dir, "exteriorData"),
		filepath.Join(dir, "base", "symLinkToExteriorData"))
	return dir, fSys, nil
}

// Make sure everything works when loading files
// in or below the loader root.
func doSanityChecksAndDropIntoBase(
	t *testing.T, l ifc.Loader) ifc.Loader {
	data, err := l.Load(path.Join("base", "okayData"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != contentOk {
		t.Fatalf("unexpected content: %v", data)
	}
	data, err = l.Load("exteriorData")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != contentExteriorData {
		t.Fatalf("unexpected content: %v", data)
	}

	// Drop in.
	l, err = l.New("base")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reading okayData works.
	data, err = l.Load("okayData")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != contentOk {
		t.Fatalf("unexpected content: %v", data)
	}

	// Reading local symlink to okayData works.
	data, err = l.Load("symLinkToOkayData")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != contentOk {
		t.Fatalf("unexpected content: %v", data)
	}
	return l
}

func TestRestrictionRootOnlyInRealLoader(t *testing.T) {
	dir, fSys, err := commonSetupForLoaderRestrictionTest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(dir)

	var l ifc.Loader

	l = newLoaderOrDie(RestrictionRootOnly, fSys, dir)

	l = doSanityChecksAndDropIntoBase(t, l)

	// Reading symlink to exteriorData fails.
	_, err = l.Load("symLinkToExteriorData")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "is not in or below") {
		t.Fatalf("unexpected err: %v", err)
	}

	// Attempt to read "up" fails, though earlier we were
	// able to read this file when root was "..".
	_, err = l.Load("../exteriorData")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "is not in or below") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRestrictionNoneInRealLoader(t *testing.T) {
	dir, fSys, err := commonSetupForLoaderRestrictionTest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(dir)

	var l ifc.Loader

	l = newLoaderOrDie(RestrictionNone, fSys, dir)

	l = doSanityChecksAndDropIntoBase(t, l)

	// Reading symlink to exteriorData works.
	_, err = l.Load("symLinkToExteriorData")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Attempt to read "up" works.
	_, err = l.Load("../exteriorData")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func splitOnNthSlash(v string, n int) (string, string) {
	left := ""
	for i := 0; i < n; i++ {
		k := strings.Index(v, "/")
		if k < 0 {
			break
		}
		left = left + v[:k+1]
		v = v[k+1:]
	}
	return left[:len(left)-1], v
}

func TestSplit(t *testing.T) {
	p := "a/b/c/d/e/f/g"
	if left, right := splitOnNthSlash(p, 2); left != "a/b" || right != "c/d/e/f/g" {
		t.Fatalf("got left='%s', right='%s'", left, right)
	}
	if left, right := splitOnNthSlash(p, 3); left != "a/b/c" || right != "d/e/f/g" {
		t.Fatalf("got left='%s', right='%s'", left, right)
	}
	if left, right := splitOnNthSlash(p, 6); left != "a/b/c/d/e/f" || right != "g" {
		t.Fatalf("got left='%s', right='%s'", left, right)
	}
}

func TestNewLoaderAtGitClone(t *testing.T) {
	rootUrl := "github.com/someOrg/someRepo"
	pathInRepo := "foo/base"
	url := rootUrl + "/" + pathInRepo
	coRoot := "/tmp"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(coRoot)
	fSys.MkdirAll(coRoot + "/" + pathInRepo)
	fSys.WriteFile(
		coRoot+"/"+pathInRepo+"/"+
			konfig.DefaultKustomizationFileName(),
		[]byte(`
whatever
`))

	repoSpec, err := git.NewRepoSpecFromUrl(url)
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	l, err := newLoaderAtGitClone(
		repoSpec, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(coRoot)), getNothing)
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if coRoot+"/"+pathInRepo != l.Root() {
		t.Fatalf("expected root '%s', got '%s'\n",
			coRoot+"/"+pathInRepo, l.Root())
	}
	if _, err = l.New(url); err == nil {
		t.Fatalf("expected cycle error 1")
	}
	if _, err = l.New(rootUrl + "/" + "foo"); err == nil {
		t.Fatalf("expected cycle error 2")
	}

	pathInRepo = "foo/overlay"
	fSys.MkdirAll(coRoot + "/" + pathInRepo)
	url = rootUrl + "/" + pathInRepo
	l2, err := l.New(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coRoot+"/"+pathInRepo != l2.Root() {
		t.Fatalf("expected root '%s', got '%s'\n",
			coRoot+"/"+pathInRepo, l2.Root())
	}
}

func TestLoaderDisallowsLocalBaseFromRemoteOverlay(t *testing.T) {
	// Define an overlay-base structure in the file system.
	topDir := "/whatever"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir + "/highBase")
	fSys.MkdirAll(cloneRoot + "/foo/base")
	fSys.MkdirAll(cloneRoot + "/foo/overlay")

	var l1 ifc.Loader

	// Establish that a local overlay can navigate
	// to the local bases.
	l1 = newLoaderOrDie(
		RestrictionRootOnly, fSys, cloneRoot+"/foo/overlay")
	if l1.Root() != cloneRoot+"/foo/overlay" {
		t.Fatalf("unexpected root %s", l1.Root())
	}
	l2, err := l1.New("../base")
	if err != nil {
		t.Fatalf("unexpected err:  %v\n", err)
	}
	if l2.Root() != cloneRoot+"/foo/base" {
		t.Fatalf("unexpected root %s", l2.Root())
	}
	l3, err := l2.New("../../../highBase")
	if err != nil {
		t.Fatalf("unexpected err:  %v\n", err)
	}
	if l3.Root() != topDir+"/highBase" {
		t.Fatalf("unexpected root %s", l3.Root())
	}

	// Establish that a Kustomization found in cloned
	// repo can reach (non-remote) bases inside the clone
	// but cannot reach a (non-remote) base outside the
	// clone but legitimately on the local file system.
	// This is to avoid a surprising interaction between
	// a remote K and local files.  The remote K would be
	// non-functional on its own since by definition it
	// would refer to a non-remote base file that didn't
	// exist in its own repository, so presumably the
	// remote K would be deliberately designed to phish
	// for local K's.
	repoSpec, err := git.NewRepoSpecFromUrl(
		"github.com/someOrg/someRepo/foo/overlay")
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	l1, err = newLoaderAtGitClone(
		repoSpec, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), getNothing)
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if l1.Root() != cloneRoot+"/foo/overlay" {
		t.Fatalf("unexpected root %s", l1.Root())
	}
	// This is okay.
	l2, err = l1.New("../base")
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if l2.Root() != cloneRoot+"/foo/base" {
		t.Fatalf("unexpected root %s", l2.Root())
	}
	// This is not okay.
	_, err = l2.New("../../../highBase")
	if err == nil {
		t.Fatalf("expected err")
	}
	if !strings.Contains(err.Error(),
		"base '/whatever/highBase' is outside '/whatever/someClone'") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestLocalLoaderReferencingGitBase(t *testing.T) {
	topDir := "/whatever"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot + "/foo/base")

	root, err := demandDirectoryRoot(fSys, topDir)
	if err != nil {
		t.Fatalf("unexpected err:  %v\n", err)
	}
	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, root, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), getNothing)
	if l1.Root() != topDir {
		t.Fatalf("unexpected root %s", l1.Root())
	}
	l2, err := l1.New("github.com/someOrg/someRepo/foo/base")
	if err != nil {
		t.Fatalf("unexpected err:  %v\n", err)
	}
	if l2.Root() != cloneRoot+"/foo/base" {
		t.Fatalf("unexpected root %s", l2.Root())
	}
}

func TestRepoDirectCycleDetection(t *testing.T) {
	topDir := "/cycles"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot)

	root, err := demandDirectoryRoot(fSys, topDir)
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, root, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), getNothing)
	p1 := "github.com/someOrg/someRepo/foo"
	rs1, err := git.NewRepoSpecFromUrl(p1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	l1.repoSpec = rs1
	_, err = l1.New(p1)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRepoIndirectCycleDetection(t *testing.T) {
	topDir := "/cycles"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot)

	root, err := demandDirectoryRoot(fSys, topDir)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	l0 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, root, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), getNothing)

	p1 := "github.com/someOrg/someRepo1"
	p2 := "github.com/someOrg/someRepo2"

	l1, err := l0.New(p1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	l2, err := l1.New(p2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_, err = l2.New(p1)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("unexpected err: %v", err)
	}
}

// Inspired by https://hassansin.github.io/Unit-Testing-http-client-in-Go
type fakeRoundTripper func(req *http.Request) *http.Response

func (f fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func makeFakeHTTPClient(fn fakeRoundTripper) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

// TestLoaderHTTP test http file loader
func TestLoaderHTTP(t *testing.T) {
	var testCasesFile = []testData{
		{
			path:            "http/file.yaml",
			expectedContent: "file content",
		},
	}

	l1 := NewFileLoaderAtRoot(MakeFakeFs(testCasesFile))
	if "/" != l1.Root() {
		t.Fatalf("incorrect root: '%s'\n", l1.Root())
	}
	for _, x := range testCasesFile {
		b, err := l1.Load(x.path)
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}
		if !reflect.DeepEqual([]byte(x.expectedContent), b) {
			t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
		}
	}

	var testCasesHTTP = []testData{
		{
			path:            "http://example.com/resource.yaml",
			expectedContent: "http content",
		},
		{
			path:            "https://example.com/resource.yaml",
			expectedContent: "https content",
		},
	}

	for _, x := range testCasesHTTP {
		hc := makeFakeHTTPClient(func(req *http.Request) *http.Response {
			u := req.URL.String()
			if x.path != u {
				t.Fatalf("expected URL %s, but got %s", x.path, u)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(x.expectedContent)),
				Header:     make(http.Header),
			}
		})
		l2 := l1
		l2.http = hc
		b, err := l2.Load(x.path)
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}
		if !reflect.DeepEqual([]byte(x.expectedContent), b) {
			t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
		}
	}

	var testCaseUnsupported = []testData{
		{
			path:            "httpsnotreal://example.com/resource.yaml",
			expectedContent: "invalid",
		},
	}
	for _, x := range testCaseUnsupported {
		hc := makeFakeHTTPClient(func(req *http.Request) *http.Response {
			t.Fatalf("unexpected request to URL %s", req.URL.String())
			return nil
		})
		l2 := l1
		l2.http = hc
		_, err := l2.Load(x.path)
		if err == nil {
			t.Fatalf("expect error but get %v", err)
		}
	}
}
