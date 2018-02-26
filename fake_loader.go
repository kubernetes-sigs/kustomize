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

// FakeLoader implements Loader interface.
type FakeLoader struct {
	Content string
}

func (f FakeLoader) Root() string {
	return "/fake/root"
}

func (f FakeLoader) New(root string) (Loader, error) {
	return f, nil
}

func (f FakeLoader) Load(location string) ([]byte, error) {
	return []byte(f.Content), nil
}

var _ Loader = FakeLoader{}
