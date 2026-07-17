// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package embeddedasset provides helpers that preserve the legacy generated
// asset API exposed by the OpenAPI packages.
package embeddedasset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// File describes one embedded asset.
type File struct {
	Name       string
	Data       []byte
	Compressed bool
	Size       int64
	Mode       os.FileMode
	ModTime    time.Time
}

// Asset returns a fresh copy of the uncompressed asset data.
func Asset(file File, name string) ([]byte, error) {
	if canonical(name) != file.Name {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	if !file.Compressed {
		return bytes.Clone(file.Data), nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(file.Data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}
	b, readErr := io.ReadAll(io.LimitReader(reader, file.Size+1))
	closeErr := reader.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read %q: %w", name, readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("read %q: %w", name, closeErr)
	}
	if int64(len(b)) != file.Size {
		return nil, fmt.Errorf("read %q: expected %d bytes, got %d", name, file.Size, len(b))
	}
	return b, nil
}

// MustAsset is like Asset but panics on error.
func MustAsset(file File, name string) []byte {
	b, err := Asset(file, name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}
	return b
}

// AssetInfo returns metadata for an embedded asset.
func AssetInfo(file File, name string) (os.FileInfo, error) {
	if canonical(name) != file.Name {
		return nil, fmt.Errorf("AssetInfo %s not found", name)
	}
	return fileInfo{file: file}, nil
}

// AssetDir returns the immediate children of a directory in the asset path.
func AssetDir(file File, name string) ([]string, error) {
	name = canonical(name)
	parts := strings.Split(file.Name, "/")
	if name == "" {
		return []string{parts[0]}, nil
	}
	dirParts := strings.Split(name, "/")
	if len(dirParts) >= len(parts) {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	for i := range dirParts {
		if dirParts[i] != parts[i] {
			return nil, fmt.Errorf("Asset %s not found", name)
		}
	}
	return []string{parts[len(dirParts)]}, nil
}

// RestoreAsset restores an asset under dir.
func RestoreAsset(file File, dir, name string) error {
	b, err := Asset(file, name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(file, name)
	if err != nil {
		return err
	}
	target := filePath(dir, name)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create asset directory: %w", err)
	}
	if err := os.WriteFile(target, b, info.Mode()); err != nil {
		return fmt.Errorf("write asset %q: %w", name, err)
	}
	if err := os.Chtimes(target, info.ModTime(), info.ModTime()); err != nil {
		return fmt.Errorf("restore timestamps for asset %q: %w", name, err)
	}
	return nil
}

// RestoreAssets restores an asset or directory recursively under dir.
func RestoreAssets(file File, dir, name string) error {
	children, err := AssetDir(file, name)
	if err != nil {
		return RestoreAsset(file, dir, name)
	}
	for _, child := range children {
		if err := RestoreAssets(file, dir, filepath.Join(name, child)); err != nil {
			return err
		}
	}
	return nil
}

type fileInfo struct {
	file File
}

func (info fileInfo) Name() string       { return info.file.Name }
func (info fileInfo) Size() int64        { return info.file.Size }
func (info fileInfo) Mode() os.FileMode  { return info.file.Mode }
func (info fileInfo) ModTime() time.Time { return info.file.ModTime }
func (info fileInfo) IsDir() bool        { return false }
func (info fileInfo) Sys() interface{}   { return nil }

func canonical(name string) string {
	return strings.ReplaceAll(name, "\\", "/")
}

func filePath(dir, name string) string {
	return filepath.Join(append([]string{dir}, strings.Split(canonical(name), "/")...)...)
}
