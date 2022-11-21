// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer //nolint:testpackage

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlBase(t *testing.T) {
	require.Equal(t, "repo", urlBase("https://github.com/org/repo"))
}

func TestUrlBaseTrailingSlash(t *testing.T) {
	require.Equal(t, "repo", urlBase("github.com/org/repo//"))
}

// simpleJoin is filepath.Join() without the side effects of filepath.Clean()
func simpleJoin(t *testing.T, elems ...string) string {
	t.Helper()

	return strings.Join(elems, string(filepath.Separator))
}

func TestLocFilePath(t *testing.T) {
	for name, tUnit := range map[string]struct {
		url, path string
	}{
		"official": {
			url:  "https://raw.githubusercontent.com/org/repo/ref/path/to/file.yaml",
			path: simpleJoin(t, "raw.githubusercontent.com", "org", "repo", "ref", "path", "to", "file.yaml"),
		},
		"empty_path": {
			url:  "https://host",
			path: "host",
		},
		"empty_path_segment": {
			url:  "https://host//",
			path: "host",
		},
		"extraneous_components": {
			url:  "http://userinfo@host:1234/path/file?query",
			path: simpleJoin(t, "host", "path", "file"),
		},
		"percent-encoding": {
			url:  "https://host/file%2Eyaml",
			path: simpleJoin(t, "host", "file%2Eyaml"),
		},
		"dot-segments": {
			url:  "https://host/path/blah/../to/foo/bar/../../file/./",
			path: simpleJoin(t, "host", "path", "to", "file"),
		},
		"extraneous_dot-segments": {
			url:  "https://host/foo/bar/baz/../../../../file",
			path: simpleJoin(t, "host", "file"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, simpleJoin(t, LocalizeDir, tUnit.path), locFilePath(tUnit.url))
		})
	}
}
