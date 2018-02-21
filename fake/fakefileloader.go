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

package fake

import "k8s.io/kubectl/pkg/loader"

type fakefileLoader struct {
	data []byte
	err  error
}

// NewFakeFileLoader returns Loader which always returns byte array and an error.
// Example: case no error: NewFakeFileLoader(yamlBytes, nil)
// Example: case return an error: NewFakeFileLoader(nil, errors.New("forced error"))
// Location parameter is unneeded because we always return data bytes or an error.
func NewFakeFileLoader(content []byte, e error) (loader.Loader, error) {
	return &fakefileLoader{data: content, err: e}, nil
}

func (l *fakefileLoader) Load() ([]byte, error) {
	return l.data, l.err
}
