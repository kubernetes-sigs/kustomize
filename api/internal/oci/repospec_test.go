// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
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
		remote     bool
	}{
		{"oci://ghcr.io/", "ghcr.io", true},
		{"oci://ghcr.io:7999/", "ghcr.io:7999", true},
	}
	var repoPaths = []string{
		"someorg/somerepo",
		"kubernetes/website",
		"somepath",
	}
	var pathNames = []string{
		"README.md",
		"foo/krusty.txt",
		"",
	}
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
		if path != "" {
			url += rootSeparator + path
		}

		return url
	}

	for _, v := range schemeAuthority {
		hostRaw := v.raw
		hostNormalized := v.normalized
		for _, repoPath := range repoPaths {
			for _, pathName := range pathNames {
				for _, tag := range tagArgs {
					for _, digest := range digestArgs {
						identifier := tag
						if digest != "" {
							identifier = digest
						}
						if identifier == "" {
							identifier = name.DefaultTag
						}

						t.Run(fmt.Sprintf("%s/%s/%s|%s|%s", url.PathEscape(hostRaw), url.PathEscape(repoPath), url.PathEscape(pathName), tag, digest), func(t *testing.T) {
							uri := makeURL(hostRaw, repoPath, pathName, tag, digest)
							rs, err := NewRepoSpecFromURL(uri)

							require.NoErrorf(t, err, "unexpected error creating RepoSpec for uri %s", uri)
							if v.remote {
								assert.Equal(t, hostNormalized, rs.Reference.Context().RegistryStr())
								assert.Equal(t, repoPath, rs.Reference.Context().RepositoryStr())
							} else {
								assert.Empty(t, rs.Reference.Context().RegistryStr())
								assert.Equal(t, hostNormalized+repoPath, rs.Reference.Context().RepositoryStr())
							}
							assert.Equal(t, identifier, rs.Reference.Identifier(), "unexpected identifier for uri %s", uri)
							assert.Equal(t, pathName, rs.KustRootPath, "unexpected KustRootPath for uri %s", uri)
						})
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
			"unsupported scheme",
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
			"invalid reference: missing registry or repository",
		},
		"tag_after_host": {
			"oci://host:tag",
			"invalid reference: missing registry or repository",
		},
		"zero_length_tag": {
			"oci://host/repo:@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			"could not parse reference: host/repo:@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
		},
		"zero_length_digest": {
			"oci://host/repo:tag@",
			"could not parse reference: ",
		},
		"zero_length_tag_and_digest": {
			"oci://host/repo:@",
			"could not parse reference: ",
		},
		"path_exits_repo": {
			"oci://ghcr.io/org/repo//path/../../exits/repo",
			"root path exits repo",
		},
		"protocol with excessive slashes": {
			"oci:////tmp/path/to/whatever",
			"could not parse reference: ",
		},
		"no_root_path": {
			"oci://ghcr.io/repo//",
			"failed to parse root path segment",
		},
		"invalid_url": {
			"oci://authority/org/repo/%-invalid-uri-so-not-parsable-by-net/url.Parse",
			"could not parse reference: authority/org/repo/%-invalid-uri-so-not-parsable-by-net/url.Parse",
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
				Reference:    name.MustParseReference("ghcr.io/someorg/somerepo"),
				KustRootPath: "somedir",
			},
		},
		{
			name:     "github container registry subdir tag",
			input:    "oci://ghcr.io/someorg/somerepo:someTag//somedir",
			pullSpec: "ghcr.io/someorg/somerepo:someTag",
			absPath:  notPulled.Join("somedir"),
			repoSpec: RepoSpec{
				Reference:    name.MustParseReference("ghcr.io/someorg/somerepo:someTag"),
				KustRootPath: "somedir",
			},
		},
		{
			name:     "github container registry tag",
			input:    "oci://ghcr.io/someorg/somerepo:someTag",
			pullSpec: "ghcr.io/someorg/somerepo:someTag",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Reference: name.MustParseReference("ghcr.io/someorg/somerepo:someTag"),
			},
		},
		{
			name:     "github container registry digest",
			input:    "oci://ghcr.io/someorg/somerepo@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			pullSpec: "ghcr.io/someorg/somerepo@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Reference: name.MustParseReference("ghcr.io/someorg/somerepo@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a"),
			},
		},
		{
			name:     "github container registry tag and digest",
			input:    "oci://ghcr.io/someorg/somerepo:someTag@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			pullSpec: "ghcr.io/someorg/somerepo:someTag@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a",
			absPath:  notPulled.String(),
			repoSpec: RepoSpec{
				Reference: name.MustParseReference("ghcr.io/someorg/somerepo:someTag@sha256:94a00394bc5a8ef503fb59db0a7d0ae9e1110866e8aee8ba40cd864cea69ea1a"),
			},
		},
		{
			name:     "arbitrary host with double-slash path delimiter",
			input:    "oci://example.org/path/to/repo//examples/multibases/dev",
			pullSpec: "example.org/path/to/repo",
			absPath:  notPulled.Join("/examples/multibases/dev"),
			repoSpec: RepoSpec{
				Reference:    name.MustParseReference("example.org/path/to/repo"),
				KustRootPath: "examples/multibases/dev",
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
	repoSpec, err := NewRepoSpecFromURL("oci://ghcr.com/org:latest")
	require.NoError(t, err)
	require.Equal(t, defaultTimeout, repoSpec.Timeout)
}
