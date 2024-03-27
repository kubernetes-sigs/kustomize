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

func TestDefaultNewDirRepo(t *testing.T) {
	for name, test := range map[string]struct {
		url, dst string
	}{
		"simple": {
			url: "https://github.com/org/repo?ref=value",
			dst: "localized-repo-value",
		},
		"slashed_ref": {
			url: "https://github.com/org/repo?ref=group/version",
			dst: "localized-repo-group-version",
		},
	} {
		t.Run(name, func(t *testing.T) {
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
func (fl *fakeLoader) Repo() string {
	return fl.root
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
		"http-scheme": {
			url:  "http://host/path",
			path: simpleJoin(t, "host", "path"),
		},
		"extraneous_components": {
			url:  "http://userinfo@host:1234/path/file?query",
			path: simpleJoin(t, "host", "path", "file"),
		},
		"empty_path": {
			url:  "https://host",
			path: "host",
		},
		"empty_path_segment": {
			url:  "https://host//",
			path: "host",
		},
		"percent-encoded_path": {
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

	// The colon is special because it was once used as the unix file separator.
	const url = "https://[2001:4860:4860::8888]/file.yaml"
	const host = "2001:4860:4860::8888"
	const file = "file.yaml"
	req.Equal(simpleJoin(t, LocalizeDir, host, file), locFilePath(url))

	fSys := filesys.MakeFsOnDisk()
	targetDir := simpleJoin(t, t.TempDir(), host)

	// We check that we can create single directory, meaning ':' not used as file separator.
	req.NoError(fSys.Mkdir(targetDir))
	_, err := fSys.Create(simpleJoin(t, targetDir, file))
	req.NoError(err)

	// We check that the directory with such name is readable.
	files, err := fSys.ReadDir(targetDir)
	req.NoError(err)
	req.Equal([]string{file}, files)
}

func TestLocFilePath_SpecialChar(t *testing.T) {
	req := require.New(t)

	// The wild card character is one of the legal uri characters with more meaning
	// to the system, so we test it here.
	const wildcard = "*"
	req.Equal(simpleJoin(t, LocalizeDir, "host", wildcard), locFilePath("https://host/*"))

	fSys := filesys.MakeFsOnDisk()
	testDir := t.TempDir()
	req.NoError(fSys.Mkdir(simpleJoin(t, testDir, "a")))
	req.NoError(fSys.WriteFile(simpleJoin(t, testDir, "b"), []byte{}))

	// We check that we can create and read from wild card-named file.
	// We check that the file system is not matching it to existing file names.
	req.NoError(fSys.WriteFile(simpleJoin(t, testDir, wildcard), []byte("test")))
	content, err := fSys.ReadFile(simpleJoin(t, testDir, wildcard))
	req.NoError(err)
	req.Equal("test", string(content))
}

func TestLocFilePath_SpecialFiles(t *testing.T) {
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

func makeConfirmedDir(t *testing.T) (filesys.FileSystem, filesys.ConfirmedDir) {
	t.Helper()

	fSys := filesys.MakeFsOnDisk()
	testDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = fSys.RemoveAll(testDir.String())
	})

	return fSys, testDir
}

func TestLocRootPath_URLComponents(t *testing.T) {
	for name, test := range map[string]struct {
		urlf, path string
	}{
		"ssh": {
			urlf: "ssh://git@github.com/org/repo//%s?ref=value",
			path: simpleJoin(t, "github.com", "org", "repo", "value"),
		},
		"rel_ssh": {
			urlf: "git@github.com:org/repo//%s?ref=value",
			path: simpleJoin(t, "github.com", "org", "repo", "value"),
		},
		"https": {
			urlf: "https://gitlab.com/org/repo//%s?ref=value",
			path: simpleJoin(t, "gitlab.com", "org", "repo", "value"),
		},
		"file": {
			urlf: "file:///var/run/repo//%s?ref=value",
			path: simpleJoin(t, FileSchemeDir, "var", "run", "repo", "value"),
		},
		"IPv6": {
			urlf: "https://[2001:4860:4860::8888]/org/repo//%s?ref=value",
			path: simpleJoin(t, "2001:4860:4860::8888", "org", "repo", "value"),
		},
		"port": {
			urlf: "https://localhost.com:8080/org/repo//%s?ref=value",
			path: simpleJoin(t, "localhost.com", "org", "repo", "value"),
		},
		"no_org": {
			urlf: "https://github.com/repo//%s?ref=value",
			path: simpleJoin(t, "github.com", "repo", "value"),
		},
		".git_suffix": {
			urlf: "https://github.com/org1/org2/repo.git//%s?ref=value",
			path: simpleJoin(t, "github.com", "org1", "org2", "repo", "value"),
		},
		"dot-segments": {
			urlf: "https://github.com/./../org/../org/repo.git//%s?ref=value",
			path: simpleJoin(t, "github.com", "org", "repo", "value"),
		},
		"no_path_delimiter": {
			urlf: "https://github.com/org/repo/%s?ref=value",
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
			u := fmt.Sprintf(test.urlf, "path/to/root")
			path := simpleJoin(t, LocalizeDir, test.path, "path", "to", "root")

			fSys, testDir := makeConfirmedDir(t)
			repoDir := simpleJoin(t, testDir.String(), "repo_random-hash")
			require.NoError(t, fSys.Mkdir(repoDir))
			rootDir := simpleJoin(t, repoDir, "path", "to", "root")
			require.NoError(t, fSys.MkdirAll(rootDir))

			actual, err := locRootPath(u, repoDir, filesys.ConfirmedDir(rootDir), fSys)
			require.NoError(t, err)
			require.Equal(t, path, actual)

			require.NoError(t, fSys.MkdirAll(simpleJoin(t, testDir.String(), path)))
		})
	}
}

func TestLocRootPath_Repo(t *testing.T) {
	const url = "https://github.com/org/repo?ref=value"
	expected := simpleJoin(t, LocalizeDir, "github.com", "org", "repo", "value")

	fSys, testDir := makeConfirmedDir(t)
	actual, err := locRootPath(url, testDir.String(), testDir, fSys)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestLocRootPath_SymlinkPath(t *testing.T) {
	const url = "https://github.com/org/repo//symlink?ref=value"

	fSys, repoDir := makeConfirmedDir(t)
	rootDir := simpleJoin(t, repoDir.String(), "actual-root")
	require.NoError(t, fSys.Mkdir(rootDir))
	require.NoError(t, os.Symlink(rootDir, simpleJoin(t, repoDir.String(), "symlink")))

	expected := simpleJoin(t, LocalizeDir, "github.com", "org", "repo", "value", "actual-root")
	actual, err := locRootPath(url, repoDir.String(), filesys.ConfirmedDir(rootDir), fSys)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestCleanedRelativePath(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	require.NoError(t, fSys.MkdirAll("/root/test"))
	require.NoError(t, fSys.WriteFile("/root/test/file.yaml", []byte("")))
	require.NoError(t, fSys.WriteFile("/root/filetwo.yaml", []byte("")))

	// Absolute path is cleaned to relative path
	cleanedPath := cleanedRelativePath(fSys, "/root/", "/root/test/file.yaml")
	require.Equal(t, "test/file.yaml", cleanedPath)

	// Winding absolute path is cleaned to relative path
	cleanedPath = cleanedRelativePath(fSys, "/root/", "/root/test/../filetwo.yaml")
	require.Equal(t, "filetwo.yaml", cleanedPath)

	// Already clean relative path stays the same
	cleanedPath = cleanedRelativePath(fSys, "/root/", "test/file.yaml")
	require.Equal(t, "test/file.yaml", cleanedPath)

	// Winding relative path is cleaned
	cleanedPath = cleanedRelativePath(fSys, "/root/", "test/../filetwo.yaml")
	require.Equal(t, "filetwo.yaml", cleanedPath)
}
