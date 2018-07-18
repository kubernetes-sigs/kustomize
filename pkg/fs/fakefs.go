/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fs

import (
	"fmt"
)

var _ FileSystem = &FakeFS{}

// FakeFS implements FileSystem using a fake in-memory filesystem.
type FakeFS struct {
	m map[string]*FakeFile
}

// MakeFakeFS returns an instance of FakeFS with no files in it.
func MakeFakeFS() *FakeFS {
	return &FakeFS{m: map[string]*FakeFile{}}
}

// Create assures a fake file appears in the in-memory file system.
func (fs *FakeFS) Create(name string) (File, error) {
	f := &FakeFile{}
	f.open = true
	fs.m[name] = f
	return fs.m[name], nil
}

// Mkdir assures a fake directory appears in the in-memory file system.
func (fs *FakeFS) Mkdir(name string) error {
	fs.m[name] = makeDir(name)
	return nil
}

// Open returns a fake file in the open state.
func (fs *FakeFS) Open(name string) (File, error) {
	if _, found := fs.m[name]; !found {
		return nil, fmt.Errorf("file %q cannot be opened", name)
	}
	return fs.m[name], nil
}

// Exists returns true if file is known.
func (fs *FakeFS) Exists(name string) bool {
	_, found := fs.m[name]
	return found
}

// IsDir returns true if the file exists and is a directory.
func (fs *FakeFS) IsDir(name string) bool {
	f, found := fs.m[name]
	if !found {
		return false
	}
	return f.dir
}

// ReadFile always returns an empty bytes and error depending on content of m.
func (fs *FakeFS) ReadFile(name string) ([]byte, error) {
	if ff, found := fs.m[name]; found {
		return ff.content, nil
	}
	return nil, fmt.Errorf("cannot read file %q", name)
}

// ReadFiles looks through all files in the fake filesystem
// and find the matching files and then read content from all of them
func (fs *FakeFS) ReadFiles(name string) (map[string][]byte, error) {
	result := map[string][]byte{}
	for p, f := range fs.m {
		if fs.pathMatch(p, name) {
			result[p] = f.content
		}
	}
	return result, nil
}

// WriteFile always succeeds and does nothing.
func (fs *FakeFS) WriteFile(name string, c []byte) error {
	ff := &FakeFile{}
	ff.Write(c)
	fs.m[name] = ff
	return nil
}

func (fs *FakeFS) pathMatch(path, pattern string) bool {
	if path == pattern {
		return true
	}
	return false
}
