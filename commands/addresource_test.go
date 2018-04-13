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
	resourceFileName    = "myWonderfulResource.yaml"
	resourceFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
	kustomizationContent = `kustomizationName: helloworld
namePrefix: some-prefix
# Labels to add to all objects and selectors.
# These labels would also be used to form the selector for apply --prune
# Named differently than “labels” to avoid confusion with metadata for this object
objectLabels:
  app: helloworld
objectAnnotations:
  note: This is an example annotation
resources: []
#- service.yaml
#- ../some-dir/
# There could also be configmaps in Base, which would make these overlays
configMapGenerator: []
# There could be secrets in Base, if just using a fork/rebase workflow
secretGenerator: []
`
)

func TestAddResourceHappyPath(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(resourceFileName, []byte(resourceFileContent))
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdAddResource(buf, os.Stderr, fakeFS)
	args := []string{resourceFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadFile(constants.KustomizationFileName)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), resourceFileName) {
		t.Errorf("expected resource name in kustomization")
	}
}

func TestAddResourceAlreadyThere(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(resourceFileName, []byte(resourceFileContent))
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdAddResource(buf, os.Stderr, fakeFS)
	args := []string{resourceFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}

	// adding an existing resource should return an error
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected already there problem")
	}
	if err.Error() != "resource "+resourceFileName+" already in kustomization file" {
		t.Errorf("unexpected error %v", err)
	}
}

func TestAddResourceNoArgs(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdAddResource(buf, os.Stderr, fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a resource file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
