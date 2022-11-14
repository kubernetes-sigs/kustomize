// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"path/filepath"
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
	}
	var orgRepos = []string{"someOrg/someRepo", "kubernetes/website"}
	var pathNames = []string{"README.md", "foo/krusty.txt", ""}
	var refArgs = []string{"group/version", "someBranch", "master", "v0.1.0", ""}

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

	var i int
	for _, v := range schemeAuthority {
		hostRaw := v.raw
		hostSpec := v.normalized
		for _, orgRepo := range orgRepos {
			for _, pathName := range pathNames {
				for _, hrefArg := range refArgs {
					t.Run(fmt.Sprintf("t%d", i), func(t *testing.T) {
						uri := makeURL(hostRaw, orgRepo, pathName, hrefArg)
						rs, err := NewRepoSpecFromURL(uri)
						require.NoErrorf(t, err, "unexpected error creating RepoSpec for uri %s", uri)
						assert.Equal(t, hostSpec, rs.Host, "unexpected host for uri %s", uri)
						assert.Equal(t, orgRepo, rs.OrgRepo, "unexpected orgRepo for uri %s", uri)
						assert.Equal(t, pathName, rs.Path, "unexpected path for uri %s", uri)
						assert.Equal(t, hrefArg, rs.Ref, "unexpected ref for uri %s", uri)
					})
					i++
				}
			}
		}
	}
}

func TestNewRepoSpecFromUrlErrors(t *testing.T) {
	badData := map[string]struct {
		url, error string
	}{
		"absolute_path": {
			"/tmp",
			"uri looks like abs path",
		},
		"no_slashes": {
			"iauhsdiuashduas",
			"url lacks orgRepo",
		},
		"bad_scheme": {
			"htxxxtp://github.com/",
			"url lacks host",
		},
		"no_org_repo": {
			"ssh://git.example.com",
			"url lacks orgRepo",
		},
		"hashicorp_git_only": {
			"git::___",
			"url lacks orgRepo",
		},
		"query_after_host": {
			"https://host?ref=group/version/minor_version",
			"url lacks orgRepo",
		},
	}

	for name, testCase := range badData {
		t.Run(name, func(t *testing.T) {
			_, err := NewRepoSpecFromURL(testCase.url)
			require.Error(t, err)
			require.Contains(t, err.Error(), testCase.error)
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
			name:      "query_slash",
			input:     "https://authority/org/repo?ref=group/version",
			cloneSpec: "https://authority/org/repo.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "https://authority/",
				OrgRepo:   "org/repo",
				Ref:       "group/version",
				GitSuffix: ".git",
			},
		},
		{
			name:      "query_git_delimiter",
			input:     "https://authority/org/repo/?ref=includes_git/for_some_reason",
			cloneSpec: "https://authority/org/repo.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "https://authority/",
				OrgRepo:   "org/repo",
				Ref:       "includes_git/for_some_reason",
				GitSuffix: ".git",
			},
		},
		{
			name:      "query_git_suffix",
			input:     "https://authority/org/repo/?ref=includes.git/for_some_reason",
			cloneSpec: "https://authority/org/repo.git",
			absPath:   notCloned.String(),
			repoSpec: RepoSpec{
				Host:      "https://authority/",
				OrgRepo:   "org/repo",
				Ref:       "includes.git/for_some_reason",
				GitSuffix: ".git",
			},
		},
		{
			name:      "non_parsable_path",
			input:     "https://authority/org/repo/%-invalid-uri-so-not-parsable-by-net/url.Parse",
			cloneSpec: "https://authority/org/repo.git",
			absPath:   notCloned.Join("%-invalid-uri-so-not-parsable-by-net/url.Parse"),
			repoSpec: RepoSpec{
				Host:      "https://authority/",
				OrgRepo:   "org/repo",
				Path:      "%-invalid-uri-so-not-parsable-by-net/url.Parse",
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
			// they are tested in TestParseQuery.
			rs.raw = ""
			rs.Dir = ""
			rs.Submodules = false
			rs.Timeout = 0
			assert.Equal(t, &tc.repoSpec, rs)
		})
	}
}

func TestNewRepoSpecFromURL_DefaultQueryParams(t *testing.T) {
	repoSpec, err := NewRepoSpecFromURL("https://github.com/org/repo")
	require.NoError(t, err)
	require.Equal(t, defaultSubmodules, repoSpec.Submodules)
	require.Equal(t, defaultTimeout, repoSpec.Timeout)
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

func TestParseQuery(t *testing.T) {
	testcases := []struct {
		name       string
		input      string
		ref        string
		submodules bool
		timeout    time.Duration
	}{
		{
			name:       "empty",
			input:      "",
			ref:        "",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "ref",
			input:      "ref=v1.0.0",
			ref:        "v1.0.0",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "ref_slash",
			input:      "ref=kustomize/v4.5.7",
			ref:        "kustomize/v4.5.7",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "version",
			input:      "version=master",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "ref_and_version",
			// A ref value takes precedence over a version value.
			input:      "version=master&ref=v1.0.0",
			ref:        "v1.0.0",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "empty_submodules",
			// Empty submodules value uses default.
			input:      "version=master&submodules=",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "bad_submodules",
			// Malformed submodules value uses default.
			input:      "version=master&submodules=maybe",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "submodules_true",
			input:      "version=master&submodules=true",
			ref:        "master",
			submodules: true,
			timeout:    defaultTimeout,
		},
		{
			name:       "submodules_false",
			input:      "version=master&submodules=false",
			ref:        "master",
			submodules: false,
			timeout:    defaultTimeout,
		},
		{
			name: "empty_timeout",
			// Empty timeout value uses default.
			input:      "version=master&timeout=",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "bad_timeout",
			// Malformed timeout value uses default.
			input:      "version=master&timeout=jiffy",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name: "zero_timeout",
			// Zero timeout value uses default.
			input:      "version=master&timeout=0",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "zero_unit_timeout",
			input:      "version=master&timeout=0s",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		{
			name:       "timeout",
			input:      "version=master&timeout=61",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    61 * time.Second,
		},
		{
			name:       "timeout_unit",
			input:      "version=master&timeout=1m1s",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    61 * time.Second,
		},
		{
			name:       "all",
			input:      "version=master&submodules=false&timeout=1m1s",
			ref:        "master",
			submodules: false,
			timeout:    61 * time.Second,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ref, timeout, submodules := parseQuery(tc.input)
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
