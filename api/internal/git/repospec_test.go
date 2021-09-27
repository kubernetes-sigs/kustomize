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
)

var orgRepos = []string{"someOrg/someRepo", "kubernetes/website"}

var pathNames = []string{"README.md", "foo/krusty.txt", ""}

var hrefArgs = []string{"someBranch", "master", "v0.1.0", ""}

var hostNamesRawAndNormalized = [][]string{
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

func makeUrl(hostFmt, orgRepo, path, href string) string {
	if len(path) > 0 {
		orgRepo = filepath.Join(orgRepo, path)
	}
	url := hostFmt + orgRepo
	if href != "" {
		url += refQuery + href
	}
	return url
}

func TestNewRepoSpecFromUrl(t *testing.T) {
	var bad [][]string
	for _, tuple := range hostNamesRawAndNormalized {
		hostRaw := tuple[0]
		hostSpec := tuple[1]
		for _, orgRepo := range orgRepos {
			for _, pathName := range pathNames {
				for _, hrefArg := range hrefArgs {
					uri := makeUrl(hostRaw, orgRepo, pathName, hrefArg)
					rs, err := NewRepoSpecFromUrl(uri)
					if err != nil {
						t.Errorf("problem %v", err)
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

var badData = [][]string{
	{"/tmp", "uri looks like abs path"},
	{"iauhsdiuashduas", "url lacks orgRepo"},
	{"htxxxtp://github.com/", "url lacks host"},
	{"ssh://git.example.com", "url lacks orgRepo"},
	{"git::___", "url lacks orgRepo"},
}

func TestNewRepoSpecFromUrlErrors(t *testing.T) {
	for _, tuple := range badData {
		_, err := NewRepoSpecFromUrl(tuple[0])
		if err == nil {
			t.Error("expected error")
		}
		if !strings.Contains(err.Error(), tuple[1]) {
			t.Errorf("unexpected error: %s", err)
		}
	}
}

func TestNewRepoSpecFromUrl_CloneSpecs(t *testing.T) {
	testcases := map[string]struct {
		input     string
		cloneSpec string
		absPath   string
		ref       string
	}{
		"t1": {
			input:     "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somedir",
			cloneSpec: "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			absPath:   notCloned.Join("somedir"),
			ref:       "",
		},
		"t2": {
			input:     "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somedir?ref=testbranch",
			cloneSpec: "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			absPath:   notCloned.Join("somedir"),
			ref:       "testbranch",
		},
		"t3": {
			input:     "https://fabrikops2.visualstudio.com/someorg/somerepo?ref=master",
			cloneSpec: "https://fabrikops2.visualstudio.com/someorg/somerepo",
			absPath:   notCloned.String(),
			ref:       "master",
		},
		"t4": {
			input:     "http://github.com/someorg/somerepo/somedir",
			cloneSpec: "https://github.com/someorg/somerepo.git",
			absPath:   notCloned.Join("somedir"),
			ref:       "",
		},
		"t5": {
			input:     "git@github.com:someorg/somerepo/somedir",
			cloneSpec: "git@github.com:someorg/somerepo.git",
			absPath:   notCloned.Join("somedir"),
			ref:       "",
		},
		"t6": {
			input:     "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git?ref=v0.1.0",
			cloneSpec: "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git",
			absPath:   notCloned.String(),
			ref:       "v0.1.0",
		},
		"t7": {
			input:     "git@bitbucket.org:company/project.git//path?ref=branch",
			cloneSpec: "git@bitbucket.org:company/project.git",
			absPath:   notCloned.Join("path"),
			ref:       "branch",
		},
		"t8": {
			input:     "https://itfs.mycompany.com/collection/project/_git/somerepos",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.String(),
			ref:       "",
		},
		"t9": {
			input:     "https://itfs.mycompany.com/collection/project/_git/somerepos?version=v1.0.0",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.String(),
			ref:       "v1.0.0",
		},
		"t10": {
			input:     "https://itfs.mycompany.com/collection/project/_git/somerepos/somedir?version=v1.0.0",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.Join("somedir"),
			ref:       "v1.0.0",
		},
		"t11": {
			input:     "git::https://itfs.mycompany.com/collection/project/_git/somerepos",
			cloneSpec: "https://itfs.mycompany.com/collection/project/_git/somerepos",
			absPath:   notCloned.String(),
			ref:       "",
		},
		"t12": {
			input:     "https://bitbucket.example.com/scm/project/repository.git",
			cloneSpec: "https://bitbucket.example.com/scm/project/repository.git",
			absPath:   notCloned.String(),
			ref:       "",
		},
	}
	for tn, tc := range testcases {
		t.Run(tn, func(t *testing.T) {
			rs, err := NewRepoSpecFromUrl(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.cloneSpec, rs.CloneSpec(), "cloneSpec mismatch")
			assert.Equal(t, tc.absPath, rs.AbsPath(), "absPath mismatch")
			assert.Equal(t, tc.ref, rs.Ref, "ref mismatch")
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
	testcases := map[string]struct {
		input      string
		path       string
		ref        string
		submodules bool
		timeout    time.Duration
	}{
		"t1": {
			// All empty.
			input:      "somerepos",
			path:       "somerepos",
			ref:        "",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t2": {
			input:      "somerepos?ref=v1.0.0",
			path:       "somerepos",
			ref:        "v1.0.0",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t3": {
			input:      "somerepos?version=master",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t4": {
			// A ref value takes precedence over a version value.
			input:      "somerepos?version=master&ref=v1.0.0",
			path:       "somerepos",
			ref:        "v1.0.0",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t5": {
			// Empty submodules value uses default.
			input:      "somerepos?version=master&submodules=",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t6": {
			// Malformed submodules value uses default.
			input:      "somerepos?version=master&submodules=maybe",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t7": {
			input:      "somerepos?version=master&submodules=true",
			path:       "somerepos",
			ref:        "master",
			submodules: true,
			timeout:    defaultTimeout,
		},
		"t8": {
			input:      "somerepos?version=master&submodules=false",
			path:       "somerepos",
			ref:        "master",
			submodules: false,
			timeout:    defaultTimeout,
		},
		"t9": {
			// Empty timeout value uses default.
			input:      "somerepos?version=master&timeout=",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t10": {
			// Malformed timeout value uses default.
			input:      "somerepos?version=master&timeout=jiffy",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t11": {
			// Zero timeout value uses default.
			input:      "somerepos?version=master&timeout=0",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t12": {
			input:      "somerepos?version=master&timeout=0s",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    defaultTimeout,
		},
		"t13": {
			input:      "somerepos?version=master&timeout=61",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    61 * time.Second,
		},
		"t14": {
			input:      "somerepos?version=master&timeout=1m1s",
			path:       "somerepos",
			ref:        "master",
			submodules: defaultSubmodules,
			timeout:    61 * time.Second,
		},
		"t15": {
			input:      "somerepos?version=master&submodules=false&timeout=1m1s",
			path:       "somerepos",
			ref:        "master",
			submodules: false,
			timeout:    61 * time.Second,
		},
	}
	for tn, tc := range testcases {
		t.Run(tn, func(t *testing.T) {
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
