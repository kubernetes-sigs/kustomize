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
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ FileSystem = realFS{}

// realFS implements FileSystem using the local filesystem.
type realFS struct{}

// MakeRealFS makes an instance of realFS.
func MakeRealFS() FileSystem {
	return realFS{}
}

// Create delegates to os.Create.
func (realFS) Create(name string) (File, error) { return os.Create(name) }

// Mkdir delegates to os.Mkdir.
func (realFS) Mkdir(name string, perm os.FileMode) error { return os.Mkdir(name, perm) }

// Open delegates to os.Open.
func (realFS) Open(name string) (File, error) { return os.Open(name) }

// Stat delegates to os.Stat.
func (realFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

// ReadFile delegates to ioutil.ReadFile.
func (realFS) ReadFile(name string) ([]byte, error) { return ioutil.ReadFile(name) }

// ReadFiles use glob to find the matching files and then read content from all of them
func (realFS) ReadFiles(name string) (map[string][]byte, error) {
	files, err := filepath.Glob(name)
	if err != nil || len(files) == 0 {
		return nil, err
	}

	output := map[string][]byte{}
	for _, file := range files {
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if bytes != nil {
			output[file] = bytes
		}
	}
	return output, nil
}

// WriteFile delegates to ioutil.WriteFile with read/write permissions.
func (realFS) WriteFile(name string, c []byte) error {
	return ioutil.WriteFile(name, c, 0666)
}
