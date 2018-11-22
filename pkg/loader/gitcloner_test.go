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
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
)

func TestIsRepoURL(t *testing.T) {

	testcases := []struct {
		input    string
		expected bool
	}{
		{
			input:    "https://github.com/org/repo",
			expected: true,
		},
		{
			input:    "github.com/org/repo",
			expected: true,
		},
		{
			input:    "git@github.com:org/repo",
			expected: true,
		},
		{
			input:    "gh:org/repo",
			expected: true,
		},
		{
			input:    "git::https://gitlab.com/org/repo",
			expected: true,
		},
		{
			input:    "/github.com/org/repo",
			expected: false,
		},
		{
			input:    "/abs/path/to/file",
			expected: false,
		},
		{
			input:    "../relative",
			expected: false,
		},
		{
			input:    "foo",
			expected: false,
		},
		{
			input:    ".",
			expected: false,
		},
		{
			input:    "",
			expected: false,
		},
	}
	for _, tc := range testcases {
		actual := isRepoUrl(tc.input)
		if actual != tc.expected {
			t.Errorf("unexpected error: unexpected result %t for input %s", actual, tc.input)
		}
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
	path := "a/b/c/d/e/f/g"
	if left, right := splitOnNthSlash(path, 2); left != "a/b" || right != "c/d/e/f/g" {
		t.Fatalf("got left='%s', right='%s'", left, right)
	}
	if left, right := splitOnNthSlash(path, 3); left != "a/b/c" || right != "d/e/f/g" {
		t.Fatalf("got left='%s', right='%s'", left, right)
	}
	if left, right := splitOnNthSlash(path, 6); left != "a/b/c/d/e/f" || right != "g" {
		t.Fatalf("got left='%s', right='%s'", left, right)
	}
}

// makeFakeGitCloner returns a cloner that ignores the
// URL argument and returns a path in a fake file system
// that should already hold the 'repo' contents.
func makeFakeGitCloner(t *testing.T, fSys fs.FileSystem, coRoot string) gitCloner {
	if !fSys.IsDir(coRoot) {
		t.Fatalf("expecting a directory at '%s'", coRoot)
	}
	return func(url string) (
		checkoutDir string, pathInCoDir string, err error) {
		_, path := splitOnNthSlash(url, 3)
		if !fSys.IsDir(coRoot + "/" + path) {
			t.Fatalf("expecting a directory at '%s'/'%s'",
				coRoot, path)
		}
		return coRoot, path, nil
	}
}

func TestGitLoader(t *testing.T) {
	rootUrl := "github.com/someOrg/someRepo"
	pathInRepo := "foo/base"
	url := rootUrl + "/" + pathInRepo
	if !isRepoUrl(url) {
		t.Fatalf("'%s' should be accepted as a repo url", url)
	}

	coRoot := "/tmp"
	fSys := fs.MakeFakeFS()
	fSys.MkdirAll(coRoot)
	fSys.MkdirAll(coRoot + "/" + pathInRepo)
	fSys.WriteFile(
		coRoot+"/"+pathInRepo+"/"+constants.KustomizationFileName,
		[]byte(`
whatever
`))
	l, err := newGitLoader(
		url, fSys, []string{},
		makeFakeGitCloner(t, fSys, coRoot))
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}
	if coRoot+"/"+pathInRepo != l.Root() {
		t.Fatalf("expected root '%s', got '%s'\n",
			coRoot+"/"+pathInRepo, l.Root())
	}
	if _, err = l.New(url); err == nil {
		t.Fatalf("expected cycle error")
	}
	if _, err = l.New(rootUrl + "/" + "foo"); err == nil {
		t.Fatalf("expected cycle error")
	}

	pathInRepo = "foo/overlay"
	fSys.MkdirAll(coRoot + "/" + pathInRepo)
	url = rootUrl + "/" + pathInRepo
	if !isRepoUrl(url) {
		t.Fatalf("'%s' should be accepted as a repo url", url)
	}
	l2, err := l.New(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coRoot+"/"+pathInRepo != l2.Root() {
		t.Fatalf("expected root '%s', got '%s'\n",
			coRoot+"/"+pathInRepo, l2.Root())
	}
}

var repoNames = []string{"someOrg/someRepo", "kubernetes/website"}

var paths = []string{"", "README.md", "foo/index.md"}

var extractFmts = []string{
	"gh:%s",
	"GH:%s",
	"gitHub.com/%s",
	"https://github.com/%s",
	"hTTps://github.com/%s",
	"git::https://gitlab.com/%s",
	"git@gitHUB.com:%s.git",
	"github.com:%s",
}

func TestExtractGithubRepoName(t *testing.T) {
	for _, repoName := range repoNames {
		for _, pathName := range paths {
			for _, extractFmt := range extractFmts {
				spec := repoName
				if len(pathName) > 0 {
					spec = filepath.Join(spec, pathName)
				}
				input := fmt.Sprintf(extractFmt, spec)
				if !isRepoUrl(input) {
					t.Errorf("Should smell like github arg: %s\n", input)
					continue
				}
				repo, path, err := extractGithubRepoName(input)
				if err != nil {
					t.Errorf("problem %v", err)
				}
				if repo != repoName {
					t.Errorf("\n"+
						"       from %s\n"+
						"    gotRepo %s\n"+
						"desiredRepo %s\n", input, repo, repoName)
				}
				if path != pathName {
					t.Errorf("\n"+
						"       from %s\n"+
						"    gotPath %s\n"+
						"desiredPath %s\n", input, path, pathName)
				}
			}
		}
	}
}
