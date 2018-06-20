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

package commands

import (
	"testing"

	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

const (
	baseDirectoryPaths = "my/path/to/wonderful/base,other/path/to/even/more/wonderful/base"
)

func TestAddBaseHappyPath(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		fakeFS.Mkdir(base, 0777)
	}
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdAddBase(fakeFS)
	args := []string{baseDirectoryPaths}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadFile(constants.KustomizationFileName)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}

	for _, base := range bases {
		if !strings.Contains(string(content), base) {
			t.Errorf("expected base name in kustomization")
		}
	}
}

func TestAddBaseAlreadyThere(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	// Create fake directories
	bases := strings.Split(baseDirectoryPaths, ",")
	for _, base := range bases {
		fakeFS.Mkdir(base, 0777)
	}
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdAddBase(fakeFS)
	args := []string{baseDirectoryPaths}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}
	// adding an existing base should return an error
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected already there problem")
	}
	var expectedErrors []string
	for _, base := range bases {
		error := "base " + base + " already in kustomization file"
		expectedErrors = append(expectedErrors, error)
		if !stringInSlice(error, expectedErrors) {
			t.Errorf("unexpected error %v", err)
		}
	}

}

func TestAddBaseNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdAddBase(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a base directory" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
