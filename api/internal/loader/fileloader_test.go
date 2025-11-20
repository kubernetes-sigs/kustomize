/// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestIsRemoteFile(t *testing.T) {
	cases := map[string]struct {
		url   string
		valid bool
	}{
		"https file": {
			"https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/examples/helloWorld/configMap.yaml",
			true,
		},
		"malformed https": {
			// TODO(annasong): Maybe we want to fix this. Needs more research.
			"https:/raw.githubusercontent.com/kubernetes-sigs/kustomize/master/examples/helloWorld/configMap.yaml",
			true,
		},
		"https dir": {
			"https://github.com/kubernetes-sigs/kustomize//examples/helloWorld/",
			true,
		},
		"no scheme": {
			"github.com/kubernetes-sigs/kustomize//examples/helloWorld/",
			false,
		},
		"ssh": {
			"ssh://git@github.com/kubernetes-sigs/kustomize.git",
			false,
		},
		"local": {
			"pod.yaml",
			false,
		},
	}
	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.valid, IsRemoteFile(test.url))
		})
	}
}

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

func makeLoader() *FileLoader {
	return NewLoaderOrDie(
		RestrictionRootOnly, MakeFakeFs(testCases), filesys.Separator)
}

func TestLoaderLoad(t *testing.T) {
	require := require.New(t)

	l1 := makeLoader()
	repo := l1.Repo()
	require.Empty(repo)
	require.Equal("/", l1.Root())

	for _, x := range testCases {
		b, err := l1.Load(x.path)
		require.NoError(err)

		if !reflect.DeepEqual([]byte(x.expectedContent), b) {
			t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
		}
	}
	l2, err := l1.New("foo/project")
	require.NoError(err)

	repo = l2.Repo()
	require.Empty(repo)
	require.Equal("/foo/project", l2.Root())

	for _, x := range testCases {
		b, err := l2.Load(strings.TrimPrefix(x.path, "foo/project/"))
		require.NoError(err)

		if !reflect.DeepEqual([]byte(x.expectedContent), b) {
			t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
		}
	}
	l2, err = l1.New("foo/project/") // Assure trailing slash stripped
	require.NoError(err)
	require.Equal("/foo/project", l2.Root())
}

func TestLoaderNewSubDir(t *testing.T) {
	require := require.New(t)

	l1, err := makeLoader().New("foo/project")
	require.NoError(err)

	l2, err := l1.New("subdir1")
	require.NoError(err)
	require.Equal("/foo/project/subdir1", l2.Root())

	x := testCases[1]
	b, err := l2.Load("fileB.yaml")
	require.NoError(err)

	if !reflect.DeepEqual([]byte(x.expectedContent), b) {
		t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
	}
}

func TestLoaderBadRelative(t *testing.T) {
	require := require.New(t)

	l1, err := makeLoader().New("foo/project/subdir1")
	require.NoError(err)
	require.Equal("/foo/project/subdir1", l1.Root())

	// Cannot cd into a file.
	l2, err := l1.New("fileB.yaml")
	require.Error(err)

	// It's not okay to stay at the same place.
	l2, err = l1.New(filesys.SelfDir)
	require.Error(err)

	// It's not okay to go up and back down into same place.
	l2, err = l1.New("../subdir1")
	require.Error(err)

	// It's not okay to go up via a relative path.
	l2, err = l1.New("..")
	require.Error(err)

	// It's not okay to go up via an absolute path.
	l2, err = l1.New("/foo/project")
	require.Error(err)

	// It's not okay to go to the root.
	l2, err = l1.New("/")
	require.Error(err)

	// It's okay to go up and down to a sibling.
	l2, err = l1.New("../subdir2")
	require.NoError(err)
	require.Equal("/foo/project/subdir2", l2.Root())

	x := testCases[2]
	b, err := l2.Load("fileC.yaml")
	require.NoError(err)
	if !reflect.DeepEqual([]byte(x.expectedContent), b) {
		t.Fatalf("in load expected %s, but got %s", x.expectedContent, b)
	}

	// It's not OK to go over to a previously visited directory.
	// Must disallow going back and forth in a cycle.
	l1, err = l2.New("../subdir1")
	require.Error(err)
}

func TestNewEmptyLoader(t *testing.T) {
	_, err := makeLoader().New("")
	require.Error(t, err)
}

func TestNewRemoteLoaderDoesNotExist(t *testing.T) {
	_, err := makeLoader().New("https://example.com/org/repo")
	require.ErrorContains(t, err, "fetch")
}

func TestLoaderLocalScheme(t *testing.T) {
	// It is unlikely but possible for a reference with a url scheme to
	// actually refer to a local file or directory.
	t.Run("file", func(t *testing.T) {
		fSys, dir := setupOnDisk(t)
		parts := []string{
			"ssh:",
			"resource.yaml",
		}
		require.NoError(t, fSys.Mkdir(dir.Join(parts[0])))
		const content = "resource config"
		require.NoError(t, fSys.WriteFile(
			dir.Join(filepath.Join(parts...)),
			[]byte(content),
		))
		actualContent, err := NewLoaderOrDie(RestrictionRootOnly,
			fSys,
			dir.String(),
		).Load(strings.Join(parts, "//"))
		require.NoError(t, err)
		require.Equal(t, content, string(actualContent))
	})
	t.Run("directory", func(t *testing.T) {
		fSys, dir := setupOnDisk(t)
		parts := []string{
			"https:",
			"root",
		}
		require.NoError(t, fSys.MkdirAll(dir.Join(filepath.Join(parts...))))
		ldr, err := NewLoaderOrDie(RestrictionRootOnly,
			fSys,
			dir.String(),
		).New(strings.Join(parts, "//"))
		require.NoError(t, err)
		require.Empty(t, ldr.Repo())
	})
}

const (
	contentOk           = "hi there, i'm OK data"
	contentExteriorData = "i am data from outside the root"
)

// Create a structure like this
//
//	/tmp/kustomize-test-random
//	├── base
//	│   ├── okayData
//	│   ├── symLinkToOkayData -> okayData
//	│   └── symLinkToExteriorData -> ../exteriorData
//	└── exteriorData
func commonSetupForLoaderRestrictionTest(t *testing.T) (string, filesys.FileSystem) {
	t.Helper()
	fSys, tmpDir := setupOnDisk(t)
	dir := tmpDir.String()

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
	return dir, fSys
}

// Make sure everything works when loading files
// in or below the loader root.
func doSanityChecksAndDropIntoBase(
	t *testing.T, l ifc.Loader) ifc.Loader {
	t.Helper()
	require := require.New(t)

	data, err := l.Load(path.Join("base", "okayData"))
	require.NoError(err)
	require.Equal(contentOk, string(data))

	data, err = l.Load("exteriorData")
	require.NoError(err)
	require.Equal(contentExteriorData, string(data))

	// Drop in.
	l, err = l.New("base")
	require.NoError(err)

	// Reading okayData works.
	data, err = l.Load("okayData")
	require.NoError(err)
	require.Equal(contentOk, string(data))

	// Reading local symlink to okayData works.
	data, err = l.Load("symLinkToOkayData")
	require.NoError(err)
	require.Equal(contentOk, string(data))

	return l
}

func TestRestrictionRootOnlyInRealLoader(t *testing.T) {
	require := require.New(t)
	dir, fSys := commonSetupForLoaderRestrictionTest(t)

	var l ifc.Loader

	l = NewLoaderOrDie(RestrictionRootOnly, fSys, dir)

	l = doSanityChecksAndDropIntoBase(t, l)

	// Reading symlink to exteriorData fails.
	_, err := l.Load("symLinkToExteriorData")
	require.Error(err)
	require.Contains(err.Error(), "is not in or below")

	// Attempt to read "up" fails, though earlier we were
	// able to read this file when root was "..".
	_, err = l.Load("../exteriorData")
	require.Error(err)
	require.Contains(err.Error(), "is not in or below")
}

func TestRestrictionNoneInRealLoader(t *testing.T) {
	dir, fSys := commonSetupForLoaderRestrictionTest(t)

	var l ifc.Loader

	l = NewLoaderOrDie(RestrictionNone, fSys, dir)

	l = doSanityChecksAndDropIntoBase(t, l)

	// Reading symlink to exteriorData works.
	_, err := l.Load("symLinkToExteriorData")
	require.NoError(t, err)

	// Attempt to read "up" works.
	_, err = l.Load("../exteriorData")
	require.NoError(t, err)
}

func splitOnNthSlash(v string, n int) (string, string) {
	left := ""
	for i := 0; i < n; i++ {
		k := strings.Index(v, "/")
		if k < 0 {
			break
		}
		left += v[:k+1]
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
	require := require.New(t)

	rootURL := "github.com/someOrg/someRepo"
	pathInRepo := "foo/base"
	url := rootURL + "/" + pathInRepo
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

	repoSpec, err := git.NewRepoSpecFromURL(url)
	require.NoError(err)

	l, err := newLoaderAtGitClone(
		repoSpec, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(coRoot)))
	require.NoError(err)
	repo := l.Repo()
	require.Equal(coRoot, repo)
	require.Equal(coRoot+"/"+pathInRepo, l.Root())

	_, err = l.New(url)
	require.Error(err)

	_, err = l.New(rootURL + "/" + "foo")
	require.Error(err)

	pathInRepo = "foo/overlay"
	fSys.MkdirAll(coRoot + "/" + pathInRepo)
	url = rootURL + "/" + pathInRepo
	l2, err := l.New(url)
	require.NoError(err)

	repo = l2.Repo()
	require.Equal(coRoot, repo)
	require.Equal(coRoot+"/"+pathInRepo, l2.Root())
}

func TestNewLoaderAtOciPull(t *testing.T) {
	require := require.New(t)

	rootURL := "github.com/someOrg/someRepo"
	pathInRepo := "foo/base"
	url := rootURL + "//" + pathInRepo
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

	repoSpec, err := oci.NewRepoSpecFromURL("oci://" + url)
	require.NoError(err)

	l, err := newLoaderAtOciPull(
		repoSpec, fSys, nil,
		oci.DoNothingPuller(filesys.ConfirmedDir(coRoot)))
	require.NoError(err)
	repo := l.Repo()
	require.Equal(coRoot, repo)
	require.Equal(coRoot+"/"+pathInRepo, l.Root())

	_, err = l.New("oci://" + url)
	require.Error(err)

	_, err = l.New("oci://" + rootURL + "//" + "foo")
	require.Error(err)

	pathInRepo = "foo/overlay"
	fSys.MkdirAll(coRoot + "/" + pathInRepo)
	url = rootURL + "//" + pathInRepo
	l2, err := l.New("oci://" + url)
	require.NoError(err)

	repo = l2.Repo()
	require.Equal(coRoot, repo)
	require.Equal(coRoot+"/"+pathInRepo, l2.Root())
}

func TestLoaderDisallowsLocalBaseFromRemoteOverlay(t *testing.T) {
	require := require.New(t)

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
	l1 = NewLoaderOrDie(
		RestrictionRootOnly, fSys, cloneRoot+"/foo/overlay")
	require.Equal(cloneRoot+"/foo/overlay", l1.Root())

	l2, err := l1.New("../base")
	require.NoError(nil)
	require.Equal(cloneRoot+"/foo/base", l2.Root())

	l3, err := l2.New("../../../highBase")
	require.NoError(err)
	require.Equal(topDir+"/highBase", l3.Root())

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
	repoSpec, err := git.NewRepoSpecFromURL(
		"github.com/someOrg/someRepo/foo/overlay")
	require.NoError(err)

	l1, err = newLoaderAtGitClone(
		repoSpec, fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)))
	require.NoError(err)
	require.Equal(cloneRoot+"/foo/overlay", l1.Root())

	// This is okay.
	l2, err = l1.New("../base")
	require.NoError(err)
	repo := l2.Repo()
	require.Empty(repo)
	require.Equal(cloneRoot+"/foo/base", l2.Root())

	// This is not okay.
	_, err = l2.New("../../../highBase")
	require.Error(err)
	require.Contains(err.Error(),
		"base '/whatever/highBase' is outside '/whatever/someClone'")
}

func TestLoaderDisallowsLocalBaseFromOciRemoteOverlay(t *testing.T) {
	require := require.New(t)

	// Define an overlay-base structure in the file system.
	topDir := "/whatever"
	pullRoot := topDir + "/somePull"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir + "/highBase")
	fSys.MkdirAll(pullRoot + "/foo/base")
	fSys.MkdirAll(pullRoot + "/foo/overlay")

	var l1 ifc.Loader

	// Establish that a local overlay can navigate
	// to the local bases.
	l1 = NewLoaderOrDie(
		RestrictionRootOnly, fSys, pullRoot+"/foo/overlay")
	require.Equal(pullRoot+"/foo/overlay", l1.Root())

	l2, err := l1.New("../base")
	require.NoError(nil)
	require.Equal(pullRoot+"/foo/base", l2.Root())

	l3, err := l2.New("../../../highBase")
	require.NoError(err)
	require.Equal(topDir+"/highBase", l3.Root())

	// Establish that a Kustomization found in pulled oci
	// repo can reach (non-remote) bases inside the pull
	// but cannot reach a (non-remote) base outside the
	// pull but legitimately on the local file system.
	// This is to avoid a surprising interaction between
	// a remote K and local files.  The remote K would be
	// non-functional on its own since by definition it
	// would refer to a non-remote base file that didn't
	// exist in its own repository, so presumably the
	// remote K would be deliberately designed to phish
	// for local K's.
	repoSpec, err := oci.NewRepoSpecFromURL(
		"oci://github.com/someOrg/someRepo//foo/overlay")
	require.NoError(err)

	l1, err = newLoaderAtOciPull(
		repoSpec, fSys, nil,
		oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot)))
	require.NoError(err)
	require.Equal(pullRoot+"/foo/overlay", l1.Root())

	// This is okay.
	l2, err = l1.New("../base")
	require.NoError(err)
	repo := l2.Repo()
	require.Empty(repo)
	require.Equal(pullRoot+"/foo/base", l2.Root())

	// This is not okay.
	_, err = l2.New("../../../highBase")
	require.Error(err)
	require.Contains(err.Error(),
		"base '/whatever/highBase' is outside '/whatever/somePull'")
}

func TestLoaderDisallowsRemoteBaseExitRepo(t *testing.T) {
	fSys, dir := setupOnDisk(t)

	repo := dir.Join("repo")
	require.NoError(t, fSys.Mkdir(repo))

	base := filepath.Join(repo, "base")
	require.NoError(t, os.Symlink(dir.String(), base))

	repoSpec, err := git.NewRepoSpecFromURL("https://github.com/org/repo/base")
	require.NoError(t, err)

	_, err = newLoaderAtGitClone(repoSpec, fSys, nil, git.DoNothingCloner(filesys.ConfirmedDir(repo)))
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("%q refers to directory outside of repo %q", base, repo))
}

func TestLoaderDisallowsRemoteBaseExitOciRepo(t *testing.T) {
	fSys, dir := setupOnDisk(t)

	repo := dir.Join("repo")
	require.NoError(t, fSys.Mkdir(repo))

	base := filepath.Join(repo, "base")
	require.NoError(t, os.Symlink(dir.String(), base))

	repoSpec, err := oci.NewRepoSpecFromURL("oci://github.com/org/repo//base")
	require.NoError(t, err)

	_, err = newLoaderAtOciPull(repoSpec, fSys, nil, oci.DoNothingPuller(filesys.ConfirmedDir(repo)))
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("%q refers to directory outside of repo %q", base, repo))
}

func TestLocalLoaderReferencingGitBase(t *testing.T) {
	require := require.New(t)

	topDir := "/whatever"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot + "/foo/base")

	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), nil)
	require.Equal(topDir, l1.Root())

	l2, err := l1.New("github.com/someOrg/someRepo/foo/base")
	require.NoError(err)
	repo := l2.Repo()
	require.Equal(cloneRoot, repo)
	require.Equal(cloneRoot+"/foo/base", l2.Root())
}

func TestLocalLoaderReferencingOciBase(t *testing.T) {
	require := require.New(t)

	topDir := "/whatever"
	pullRoot := topDir + "/somePull"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(pullRoot + "/foo/base")

	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		nil, oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot)))
	require.Equal(topDir, l1.Root())

	l2, err := l1.New("oci://github.com/someOrg/someRepo//foo/base")
	require.NoError(err)
	repo := l2.Repo()
	require.Equal(pullRoot, repo)
	require.Equal(pullRoot+"/foo/base", l2.Root())
}

func TestRepoDirectCycleDetection(t *testing.T) {
	require := require.New(t)

	topDir := "/cycles"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot)

	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), nil)
	p1 := "github.com/someOrg/someRepo/foo"
	rs1, err := git.NewRepoSpecFromURL(p1)
	require.NoError(err)

	l1.repoSpec = rs1
	_, err = l1.New(p1)
	require.Error(err)
	require.Contains(err.Error(), "cycle detected")
}

func TestOciRepoDirectCycleDetection(t *testing.T) {
	require := require.New(t)

	topDir := "/cycles"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot)

	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		nil, oci.DoNothingPuller(filesys.ConfirmedDir(cloneRoot)))
	p1 := "oci://github.com/someOrg/someRepo/foo"
	rs1, err := oci.NewRepoSpecFromURL(p1)
	require.NoError(err)

	l1.ociSpec = rs1
	_, err = l1.New(p1)
	require.Error(err)
	require.Contains(err.Error(), "cycle detected")
}

func TestRepoIndirectCycleDetection(t *testing.T) {
	require := require.New(t)

	topDir := "/cycles"
	cloneRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(cloneRoot)

	l0 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)), nil)

	p1 := "github.com/someOrg/someRepo1"
	p2 := "github.com/someOrg/someRepo2"

	l1, err := l0.New(p1)
	require.NoError(err)

	l2, err := l1.New(p2)
	require.NoError(err)

	_, err = l2.New(p1)
	require.Error(err)
	require.Contains(err.Error(), "cycle detected")
}

func TestOciRepoIndirectCycleDetection(t *testing.T) {
	require := require.New(t)

	topDir := "/cycles"
	pullRoot := topDir + "/someClone"
	fSys := filesys.MakeFsInMemory()
	fSys.MkdirAll(topDir)
	fSys.MkdirAll(pullRoot)

	l0 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		nil, oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot)))

	p1 := "oci://github.com/someOrg/someRepo1"
	p2 := "oci://github.com/someOrg/someRepo2"

	l1, err := l0.New(p1)
	require.NoError(err)

	l2, err := l1.New(p2)
	require.NoError(err)

	_, err = l2.New(p1)
	require.Error(err)
	require.Contains(err.Error(), "cycle detected")
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
	require := require.New(t)

	var testCasesFile = []testData{
		{
			path:            "http/file.yaml",
			expectedContent: "file content",
		},
	}

	l1 := NewLoaderOrDie(
		RestrictionRootOnly, MakeFakeFs(testCasesFile), filesys.Separator)
	require.Equal("/", l1.Root())

	for _, x := range testCasesFile {
		b, err := l1.Load(x.path)
		require.NoError(err)
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
			require.Equal(x.path, u)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(x.expectedContent)),
				Header:     make(http.Header),
			}
		})
		l2 := l1
		l2.http = hc
		b, err := l2.Load(x.path)
		require.NoError(err)
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
		require.Error(err)
	}
}

// setupOnDisk sets up a file system on disk and directory that is cleaned after
// test completion.
// TODO(annasong): Move all loader tests that require real file system into
// api/krusty.
func setupOnDisk(t *testing.T) (filesys.FileSystem, filesys.ConfirmedDir) {
	t.Helper()

	fSys := filesys.MakeFsOnDisk()
	dir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = fSys.RemoveAll(dir.String())
	})
	return fSys, dir
}

// TestLoaderHTTPMalformedURL tests that malformed URLs are properly handled
// to prevent infinite loops in http.Client.Get
func TestLoaderHTTPMalformedURL(t *testing.T) {
	require := require.New(t)
	malformedURL := "https://example.com/example?ref=main - ../../example/example.yaml"
	l1 := NewLoaderOrDie(
		RestrictionRootOnly, MakeFakeFs([]testData{}), filesys.Separator)
	_, err := l1.Load(malformedURL)
	require.Error(err)
	require.Equal("HTTP Error: status code 500 (Internal Server Error)", err.Error())
}
