// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package embeddedasset

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testAssetPath = "root/nested/asset.txt"

var testModTime = time.Unix(1700000000, 0) //nolint:gochecknoglobals

func TestAsset(t *testing.T) {
	contents := []byte("asset contents")
	for name, compressed := range map[string]bool{
		"raw":        false,
		"compressed": true,
	} {
		t.Run(name, func(t *testing.T) {
			data := contents
			if compressed {
				data = gzipData(t, contents)
			}
			file := newTestFile(data, int64(len(contents)), compressed)

			got, err := Asset(file, testAssetPath)
			require.NoError(t, err)
			require.Equal(t, contents, got)

			got[0] ^= 0xff
			again, err := Asset(file, "root\\nested\\asset.txt")
			require.NoError(t, err)
			require.Equal(t, contents, again, "Asset must return a fresh byte slice")
		})
	}

	_, err := Asset(newTestFile(nil, 0, false), "missing")
	require.ErrorContains(t, err, "Asset missing not found")
}

func TestAssetRejectsInvalidCompressedData(t *testing.T) {
	t.Run("invalid gzip", func(t *testing.T) {
		file := newTestFile([]byte("not gzip"), 1, true)
		_, err := Asset(file, testAssetPath)
		require.ErrorContains(t, err, "read \"root/nested/asset.txt\"")
	})

	t.Run("invalid gzip body", func(t *testing.T) {
		data := gzipData(t, []byte("asset contents"))
		data[len(data)-1] ^= 0xff
		file := newTestFile(data, int64(len("asset contents")), true)
		_, err := Asset(file, testAssetPath)
		require.ErrorContains(t, err, "read \"root/nested/asset.txt\"")
	})

	t.Run("declared size too small", func(t *testing.T) {
		contents := []byte("asset contents")
		file := newTestFile(gzipData(t, contents), int64(len(contents)-1), true)
		_, err := Asset(file, testAssetPath)
		require.ErrorContains(t, err, "expected 13 bytes, got 14")
	})

	t.Run("declared size too large", func(t *testing.T) {
		contents := []byte("asset contents")
		file := newTestFile(gzipData(t, contents), int64(len(contents)+1), true)
		_, err := Asset(file, testAssetPath)
		require.ErrorContains(t, err, "expected 15 bytes, got 14")
	})
}

func TestMustAsset(t *testing.T) {
	contents := []byte("asset contents")
	file := newTestFile(contents, int64(len(contents)), false)
	require.Equal(t, contents, MustAsset(file, testAssetPath))
	require.Panics(t, func() {
		MustAsset(file, "missing")
	})
}

func TestAssetInfo(t *testing.T) {
	contents := []byte("asset contents")
	file := newTestFile(contents, int64(len(contents)), false)

	info, err := AssetInfo(file, "root\\nested\\asset.txt")
	require.NoError(t, err)
	require.Equal(t, "asset.txt", info.Name())
	require.Equal(t, int64(len(contents)), info.Size())
	require.Equal(t, os.FileMode(0o640), info.Mode())
	require.Equal(t, testModTime, info.ModTime())
	require.False(t, info.IsDir())
	require.Nil(t, info.Sys())

	_, err = AssetInfo(file, "missing")
	require.ErrorContains(t, err, "AssetInfo missing not found")
}

func TestAssetDir(t *testing.T) {
	file := newTestFile(nil, 0, false)
	tests := []struct {
		name     string
		path     string
		expected []string
		wantErr  bool
	}{
		{name: "root", path: "", expected: []string{"root"}},
		{name: "first level", path: "root", expected: []string{"nested"}},
		{name: "nested", path: "root/nested", expected: []string{"asset.txt"}},
		{name: "backslashes", path: "root\\nested", expected: []string{"asset.txt"}},
		{name: "file", path: testAssetPath, wantErr: true},
		{name: "missing", path: "missing", wantErr: true},
		{name: "trailing slash", path: "root/nested/", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			children, err := AssetDir(file, test.path)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, children)
		})
	}
}

func TestRestoreAsset(t *testing.T) {
	contents := []byte("asset contents")
	file := newTestFile(contents, int64(len(contents)), false)
	dir := t.TempDir()

	require.NoError(t, RestoreAsset(file, dir, "root\\nested\\asset.txt"))
	restored, err := os.ReadFile(filepath.Join(dir, "root", "nested", "asset.txt"))
	require.NoError(t, err)
	require.Equal(t, contents, restored)

	require.Error(t, RestoreAsset(file, dir, "missing"))
}

func TestRestoreAssets(t *testing.T) {
	contents := []byte("asset contents")
	file := newTestFile(contents, int64(len(contents)), false)

	for _, name := range []string{"", "root", testAssetPath} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, RestoreAssets(file, dir, name))
			restored, err := os.ReadFile(filepath.Join(dir, "root", "nested", "asset.txt"))
			require.NoError(t, err)
			require.Equal(t, contents, restored)
		})
	}

	require.Error(t, RestoreAssets(file, t.TempDir(), "missing"))
}

func TestCanonical(t *testing.T) {
	tests := map[string]string{
		"":                        "",
		"root/nested/asset.txt":   "root/nested/asset.txt",
		"root\\nested\\asset.txt": "root/nested/asset.txt",
		"root/nested\\asset.txt":  "root/nested/asset.txt",
	}
	for input, expected := range tests {
		require.Equal(t, expected, canonical(input))
	}
}

func TestFilePath(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "root", "nested", "asset.txt")
	require.Equal(t, expected, filePath(dir, "root/nested/asset.txt"))
	require.Equal(t, expected, filePath(dir, "root\\nested\\asset.txt"))
}

func newTestFile(data []byte, size int64, compressed bool) File {
	return File{
		Path:       testAssetPath,
		Data:       data,
		Compressed: compressed,
		Size:       size,
		Mode:       0o640,
		ModTime:    testModTime,
	}
}

func gzipData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return buffer.Bytes()
}
