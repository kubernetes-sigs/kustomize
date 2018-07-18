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
	"os"
	"path"
	"reflect"
	"testing"
)

func TestReadFilesRealFS(t *testing.T) {
	x := MakeRealFS()
	testDir := "kustomize_testing_dir"
	err := x.Mkdir(testDir)
	defer os.RemoveAll(testDir)

	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !x.Exists(testDir) {
		t.Fatalf("expected existence")
	}
	if !x.IsDir(testDir) {
		t.Fatalf("expected directory")
	}

	err = x.WriteFile(path.Join(testDir, "foo"), []byte(`foo`))
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !x.Exists(path.Join(testDir, "foo")) {
		t.Fatalf("expected foo")
	}
	if x.IsDir(path.Join(testDir, "foo")) {
		t.Fatalf("expected foo not to be a directory")
	}

	err = x.WriteFile(path.Join(testDir, "bar"), []byte(`bar`))
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	expected := map[string][]byte{
		path.Join(testDir, "foo"): []byte(`foo`),
		path.Join(testDir, "bar"): []byte(`bar`),
	}

	content, err := x.ReadFiles("kustomize_testing_dir/*")
	if !reflect.DeepEqual(content, expected) {
		t.Fatalf("actual: %+v doesn't match expected: %+v", content, expected)

	}
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}
