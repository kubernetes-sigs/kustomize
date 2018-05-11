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

	"strings"

	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

const (
	goodPrefixValue = "acme-"
)

func TestSetNamePrefixHappyPath(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdSetNamePrefix(buf, os.Stderr, fakeFS)
	args := []string{goodPrefixValue}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadFile(constants.KustomizationFileName)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), goodPrefixValue) {
		t.Errorf("expected prefix value in kustomization file")
	}
}

func TestSetNamePrefixNoArgs(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdSetNamePrefix(buf, os.Stderr, fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify exactly one prefix value" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
