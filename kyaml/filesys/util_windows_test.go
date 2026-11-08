// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// On Windows, filepath treats both '\' and '/' as separators and paths
// may start with a volume name. PathSplit used to only recognize '\',
// and recursed forever (a stack overflow) on "/"-rooted and
// volume-rooted paths.
func TestPathSplitAndJoin_Windows(t *testing.T) {
	cases := map[string]struct {
		path     string
		expected []string
		joined   string
	}{
		"slash rooted":     {`/a/b`, []string{"", "a", "b"}, `\a\b`},
		"backslash rooted": {`\a\b`, []string{"", "a", "b"}, `\a\b`},
		"mixed separators": {`a/b\c`, []string{"a", "b", "c"}, `a\b\c`},
		"volume root only": {`C:\`, []string{`C:\`}, `C:\`},
		"volume rooted":    {`C:\a\b`, []string{`C:\`, "a", "b"}, `C:\a\b`},
		"volume relative":  {`C:a\b`, []string{"C:", "a", "b"}, `C:a\b`},
		"UNC":              {`\\host\share\a`, []string{`\\host\share\`, "a"}, `\\host\share\a`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			parts := PathSplit(tc.path)
			assert.Equal(t, tc.expected, parts)
			assert.Equal(t, tc.joined, PathJoin(parts))
		})
	}
}

// A volume root is the absolute part of the path, so position 0 is
// the first element after it, as with a "\"-rooted path.
func TestInsertPathPart_Windows(t *testing.T) {
	assert.Equal(t, `\x\a\b`, InsertPathPart(`/a/b`, 0, "x"))
	assert.Equal(t, `C:\x\a\b`, InsertPathPart(`C:\a\b`, 0, "x"))
	assert.Equal(t, `C:\a\b\x`, InsertPathPart(`C:\a\b`, 99, "x"))
}

func TestStripSeps_Windows(t *testing.T) {
	assert.Equal(t, `a`, StripLeadingSeps(`/\a`))
	assert.Equal(t, `a`, StripTrailingSeps(`a\/`))
}

// The in-memory file system has a single root; a volume, e.g. "C:", is a
// top-level entry under it. Real Windows paths, e.g. from os.UserHomeDir,
// used to make lookups recurse forever, and "/" did not resolve to the
// root. Paths reported by Walk must keep the volume, since tests mirror
// real directories into memory and check the walked paths on disk.
func TestFsInMemory_Windows(t *testing.T) {
	fSys := MakeFsInMemory()
	require.NoError(t, fSys.WriteFile(`/a/b.txt`, []byte("b")))
	require.NoError(t, fSys.WriteFile(`C:\c\d.txt`, []byte("d")))

	assert.True(t, fSys.IsDir("/"))
	assert.True(t, fSys.Exists(`\a\b.txt`))
	assert.True(t, fSys.IsDir(`C:\`))
	assert.True(t, fSys.Exists(`C:/c/d.txt`))
	assert.False(t, fSys.Exists(`C:\a\b.txt`), "volumes are distinct from the root")
	assert.False(t, fSys.Exists(`D:\c\d.txt`), "volumes are distinct from each other")
	assert.False(t, fSys.Exists(`C:\Users\someone\.config`))

	matches, err := fSys.Glob("C:/c/*.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{`C:\c\d.txt`}, matches, "a \"/\" pattern matches like filepath.Glob")

	var files []string
	require.NoError(t, fSys.Walk("/", func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}))
	assert.Equal(t, []string{`C:\c\d.txt`, `\a\b.txt`}, files)

	content, err := fSys.ReadFile(`C:\c\d.txt`)
	require.NoError(t, err)
	assert.Equal(t, "d", string(content))
}
