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

package set

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
)

func TestSetImageTagsHappyPath(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomization()

	cmd := newCmdSetImageTag(fakeFS)
	args := []string{"image1:tag1", "image2:tag2", "localhost:5000/operator:1.0.0",
		"foo.bar.baz:5000/one/two@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"}
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
- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
  name: foo.bar.baz:5000/one/two
- name: image1
  newTag: tag1
- name: image2
  newTag: tag2
- name: localhost:5000/operator
  newTag: 1.0.0
`)
	if !strings.Contains(string(content), string(expected)) {
		t.Errorf("expected imageTags in kustomization file")
	}
}

func TestSetImageTagsOverride(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomization()

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
	if err.Error() != "no image specified" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
