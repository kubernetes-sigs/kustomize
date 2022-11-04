// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer //nolint:testpackage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestUrlBase(t *testing.T) {
	tests := map[string]struct {
		url, base string
	}{
		"simple": {
			url:  "https://github.com/org/repo",
			base: "repo",
		},
		"trailing_slash": {
			url:  "github.com/org/repo//",
			base: "repo",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.base, urlBase(test.url))
		})
	}
}

// simpleJoin is filepath.Join() without side effects
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

func TestLocFilePathColon(t *testing.T) {
	req := require.New(t)

	// colon once used as unix file separator; also check IPv6 processing
	const url = "https://[2001:4860:4860::8888]/file.yaml"
	const host = "2001:4860:4860::8888"
	const file = "file.yaml"
	req.Equal(simpleJoin(t, LocalizeDir, host, file), locFilePath(url))

	fSys := filesys.MakeFsOnDisk()
	targetDir := simpleJoin(t, t.TempDir(), host)

	// make sure to create single directory to check that ':' not used as file separators
	req.NoError(fSys.Mkdir(targetDir))
	_, err := fSys.Create(simpleJoin(t, targetDir, file))
	req.NoError(err)

	// readable
	files, err := fSys.ReadDir(targetDir)
	req.NoError(err)
	req.Equal([]string{file}, files)
}

func TestLocFilePathSpecialChar(t *testing.T) {
	req := require.New(t)

	const wildcard = "*"
	req.Equal(simpleJoin(t, LocalizeDir, "host", wildcard), locFilePath("https://host/*"))

	fSys := filesys.MakeFsOnDisk()
	testDir := t.TempDir()
	req.NoError(fSys.Mkdir(simpleJoin(t, testDir, "a")))
	req.NoError(fSys.WriteFile(simpleJoin(t, testDir, "b"), []byte{}))
	// check wildcard name not being matched to existing files
	// and can be successfully created
	req.NoError(fSys.WriteFile(simpleJoin(t, testDir, wildcard), []byte("test")))

	content, err := fSys.ReadFile(simpleJoin(t, testDir, wildcard))
	req.NoError(err)
	req.Equal("test", string(content))
}

func TestLocFilePathSpecialFiles(t *testing.T) {
	for name, tFSys := range map[string]struct {
		urlPath           string
		pathDir, pathFile string
	}{
		"windows_reserved_name": {
			urlPath:  "/aux/file",
			pathDir:  "aux",
			pathFile: "file",
		},
		"hidden_files": {
			urlPath:  "/.../.file",
			pathDir:  "...",
			pathFile: ".file",
		},
	} {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			expectedPath := simpleJoin(t, LocalizeDir, "host", tFSys.pathDir, tFSys.pathFile)
			req.Equal(expectedPath, locFilePath("https://host"+tFSys.urlPath))

			fSys := filesys.MakeFsOnDisk()
			targetDir := simpleJoin(t, t.TempDir(), tFSys.pathDir)
			req.NoError(fSys.Mkdir(targetDir))
			req.NoError(fSys.WriteFile(simpleJoin(t, targetDir, tFSys.pathFile), []byte("test")))

			content, err := fSys.ReadFile(simpleJoin(t, targetDir, tFSys.pathFile))
			req.NoError(err)
			req.Equal([]byte("test"), content)
		})
	}
}

func TestLocRootPath(t *testing.T) {
	// kustomize does not support ports yet
	// UIs seem to constrain special characters in org, repo
	// path/to/root is actually path instead of url
	// git limits special characters in ref
	// TODO(annasong): fix skipped tests
	for name, tSamePath := range map[string]struct {
		skip       bool
		urlf, path string
	}{
		"ssh_non-github": {
			// RepoSpec bug
			skip: true,
			urlf: "ssh://git@gitlab.com/org/repo//%s?ref=value",
			path: simpleJoin(t, "gitlab.com", "org", "repo", "value"),
		},
		"rel_ssh_non-github": {
			// RepoSpec bug
			skip: true,
			urlf: "git@gitlab.com:org/repo//%s?ref=value",
			path: simpleJoin(t, "gitlab.com", "org", "repo", "value"),
		},
		"https_.git_suffix": {
			urlf: "https://gitlab.com/org/repo.git//%s?ref=value",
			path: simpleJoin(t, "gitlab.com", "org", "repo", "value"),
		},
		"gh_shorthand": {
			urlf: "gh:org/repo//%s?ref=value",
			path: simpleJoin(t, "github.com", "org", "repo", "value"),
		},
		"illegal_windows_dir": {
			urlf: "https://gitlab.com/org./repo..git//%s?ref=value",
			path: simpleJoin(t, "gitlab.com", "org.", "repo.", "value"),
		},
		"ref_has_slash": {
			urlf: "https://gitlab.com/org/repo//%s?ref=group/version/kind",
			path: simpleJoin(t, "gitlab.com", "org", "repo", "group", "version", "kind"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			if tSamePath.skip {
				t.Skip()
			}
			req := require.New(t)
			fSys := filesys.MakeFsOnDisk()

			url := fmt.Sprintf(tSamePath.urlf, "path/to/root")
			path := simpleJoin(t, LocalizeDir, tSamePath.path, "path", "to", "root")

			testDir := t.TempDir()
			repoDir := simpleJoin(t, testDir, "repo-random_hash")
			req.NoError(fSys.Mkdir(repoDir))
			rootDir := simpleJoin(t, repoDir, "path", "to", "root")
			req.NoError(fSys.MkdirAll(rootDir))

			actual := locRootPath(url, filesys.ConfirmedDir(repoDir), filesys.ConfirmedDir(rootDir))
			req.Equal(path, actual)

			req.NoError(fSys.MkdirAll(simpleJoin(t, testDir, path)))
		})
	}
}

func TestLocRootPathRepo(t *testing.T) {
	for name, test := range map[string]struct {
		skip      bool
		url, path string
	}{
		"simple": {
			url:  "https://github.com/org/repo?ref=value",
			path: simpleJoin(t, LocalizeDir, "github.com", "org", "repo", "value"),
		},
		"long_org_path": {
			url:  "https://github.com/parent-org/child-org/repo.git?ref=value",
			path: simpleJoin(t, LocalizeDir, "github.com", "parent-org", "child-org", "repo", "value"),
		},
		"ref_slash": {
			skip: true,
			url:  "https://github.com/org/repo?ref=group/version",
			path: simpleJoin(t, LocalizeDir, "github.com", "org", "repo", "group", "version"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.Skip()
			}
			testDir := t.TempDir()
			repoDir := filesys.ConfirmedDir(testDir)
			require.Equal(t, test.path, locRootPath(test.url, repoDir, repoDir))
		})
	}
}

func TestLocRootPathSymlinkPath(t *testing.T) {
	req := require.New(t)
	fSys := filesys.MakeFsOnDisk()

	const url = "https://github.com/org/repo//symlink?ref=value"
	repoDir := t.TempDir()
	rootDir := simpleJoin(t, repoDir, "actual-root")
	req.NoError(fSys.Mkdir(rootDir))
	req.NoError(os.Symlink(rootDir, simpleJoin(t, repoDir, "symlink")))

	expected := simpleJoin(t, LocalizeDir, "github.com", "org", "repo", "value", "actual-root")
	req.Equal(expected, locRootPath(url, filesys.ConfirmedDir(repoDir), filesys.ConfirmedDir(rootDir)))
}

func TestDefaultNewDirRepo(t *testing.T) {
	for name, test := range map[string]struct {
		skip     bool
		url, dst string
	}{
		"simple": {
			url: "https://github.com/org/repo?ref=value",
			dst: "localized-repo-value",
		},
		// TODO(annasong): Fix test
		// RepoSpec bug
		"slashed_ref": {
			skip: true,
			url:  "https://github.com/org/repo?ref=group/version",
			dst:  "localized-repo-group-version",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.Skip()
			}
			repoSpec, err := git.NewRepoSpecFromURL(test.url)
			require.NoError(t, err)
			require.Equal(t, test.dst, defaultNewDir(&fakeLoader{t.TempDir()}, repoSpec))
		})
	}
}

type fakeLoader struct {
	root string
}

func (fl *fakeLoader) Root() string {
	return fl.root
}
func (fl *fakeLoader) Repo() (string, bool) {
	return fl.root, true
}
func (fl *fakeLoader) Load(_ string) ([]byte, error) {
	return []byte{}, nil
}
func (fl *fakeLoader) New(path string) (ifc.Loader, error) {
	return &fakeLoader{path}, nil
}
func (fl *fakeLoader) Cleanup() error {
	return nil
}
