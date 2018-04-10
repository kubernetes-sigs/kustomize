/*
Copyright 2017 The Kubernetes Authors.

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

package commands

import (
	"bytes"
	"os"
	"testing"

	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

func TestInitHappyPath(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdInit(buf, os.Stderr, fakeFS)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := fakeFS.Open(constants.KustomizeFileName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	file := f.(*fs.FakeFile)
	if !file.ContentMatches([]byte(manifestTemplate)) {
		t.Fatalf("actual: %v doesn't match expected: %v",
			string(file.GetContent()), manifestTemplate)
	}
}

func TestInitFileAlreadyExist(t *testing.T) {
	content := "hey there"
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizeFileName, []byte(content))

	buf := bytes.NewBuffer([]byte{})
	cmd := newCmdInit(buf, os.Stderr, fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != `"`+constants.KustomizeFileName+`" already exists` {
		t.Fatalf("unexpected error: %v", err)
	}
}
