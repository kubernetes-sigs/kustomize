// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import (
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
)

// RootedFS exposes a FileSystem subtree using the standard io/fs interfaces.
// Call Close when the RootedFS is no longer needed.
type RootedFS struct {
	fsys      fs.FS
	close     func() error
	closeOnce sync.Once
	closeErr  error
}

var (
	_ fs.FS         = (*RootedFS)(nil)
	_ fs.ReadFileFS = (*RootedFS)(nil)
	_ fs.ReadDirFS  = (*RootedFS)(nil)
	_ fs.StatFS     = (*RootedFS)(nil)
	_ fs.SubFS      = (*RootedFS)(nil)
)

// NewRootedFS returns an io/fs view of fSys rooted at root.
//
// When fSys is the on-disk filesystem returned by MakeFsOnDisk, the view is
// backed by os.Root so that operations cannot escape root through relative
// paths or symbolic links. Other FileSystem implementations are adapted to
// the same relative, slash-separated path convention used by io/fs.
func NewRootedFS(fSys FileSystem, root string) (*RootedFS, error) {
	confirmedRoot, err := ConfirmDir(fSys, root)
	if err != nil {
		return nil, err
	}

	if isOnDiskFileSystem(fSys) {
		osRoot, err := os.OpenRoot(confirmedRoot.String())
		if err != nil {
			return nil, err //nolint:wrapcheck // Preserve os.OpenRoot's PathError.
		}
		return &RootedFS{
			fsys:  osRoot.FS(),
			close: osRoot.Close,
		}, nil
	}

	return &RootedFS{
		fsys: &fileSystemFS{
			fsys: fSys,
			root: confirmedRoot,
		},
		close: func() error { return nil },
	}, nil
}

func isOnDiskFileSystem(fSys FileSystem) bool {
	switch typed := fSys.(type) {
	case fsOnDisk, *fsOnDisk:
		return true
	case FileSystemOrOnDisk:
		return typed.FileSystem == nil || isOnDiskFileSystem(typed.FileSystem)
	case *FileSystemOrOnDisk:
		return typed.FileSystem == nil || isOnDiskFileSystem(typed.FileSystem)
	default:
		return false
	}
}

// Open implements fs.FS.
func (r *RootedFS) Open(name string) (fs.File, error) {
	return r.fsys.Open(name) //nolint:wrapcheck // Preserve the io/fs PathError contract.
}

// ReadFile implements fs.ReadFileFS.
func (r *RootedFS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(r.fsys, name) //nolint:wrapcheck // Preserve the io/fs PathError contract.
}

// ReadDir implements fs.ReadDirFS.
func (r *RootedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(r.fsys, name) //nolint:wrapcheck // Preserve the io/fs PathError contract.
}

// Stat implements fs.StatFS.
func (r *RootedFS) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(r.fsys, name) //nolint:wrapcheck // Preserve the io/fs PathError contract.
}

// Sub implements fs.SubFS. The returned filesystem remains valid until the
// parent RootedFS is closed.
func (r *RootedFS) Sub(dir string) (fs.FS, error) {
	return fs.Sub(r.fsys, dir) //nolint:wrapcheck // Preserve the io/fs PathError contract.
}

// Close releases resources held by the rooted filesystem.
func (r *RootedFS) Close() error {
	r.closeOnce.Do(func() {
		r.closeErr = r.close()
	})
	return r.closeErr
}

type fileSystemFS struct {
	fsys FileSystem
	root ConfirmedDir
}

func (f *fileSystemFS) Open(name string) (fs.File, error) {
	mapped, err := f.mappedPath("open", name)
	if err != nil {
		return nil, err
	}
	file, err := f.fsys.Open(mapped)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, newPathError("open", name, err)
	}
	if info.IsDir() {
		if err := file.Close(); err != nil {
			return nil, newPathError("open", name, err)
		}
		return &fileSystemFile{
			fsys: f,
			name: name,
			info: normalizeFileInfo(info),
		}, nil
	}
	return &fileSystemFile{
		File: file,
		fsys: f,
		name: name,
	}, nil
}

func (f *fileSystemFS) ReadFile(name string) ([]byte, error) {
	mapped, err := f.mappedPath("readfile", name)
	if err != nil {
		return nil, err
	}
	content, err := f.fsys.ReadFile(mapped)
	if err != nil {
		return nil, newPathError("readfile", name, err)
	}
	return content, nil
}

func (f *fileSystemFS) ReadDir(name string) ([]fs.DirEntry, error) {
	mapped, err := f.mappedPath("readdir", name)
	if err != nil {
		return nil, err
	}
	names, err := f.fsys.ReadDir(mapped)
	if err != nil {
		return nil, newPathError("readdir", name, err)
	}
	sort.Strings(names)

	entries := make([]fs.DirEntry, 0, len(names))
	for _, childName := range names {
		child, err := f.fsys.Open(filepath.Join(mapped, childName))
		if err != nil {
			return nil, newPathError("readdir", path.Join(name, childName), err)
		}
		info, statErr := child.Stat()
		closeErr := child.Close()
		if statErr != nil {
			return nil, newPathError("readdir", path.Join(name, childName), statErr)
		}
		if closeErr != nil {
			return nil, newPathError("readdir", path.Join(name, childName), closeErr)
		}
		entries = append(entries, fs.FileInfoToDirEntry(normalizeFileInfo(info)))
	}
	return entries, nil
}

func (f *fileSystemFS) Stat(name string) (fs.FileInfo, error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, newPathError("stat", name, err)
	}
	return info, nil
}

func (f *fileSystemFS) mappedPath(op, name string) (string, error) {
	if !fs.ValidPath(name) {
		return "", newPathError(op, name, fs.ErrInvalid)
	}
	localName, err := filepath.Localize(name)
	if err != nil {
		return "", newPathError(op, name, fs.ErrInvalid)
	}
	return f.root.Join(localName), nil
}

func newPathError(op, name string, err error) error {
	return &fs.PathError{Op: op, Path: name, Err: err}
}

type fileSystemFile struct {
	File
	fsys *fileSystemFS
	name string
	info fs.FileInfo

	closed bool

	readDirMu     sync.Mutex
	dirEntries    []fs.DirEntry
	dirEntriesSet bool
	dirOffset     int
}

var _ fs.ReadDirFile = (*fileSystemFile)(nil)

func (f *fileSystemFile) Read(p []byte) (int, error) {
	if f.File != nil {
		return f.File.Read(p) //nolint:wrapcheck // Preserve io.Reader errors, including io.EOF.
	}
	if f.closed {
		return 0, newPathError("read", f.name, os.ErrClosed)
	}
	return 0, newPathError("read", f.name, syscall.EISDIR)
}

func (f *fileSystemFile) Close() error {
	if f.File != nil {
		return f.File.Close() //nolint:wrapcheck // Preserve the adapted file's close error.
	}
	f.closed = true
	return nil
}

func (f *fileSystemFile) Stat() (fs.FileInfo, error) {
	if f.info != nil {
		return f.info, nil
	}
	info, err := f.File.Stat()
	if err != nil {
		return nil, newPathError("stat", f.name, err)
	}
	return normalizeFileInfo(info), nil
}

// ReadDir implements fs.ReadDirFile.
func (f *fileSystemFile) ReadDir(n int) ([]fs.DirEntry, error) {
	f.readDirMu.Lock()
	defer f.readDirMu.Unlock()

	if !f.dirEntriesSet {
		entries, err := f.fsys.ReadDir(f.name)
		if err != nil {
			return nil, err
		}
		f.dirEntries = entries
		f.dirEntriesSet = true
	}

	if n <= 0 {
		entries := f.dirEntries[f.dirOffset:]
		f.dirOffset = len(f.dirEntries)
		return entries, nil
	}
	if f.dirOffset >= len(f.dirEntries) {
		return nil, io.EOF
	}

	end := min(f.dirOffset+n, len(f.dirEntries))
	entries := f.dirEntries[f.dirOffset:end]
	f.dirOffset = end
	return entries, nil
}

type normalizedFileInfo struct {
	fs.FileInfo
}

func (i normalizedFileInfo) Mode() fs.FileMode {
	mode := i.FileInfo.Mode()
	if i.IsDir() {
		mode |= fs.ModeDir
	}
	return mode
}

func normalizeFileInfo(info fs.FileInfo) fs.FileInfo {
	if info.IsDir() && info.Mode()&fs.ModeDir == 0 {
		return normalizedFileInfo{FileInfo: info}
	}
	return info
}
