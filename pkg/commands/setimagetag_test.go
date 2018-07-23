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
	"strings"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

func TestSetImageTagsHappyPath(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdSetImageTag(fakeFS)
	args := []string{"image1:tag1", "image2:tag2"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadFile(constants.KustomizationFileName)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	expected := []byte(`
imageTags:
- name: image1
  newTag: tag1
- name: image2
  newTag: tag2
`)
	if !strings.Contains(string(content), string(expected)) {
		t.Errorf("expected imageTags in kustomization file")
	}
}

func TestSetImageTagsOverride(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))

	cmd := newCmdSetImageTag(fakeFS)
	args := []string{"image1:tag1", "image2:tag1"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	args = []string{"image2:tag2", "image3:tag3"}
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadFile(constants.KustomizationFileName)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	expected := []byte(`
imageTags:
- name: image1
  newTag: tag1
- name: image2
  newTag: tag2
- name: image3
  newTag: tag3
`)
	if !strings.Contains(string(content), string(expected)) {
		t.Errorf("expected imageTags in kustomization file %s", string(content))
	}
}

func TestSetImageTagsNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdSetImageTag(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "No image and newTag specified." {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
