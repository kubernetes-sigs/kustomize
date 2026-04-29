// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepoSpecFromUrl_Permute(t *testing.T) {
	// Generate all many permutations of host, RepoPath, pathName, and ref.
	// Not all combinations make sense, but the parsing is very permissive and
	// we probably stil don't want to break backwards compatibility for things
	// that are unintentionally supported.
	var schemeAuthority = []struct {
		raw        string
		normalized string
	}{
		{"oci://ghcr.io/", "ghcr.io/"},
		{"oci://ghcr.io:7999/", "ghcr.io:7999/"},
		{"oci-layout://", "oci-layout://"},
	}
	var repoPaths = []string{
		"someOrg/someRepo",
		"kubernetes/website",
		"somepath",
	}
	var pathNames = []string{"README.md", "foo/krusty.txt", ""}
	var tagArgs = []string{
		"some_tag",
		"someTag",
		"latest",
		"v0.1.0",
		"",
	}
	var digestArgs = []string{
		"sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
		"",
	}

	makeURL := func(hostFmt, repoPath, path, tag, digest string) string {
		url := hostFmt + repoPath

		if tag != "" {
			url += tagSeparator + tag
		}
		if digest != "" {
			url += digestSeparator + digest
		}
		if len(path) > 0 {
			url += rootSeparator + path
		}

		return url
	}

	var i int
	for _, v := range schemeAuthority {
		hostRaw := v.raw
		hostSpec := v.normalized
		for _, repoPath := range repoPaths {
			for _, pathName := range pathNames {
				for _, tag := range tagArgs {
					for _, digest := range digestArgs {
						t.Run(fmt.Sprintf("t%d", i), func(t *testing.T) {
							uri := makeURL(hostRaw, repoPath, pathName, tag, digest)
							rs, err := NewRepoSpecFromURL(uri)
							require.NoErrorf(t, err, "unexpected error creating RepoSpec for uri %s", uri)
							assert.Equal(t, hostSpec, rs.Host, "unexpected host for uri %s", uri)
							assert.Equal(t, repoPath, rs.RepoPath, "unexpected RepoPath for uri %s", uri)
							assert.Equal(t, pathName, rs.KustRootPath, "unexpected KustRootPath for uri %s", uri)
							assert.Equal(t, tag, rs.Tag, "unexpected tag for uri %s", uri)
							assert.Equal(t, digest, rs.Digest, "unexpected digest for uri %s", uri)
						})
						i++
					}
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
		"relative path": {
			"../../tmp",
			"unsupported scheme",
		},
		"local path that looks somewhat like a github url": {
			"src/ghcr.com/org/repo/path",
			"unsupported scheme",
		},
		"bad_scheme": {
			"ocierr://ghcr.io/",
			"unsupported scheme",
		},
		"no_image_repo": {
			"oci://ghcr.io",
			"failed to parse repo path segment",
		},
		"tag_after_host": {
			"oci://host:tag",
			"failed to parse repo path segment",
		},
		"zero_length_tag": {
			"oci://host/repo:@asdfsdfsdf",
			"failed to parse tag segment",
		},
		"zero_length_digest": {
			"oci://host/repo:tag@",
			"failed to parse digest",
		},
		"path_exits_repo": {
			"oci://ghcr.io/org/repo//path/../../exits/repo",
			"root path exits repo",
		},
		"protocol with excessive slashes": {
			"oci-layout:////tmp/path/to/whatever",
			"failed to parse repo path segment",
		},
		"no_root_path": {
			"oci://ghcr.io/repo//",
			"failed to parse root path segment",
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
		name     string
		input    string
		repoSpec RepoSpec
		pullSpec string
		absPath  string
	}{
		{
			name:     "github container registry subdir",
			input:    "oci://ghcr.io/someorg/somerepo//somedir",
			pullSpec: "ghcr.io/someorg/somerepo",
			absPath:  notPulled.Join("somedir"),
			repoSpec: RepoSpec{
				Host:         "ghcr.io/",
				RepoPath:     "someorg/somerepo",
				KustRootPath: "somedir",
			},
		},
		{
			name:     "github container registry subdir tag",
			input:    "oci://ghcr.io/someorg/somerepo:someTag//somedir",
			pullSpec: "ghcr.io/someorg/somerepo:someTag",
			absPath:  notPulled.Join("somedir"),
			repoSpec: RepoSpec{
				Host:         "ghcr.io/",
				RepoPath:     "someorg/somerepo",
				KustRootPath: "somedir",
				Tag:          "someTag",
			},
		},
		{
			name:     "github container registry tag",
			input:    "oci://ghcr.io/someorg/somerepo:someTag",
			pullSpec: "ghcr.io/someorg/somerepo:someTag",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "ghcr.io/",
				RepoPath: "someorg/somerepo",
				Tag:      "someTag",
			},
		},
		{
			name:     "github container registry digest",
			input:    "oci://ghcr.io/someorg/somerepo@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			pullSpec: "ghcr.io/someorg/somerepo@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "ghcr.io/",
				RepoPath: "someorg/somerepo",
				Digest:   "sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
				Tag:      "",
			},
		},
		{
			name:     "github container registry tag and digest",
			input:    "oci://ghcr.io/someorg/somerepo:someTag@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			pullSpec: "ghcr.io/someorg/somerepo:someTag@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "ghcr.io/",
				RepoPath: "someorg/somerepo",
				Tag:      "someTag",
				Digest:   "sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			},
		},
		{
			name:     "oci layout with path and tag",
			input:    "oci-layout://a/b/c/someRepo:v2.5.4//somepath",
			pullSpec: "oci-layout://a/b/c/someRepo:v2.5.4",
			absPath:  notPulled.Join("somepath"),
			repoSpec: RepoSpec{
				Host:         "oci-layout://",
				RepoPath:     "a/b/c/someRepo",
				KustRootPath: "somepath",
				Tag:          "v2.5.4",
			},
		},
		{
			name:     "oci layout with two slashes, with kust root path and tag",
			input:    "oci-layout://a/b/c/someRepo:sometag//somepath",
			pullSpec: "oci-layout://a/b/c/someRepo:sometag",
			absPath:  notPulled.Join("somepath"),
			repoSpec: RepoSpec{
				Host:         "oci-layout://",
				RepoPath:     "a/b/c/someRepo",
				KustRootPath: "somepath",
				Tag:          "sometag",
			},
		},
		{
			name:     "oci layout with two slashes, with tag and no kust root path",
			input:    "oci-layout://a/b/c/someRepo:sometag",
			pullSpec: "oci-layout://a/b/c/someRepo:sometag",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "oci-layout://",
				RepoPath: "a/b/c/someRepo",
				Tag:      "sometag",
			},
		},
		{
			name:     "oci layout with three slashes, with tag and no kust root path",
			input:    "oci-layout:///a/b/c/someRepo:tag",
			pullSpec: "oci-layout:///a/b/c/someRepo:tag",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "oci-layout://",
				RepoPath: "/a/b/c/someRepo",
				Tag:      "tag",
			},
		},
		{
			name:     "oci layout with three slashes, no kust root path ortag",
			input:    "oci-layout:///a/b/c/someRepo",
			pullSpec: "oci-layout:///a/b/c/someRepo",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "oci-layout://",
				RepoPath: "/a/b/c/someRepo",
			},
		},
		{
			name:     "oci layout with three slashes, no repo or kust root path or tag",
			input:    "oci-layout:///",
			pullSpec: "oci-layout:///",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "oci-layout://",
				RepoPath: "/",
			},
		},
		{
			name:     "arbitrary host with double-slash path delimiter",
			input:    "oci://example.org/path/to/repo//examples/multibases/dev",
			pullSpec: "example.org/path/to/repo",
			absPath:  notPulled.Join("/examples/multibases/dev"),
			repoSpec: RepoSpec{
				Host:         "example.org/",
				RepoPath:     "path/to/repo",
				KustRootPath: "examples/multibases/dev",
			},
		},
		{
			name:     "non_parsable_path",
			input:    "oci://authority/org/repo/%-invalid-uri-so-not-parsable-by-net/url.Parse",
			pullSpec: "authority/org/repo/%-invalid-uri-so-not-parsable-by-net/url.Parse",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Host:     "authority/",
				RepoPath: "org/repo/%-invalid-uri-so-not-parsable-by-net/url.Parse",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rs, err := NewRepoSpecFromURL(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.pullSpec, rs.PullSpec(), "pullSpec mismatch")
			assert.Equal(t, tc.absPath, rs.AbsPath(), "absPath mismatch")
			// some values have defaults. Clear them here so test cases remain compact.
			// This means submodules and timeout cannot be tested here. That's fine since
			// they are tested in TestParseQuery.
			rs.raw = ""
			rs.Dir = ""
			rs.Timeout = 0
			assert.Equal(t, &tc.repoSpec, rs) //nolint:gosec
		})
	}
}

func TestNewRepoSpecFromURL_DefaultQueryParams(t *testing.T) {
	repoSpec, err := NewRepoSpecFromURL("oci://ghcr.com/org")
	require.NoError(t, err)
	require.Equal(t, defaultTimeout, repoSpec.Timeout)
}
