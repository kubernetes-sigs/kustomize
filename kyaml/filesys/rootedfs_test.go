// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestRootedFS(t *testing.T) {
	for name, setup := range map[string]func(*testing.T) (FileSystem, string){
		"memory": func(t *testing.T) (FileSystem, string) {
			t.Helper()
			fSys := MakeFsInMemory()
			root := filepath.Join(Separator, "root")
			require.NoError(t, fSys.MkdirAll(root))
			return fSys, root
		},
		"disk": func(t *testing.T) (FileSystem, string) {
			t.Helper()
			return MakeFsOnDisk(), t.TempDir()
		},
	} {
		t.Run(name, func(t *testing.T) {
			fSys, root := setup(t)
			require.NoError(t, fSys.MkdirAll(filepath.Join(root, "nested")))
			require.NoError(t, fSys.WriteFile(filepath.Join(root, "nested", "file.yaml"), []byte("content")))
			if isOnDiskFileSystem(fSys) {
				forceMetadataUpdateOnWindows(t, root)
			}

			rooted, err := NewRootedFS(fSys, root)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, rooted.Close())
			})

			require.NoError(t, fstest.TestFS(rooted, "nested/file.yaml"))
			content, err := fs.ReadFile(rooted, "nested/file.yaml")
			require.NoError(t, err)
			require.Equal(t, "content", string(content))

			entries, err := fs.ReadDir(rooted, ".")
			require.NoError(t, err)
			require.Len(t, entries, 1)
			require.Equal(t, "nested", entries[0].Name())
			require.True(t, entries[0].IsDir())

			info, err := fs.Stat(rooted, "nested/file.yaml")
			require.NoError(t, err)
			require.Equal(t, int64(len("content")), info.Size())

			sub, err := fs.Sub(rooted, "nested")
			require.NoError(t, err)
			content, err = fs.ReadFile(sub, "file.yaml")
			require.NoError(t, err)
			require.Equal(t, "content", string(content))

			_, err = rooted.Open("../file.yaml")
			require.ErrorIs(t, err, fs.ErrInvalid)
			require.NoError(t, rooted.Close())
		})
	}
}

func forceMetadataUpdateOnWindows(t *testing.T, root string) {
	t.Helper()
	if runtime.GOOS != "windows" {
		return
	}

	// Directory enumeration and Stat can temporarily observe different MFT metadata.
	// Rewriting the current timestamp synchronizes them before fstest.TestFS compares them.
	// See https://go.dev/issue/42637.
	require.NoError(t, filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk %q: %w", path, err)
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get directory entry info for %q: %w", path, err)
		}
		statInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %w", path, err)
		}
		if entryInfo.ModTime() == statInfo.ModTime() {
			return nil
		}
		if err := os.Chtimes(path, statInfo.ModTime(), statInfo.ModTime()); err != nil {
			return fmt.Errorf("failed to synchronize metadata for %q: %w", path, err)
		}
		return nil
	}))
}

func TestRootedFSOnDiskRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "inside.yaml")
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	require.NoError(t, os.WriteFile(inside, []byte("inside"), 0o600))
	require.NoError(t, os.WriteFile(outside, []byte("outside"), 0o600))

	rooted, err := NewRootedFS(MakeFsOnDisk(), root)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, rooted.Close())
	})

	insideLink := filepath.Join(root, "inside-link.yaml")
	if err := os.Symlink(filepath.Base(inside), insideLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	content, err := fs.ReadFile(rooted, filepath.Base(insideLink))
	require.NoError(t, err)
	require.Equal(t, "inside", string(content))

	escapeLink := filepath.Join(root, "escape-link.yaml")
	require.NoError(t, os.Symlink(outside, escapeLink))
	_, err = fs.ReadFile(rooted, filepath.Base(escapeLink))
	require.Error(t, err)
}

func TestNewRootedFSRequiresDirectory(t *testing.T) {
	fSys := MakeFsInMemory()
	require.NoError(t, fSys.WriteFile("file.yaml", []byte("content")))

	_, err := NewRootedFS(fSys, "file.yaml")
	require.Error(t, err)
}
