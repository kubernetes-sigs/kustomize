// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepoSpecFromUrl_Permute(t *testing.T) {
	// Generate all many permutations of host, orgRepo, pathName, and ref.
	// Not all combinations make sense, but the parsing is very permissive and
	// we probably stil don't want to break backwards compatibility for things
	// that are unintentionally supported.
	var schemeAuthority = []struct{ raw, normalized string }{
		{"gh:", "gh:"},
		{"GH:", "gh:"},
		{"gitHub.com/", "https://github.com/"},
		{"github.com:", "https://github.com/"},
		{"http://github.com/", "https://github.com/"},
		{"https://github.com/", "https://github.com/"},
		{"hTTps://github.com/", "https://github.com/"},
		{"https://git-codecommit.us-east-2.amazonaws.com/", "https://git-codecommit.us-east-2.amazonaws.com/"},
		{"https://fabrikops2.visualstudio.com/", "https://fabrikops2.visualstudio.com/"},
		{"ssh://git.example.com:7999/", "ssh://git.example.com:7999/"},
		{"git::https://gitlab.com/", "https://gitlab.com/"},
		{"git::http://git.example.com/", "http://git.example.com/"},
		{"git::https://git.example.com/", "https://git.example.com/"},
		{"git@github.com:", "git@github.com:"},
		{"git@github.com/", "git@github.com:"},
		{"org-12345@github.com:", "org-12345@github.com:"},
		{"org-12345@github.com/", "org-12345@github.com:"},
		{"git::git@github.com:", "git@github.com:"},
	}
	var orgRepos = []string{"someOrg/someRepo", "kubernetes/website"}
	var pathNames = []string{"README.md", "foo/krusty.txt", ""}
	var refArgs = []string{"someBranch", "master", "v0.1.0", ""}

	makeURL := func(hostFmt, orgRepo, path, ref string) string {
		if len(path) > 0 {
			orgRepo = filepath.Join(orgRepo, path)
		}
		url := hostFmt + orgRepo
		if ref != "" {
			url += refQuery + ref
		}
		return url
	}

	var bad [][]string
	for _, v := range schemeAuthority {
		hostRaw := v.raw
		hostSpec := v.normalized
		for _, orgRepo := range orgRepos {
			for _, pathName := range pathNames {
				for _, hrefArg := range refArgs {
					uri := makeURL(hostRaw, orgRepo, pathName, hrefArg)
					rs, err := NewRepoSpecFromURL(uri)
					if err != nil {
						t.Errorf("problem %v", err)
						continue
					}
					if rs.Host != hostSpec {
						bad = append(bad, []string{"host", uri, rs.Host, hostSpec})
					}
					if rs.OrgRepo != orgRepo {
						bad = append(bad, []string{"orgRepo", uri, rs.OrgRepo, orgRepo})
					}
					if rs.Path != pathName {
						bad = append(bad, []string{"path", uri, rs.Path, pathName})
					}
					if rs.Ref != hrefArg {
						bad = append(bad, []string{"ref", uri, rs.Ref, hrefArg})
					}
				}
			}
		}
	}
	if len(bad) > 0 {
		for _, tuple := range bad {
			fmt.Printf("\n"+
				"     from uri: %s\n"+
				"  actual %4s: %s\n"+
				"expected %4s: %s\n",
				tuple[1], tuple[0], tuple[2], tuple[0], tuple[3])
		}
		t.Fail()
	}
}

func TestNewRepoSpecFromUrlErrors(t *testing.T) {
	var badData = []struct{ url, error string }{
		{"/tmp", "uri looks like abs path"},
		{"iauhsdiuashduas", "url lacks orgRepo"},
		{"htxxxtp://github.com/", "url lacks host"},
		{"ssh://git.example.com", "url lacks orgRepo"},
		{"git::___", "url lacks orgRepo"},
		{"git::org-12345@github.com:kubernetes-sigs/kustomize", "git protocol on github.com only allows git@ user"},
	}

	for _, testCase := range badData {
		t.Run(testCase.error, func(t *testing.T) {
			_, err := NewRepoSpecFromURL(testCase.url)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), testCase.error) {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestNewRepoSpecFromUrl_Smoke(t *testing.T) {
	// A set of end to end parsing tests that smoke out obvious issues
	// No tests for submodules and timeout as the expectations are hard-coded
	// to the defaults for compactness.
	testcases := []struct {
		name      string
		input     string
		repoSpec  RepoSpec
		cloneSpec string
		absPath   string
		skip      string
	}{
		{
			name:      "t1",
			input:     "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somedir",
			cloneSpec: "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			absPath:   notCloned.Join("somedir"),
			repoSpec: RepoSpec{
				Host:      "https://git-codecommit.us-east-2.amazonaws.com/",
				OrgRepo:   "someorg/somerepo",
				Path:      "somedir",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t2",
			input:     "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somedir?ref=testbranch",
			cloneSpec: "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			absPath:   notCloned.Join("somedir"),
			repoSpec: RepoSpec{
				Host:      "https://git-codecommit.us-east-2.amazonaws.com/",
				OrgRepo:   "someorg/somerepo",
				Path:      "somedir",
				Ref:       "testbranch",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t3",
			input:     "https://fabrikops2.visualstudio.com/someorg/somerepo?ref=master",
			cloneSpec: "https://fabrikops2.visualstudio.com/someorg/somerepo",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "https://fabrikops2.visualstudio.com/",
				OrgRepo:   "someorg/somerepo",
				Ref:       "master",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t4",
			input:     "http://github.com/someorg/somerepo/somedir",
			cloneSpec: "https://github.com/someorg/somerepo.git",
			absPath:   notCloned.Join("somedir"),
			repoSpec: RepoSpec{
				Host:      "https://github.com/",
				OrgRepo:   "someorg/somerepo",
				Path:      "somedir",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t5",
			input:     "git@github.com:someorg/somerepo/somedir",
			cloneSpec: "git@github.com:someorg/somerepo.git",
			absPath:   notCloned.Join("somedir"),
			repoSpec: RepoSpec{
				Host:      "git@github.com:",
				OrgRepo:   "someorg/somerepo",
				Path:      "somedir",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t6",
			input:     "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git?ref=v0.1.0",
			cloneSpec: "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "git@gitlab2.sqtools.ru:10022/",
				OrgRepo:   "infra/kubernetes/thanos-base",
				Ref:       "v0.1.0",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t7",
			input:     "git@bitbucket.org:company/project.git//path?ref=branch",
			cloneSpec: "git@bitbucket.org:company/project.git",
			absPath:   notCloned.Join("path"),
			repoSpec: RepoSpec{
				Host:      "git@bitbucket.org:company/",
				OrgRepo:   "project",
				Path:      "/path",
				Ref:       "branch",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t8",
			input:     "https://itfs.mycompany.com/collection/project/_git/somerepos",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "https://itfs.mycompany.com/collection/project/_git/",
				OrgRepo: "somerepos",
			},
		},
		{
			name:      "t9",
			input:     "https://itfs.mycompany.com/collection/project/_git/somerepos?version=v1.0.0",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "https://itfs.mycompany.com/collection/project/_git/",
				OrgRepo: "somerepos",
				Ref:     "v1.0.0",
			},
		},
		{
			name:      "t10",
			input:     "https://itfs.mycompany.com/collection/project/_git/somerepos/somedir?version=v1.0.0",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.Join("somedir"),
			repoSpec: RepoSpec{
				Host:    "https://itfs.mycompany.com/collection/project/_git/",
				OrgRepo: "somerepos",
				Path:    "/somedir",
				Ref:     "v1.0.0",
			},
		},
		{
			name:      "t11",
			input:     "git::https://itfs.mycompany.com/collection/project/_git/somerepos",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "https://itfs.mycompany.com/collection/project/_git/",
				OrgRepo: "somerepos",
			},
		},
		{
			name:      "t12",
			input:     "https://bitbucket.example.com/scm/project/repository.git",
			cloneSpec: "https://bitbucket.example.com/scm/project/repository.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "https://bitbucket.example.com/",
				OrgRepo:   "scm/project/repository",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t13",
			input:     "ssh://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somepath",
			cloneSpec: "ssh://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			absPath:   notCloned.Join("somepath"),
			repoSpec: RepoSpec{
				Host:      "ssh://git-codecommit.us-east-2.amazonaws.com/",
				OrgRepo:   "someorg/somerepo",
				Path:      "somepath",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t14",
			input:     "git@github.com/someorg/somerepo/somepath",
			cloneSpec: "git@github.com:someorg/somerepo.git",
			absPath:   notCloned.Join("somepath"),
			repoSpec: RepoSpec{
				Host:      "git@github.com:",
				OrgRepo:   "someorg/somerepo",
				Path:      "somepath",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t15",
			input:     "https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6",
			cloneSpec: "https://github.com/kubernetes-sigs/kustomize.git",
			absPath:   notCloned.Join("/examples/multibases/dev"),
			repoSpec: RepoSpec{
				Host:      "https://github.com/",
				OrgRepo:   "kubernetes-sigs/kustomize",
				Path:      "/examples/multibases/dev/",
				Ref:       "v1.0.6",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t16",
			input:     "file://a/b/c/someRepo.git/somepath?ref=someBranch",
			cloneSpec: "file://a/b/c/someRepo.git",
			absPath:   notCloned.Join("somepath"),
			repoSpec: RepoSpec{
				Host:      "file://",
				OrgRepo:   "a/b/c/someRepo",
				Path:      "somepath",
				Ref:       "someBranch",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t17",
			input:     "file://a/b/c/someRepo//somepath?ref=someBranch",
			cloneSpec: "file://a/b/c/someRepo",
			absPath:   notCloned.Join("somepath"),
			repoSpec: RepoSpec{
				Host:    "file://",
				OrgRepo: "a/b/c/someRepo",
				Path:    "somepath",
				Ref:     "someBranch",
			},
		},
		{
			name:      "t18",
			input:     "file://a/b/c/someRepo?ref=someBranch",
			cloneSpec: "file://a/b/c/someRepo",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "file://",
				OrgRepo: "a/b/c/someRepo",
				Ref:     "someBranch",
			},
		},
		{
			name:      "t19",
			input:     "file:///a/b/c/someRepo?ref=someBranch",
			cloneSpec: "file:///a/b/c/someRepo",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "file://",
				OrgRepo: "/a/b/c/someRepo",
				Ref:     "someBranch",
			},
		},
		{
			name:      "t20",
			input:     "ssh://git@github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6",
			cloneSpec: "git@github.com:kubernetes-sigs/kustomize.git",
			absPath:   notCloned.Join("examples/multibases/dev"),
			repoSpec: RepoSpec{
				Host:      "git@github.com:",
				OrgRepo:   "kubernetes-sigs/kustomize",
				Path:      "/examples/multibases/dev",
				Ref:       "v1.0.6",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t21",
			input:     "file:///a/b/c/someRepo",
			cloneSpec: "file:///a/b/c/someRepo",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "file://",
				OrgRepo: "/a/b/c/someRepo",
			},
		},
		{
			name:      "t22",
			input:     "file:///",
			cloneSpec: "file:///",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:    "file://",
				OrgRepo: "/",
			},
		},
		{
			name:      "t23",
			skip:      "the `//` repo separator does not work",
			input:     "https://fake-git-hosting.org/path/to/repo//examples/multibases/dev",
			cloneSpec: "https://fake-git-hosting.org/path/to.git",
			absPath:   notCloned.Join("/examples/multibases/dev"),
			repoSpec: RepoSpec{
				Host:      "https://fake-git-hosting.org/",
				OrgRepo:   "path/to/repo",
				Path:      "/examples/multibases/dev",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t24",
			skip:      "the `//` repo separator does not work",
			input:     "ssh://alice@acme.co/path/to/repo//examples/multibases/dev",
			cloneSpec: "ssh://alice@acme.co/path/to/repo.git",
			absPath:   notCloned.Join("/examples/multibases/dev"),
			repoSpec: RepoSpec{
				Host:      "ssh://alice@acme.co",
				OrgRepo:   "path/to/repo",
				Path:      "/examples/multibases/dev",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t25",
			input:     "https://org-12345@github.com/kubernetes-sigs/kustomize",
			cloneSpec: "org-12345@github.com:kubernetes-sigs/kustomize.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "org-12345@github.com:",
				OrgRepo:   "kubernetes-sigs/kustomize",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t26",
			input:     "ssh://org-12345@github.com/kubernetes-sigs/kustomize",
			cloneSpec: "org-12345@github.com:kubernetes-sigs/kustomize.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "org-12345@github.com:",
				OrgRepo:   "kubernetes-sigs/kustomize",
				GitSuffix: ".git",
			},
		},
		{
			name:      "t27",
			input:     "org-12345@github.com/kubernetes-sigs/kustomize",
			cloneSpec: "org-12345@github.com:kubernetes-sigs/kustomize.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "org-12345@github.com:",
				OrgRepo:   "kubernetes-sigs/kustomize",
				GitSuffix: ".git",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			rs, err := NewRepoSpecFromURL(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.cloneSpec, rs.CloneSpec(), "cloneSpec mismatch")
			assert.Equal(t, tc.absPath, rs.AbsPath(), "absPath mismatch")
			// some values have defaults. Clear them here so test cases remain compact.
			// This means submodules and timeout cannot be tested here. That's fine since
			// they are tested in TestPeelQuery.
			rs.raw = ""
			rs.Dir = ""
			rs.Submodules = false
			rs.Timeout = 0
			assert.Equal(t, &tc.repoSpec, rs)
		})
	}
}

func TestIsAzureHost(t *testing.T) {
	testcases := []struct {
		input  string
		expect bool
	}{
		{
			input:  "https://git-codecommit.us-east-2.amazonaws.com",
			expect: false,
		},
		{
			input:  "ssh://git-codecommit.us-east-2.amazonaws.com",
			expect: false,
		},
		{
			input:  "https://fabrikops2.visualstudio.com/",
			expect: true,
		},
		{
			input:  "https://dev.azure.com/myorg/myproject/",
			expect: true,
		},
	}
	for _, testcase := range testcases {
		actual := isAzureHost(testcase.input)
		if actual != testcase.expect {
			t.Errorf("IsAzureHost: expected %v, but got %v on %s", testcase.expect, actual, testcase.input)
		}
	}
}

func TestPeelQuery(t *testing.T) {
	testcases := []struct {
		name       string
		input      string
		path       string
		ref        string
		submodules bool
		timeout    time.Duration
	}{
		{
			name: "t1",
			// All empty.
			input:      "somerepos",
			path:       "somerepos",
			ref:        "",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "t2",
			input:      "somerepos?ref=v1.0.0",
			path:       "somerepos",
			ref:        "v1.0.0",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "t3",
			input:      "somerepos?version=master",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "t4",
			// A ref value takes precedence over a version value.
			input:      "somerepos?version=master&ref=v1.0.0",
			path:       "somerepos",
			ref:        "v1.0.0",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "t5",
			// Empty submodules value uses default.
			input:      "somerepos?version=master&submodules=",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "t6",
			// Malformed submodules value uses default.
			input:      "somerepos?version=master&submodules=maybe",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "t7",
			input:      "somerepos?version=master&submodules=true",
			path:       "somerepos",
			ref:        "master",
			submodules: true,
			timeout:    defaultTimeout,
		},
		{
			name:       "t8",
			input:      "somerepos?version=master&submodules=false",
			path:       "somerepos",
			ref:        "master",
			submodules: false,
			timeout:    defaultTimeout,
		},
		{
			name: "t9",
			// Empty timeout value uses default.
			input:      "somerepos?version=master&timeout=",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "t10",
			// Malformed timeout value uses default.
			input:      "somerepos?version=master&timeout=jiffy",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "t11",
			// Zero timeout value uses default.
			input:      "somerepos?version=master&timeout=0",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "t12",
			input:      "somerepos?version=master&timeout=0s",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "t13",
			input:      "somerepos?version=master&timeout=61",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    61 * time.Second,
		},
		{
			name:       "t14",
			input:      "somerepos?version=master&timeout=1m1s",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    61 * time.Second,
		},
		{
			name:       "t15",
			input:      "somerepos?version=master&submodules=false&timeout=1m1s",
			path:       "somerepos",
			ref:        "master",
			submodules: false,
			timeout:    61 * time.Second,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path, ref, timeout, submodules := peelQuery(tc.input)
			assert.Equal(t, tc.path, path, "path mismatch")
			assert.Equal(t, tc.ref, ref, "ref mismatch")
			assert.Equal(t, tc.timeout, timeout, "timeout mismatch")
			assert.Equal(t, tc.submodules, submodules, "submodules mismatch")
		})
	}
}

func TestIsAWSHost(t *testing.T) {
	testcases := []struct {
		input  string
		expect bool
	}{
		{
			input:  "https://git-codecommit.us-east-2.amazonaws.com",
			expect: true,
		},
		{
			input:  "ssh://git-codecommit.us-east-2.amazonaws.com",
			expect: true,
		},
		{
			input:  "git@github.com:",
			expect: false,
		},
		{
			input:  "http://github.com/",
			expect: false,
		},
	}
	for _, testcase := range testcases {
		actual := isAWSHost(testcase.input)
		if actual != testcase.expect {
			t.Errorf("IsAWSHost: expected %v, but got %v on %s", testcase.expect, actual, testcase.input)
		}
	}
}
