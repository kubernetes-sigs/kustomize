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

package loader

import (
	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

type fileLoader struct {
	fs fs.FileSystem

	path string
}

func NewFileLoader(fs fs.FileSystem, path string) (Loader, error) {
	return &fileLoader{fs: fs, path: path}, nil
}

func (l *fileLoader) Load() ([]byte, error) {
	return l.fs.ReadFile(l.path)
}
