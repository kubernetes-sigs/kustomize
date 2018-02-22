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
	"reflect"
	"testing"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

func TestFileLoaderHappyPath(t *testing.T) {
	fakefs := fs.MakeFakeFS()
	location := "foo"
	content := []byte("bar")
	fakefs.WriteFile(location, content)
	l, err := newFileLoader(fakefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := l.load(location)
	if err != nil {
		t.Fatalf("unexpected error in Load: %v", err)
	}
	if !reflect.DeepEqual(b, content) {
		t.Fatalf("expected %s, but got %s", content, b)
	}
}

func TestFileLoaderFileNotFound(t *testing.T) {
	fakefs := fs.MakeFakeFS()
	l, err := newFileLoader(fakefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = l.load("/path/does/not/exist")
	if err == nil {
		t.Fatal("expected error in Load, but no error returned")
	}
}
