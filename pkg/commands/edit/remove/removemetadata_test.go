/*
Copyright 2019 The Kubernetes Authors.

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

package remove

import (
	"fmt"
	"sigs.k8s.io/kustomize/v3/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"strings"
	"testing"
)

func makeKustomizationFS() fs.FileSystem {
	fakeFS := fs.MakeFakeFS()
	commonLabels := []string{"label1: val1", "label2: val2"}
	commonAnnotations := []string{"annotation1: val1", "annotation2: val2"}

	fakeFS.WriteTestKustomizationWith([]byte(
		fmt.Sprintf("commonLabels:\n  %s\ncommonAnnotations:\n  %s",
			strings.Join(commonLabels, "\n  "), strings.Join(commonAnnotations, "\n  "))))
	return fakeFS
}

func readKustomizationFS(t *testing.T, fakeFS fs.FileSystem) *types.Kustomization {
	kf, err := kustfile.NewKustomizationFile(fakeFS)
	if err != nil {
		t.Errorf("unexpected new error %v", err)
	}
	m, err := kf.Read()
	if err != nil {
		t.Errorf("unexpected read error %v", err)
	}
	return m
}

func makeKustomization(t *testing.T) *types.Kustomization {
	fakeFS := makeKustomizationFS()
	return readKustomizationFS(t, fakeFS)
}

func TestRemoveAnnotation(t *testing.T) {
	var o removeMetadataOptions
	o.metadata = []string{"annotation1"}

	m := makeKustomization(t)
	err := o.removeAnnotations(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}

	// adding the same test input should not work
	err = o.removeAnnotations(m)
	if err == nil {
		t.Errorf("expected not exist in kustomization file error")
	}

	_, exists := m.CommonAnnotations["annotation1"]
	if exists {
		t.Errorf("annotation1 must be deleted")
	}

	_, exists = m.CommonAnnotations["annotation2"]
	if !exists {
		t.Errorf("annotation2 must exist")
	}
}

func TestRemoveAnnotationIgnore(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fakeFS, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"annotation3"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveAnnotationNoDefinition(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomizationWith([]byte(""))

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fakeFS, v.ValidatorArray)
	args := []string{"annotation1,annotation2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "commonAnnotations is not defined in kustomization file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveAnnotationNoDefinitionIgnore(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomizationWith([]byte(""))

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"annotation1,annotation2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveAnnotationNoArgs(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fakeFS, v.ValidatorArray)
	err := cmd.Execute()
	v.VerifyNoCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify label" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveAnnotationInvalidFormat(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeSadMapValidator(t)
	cmd := newCmdRemoveAnnotation(fakeFS, v.ValidatorArray)
	args := []string{"nospecialchars%^=@"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != validators.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveAnnotationMultipleArgs(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fakeFS, v.ValidatorArray)
	args := []string{"annotation1,annotation2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	splitArgs := strings.Split(args[0], ",")
	for _, k := range splitArgs {
		if _, exist := m.CommonAnnotations[k]; exist {
			t.Errorf("%s must be deleted", k)
		}
	}
}

func TestRemoveAnnotationMultipleArgsInvalidFormat(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeSadMapValidator(t)
	cmd := newCmdRemoveAnnotation(fakeFS, v.ValidatorArray)
	args := []string{"annotation1", "annotation2"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "labels must be comma-separated, with no spaces" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveLabel(t *testing.T) {
	var o removeMetadataOptions
	o.metadata = []string{"label1"}

	m := makeKustomization(t)
	err := o.removeLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}

	// adding the same test input should not work
	err = o.removeLabels(m)
	if err == nil {
		t.Errorf("expected not exist in kustomization file error")
	}

	_, exists := m.CommonLabels["label1"]
	if exists {
		t.Errorf("label1 must be deleted")
	}

	_, exists = m.CommonLabels["label2"]
	if !exists {
		t.Errorf("label2 must exist")
	}
}

func TestRemoveLabelIgnore(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"label3"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveLabelNoDefinition(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomizationWith([]byte(""))

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	args := []string{"label1,label2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "commonLabels is not defined in kustomization file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveLabelNoDefinitionIgnore(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomizationWith([]byte(""))

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"label1,label2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveLabelNoArgs(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	err := cmd.Execute()
	v.VerifyNoCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify label" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveLabelInvalidFormat(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeSadMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	args := []string{"exclamation!"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != validators.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveLabelMultipleArgs(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	args := []string{"label1,label2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fakeFS)
	splitArgs := strings.Split(args[0], ",")
	for _, k := range splitArgs {
		if _, exist := m.CommonLabels[k]; exist {
			t.Errorf("%s must be deleted", k)
		}
	}
}

func TestRemoveLabelMultipleArgsInvalidFormat(t *testing.T) {
	fakeFS := makeKustomizationFS()

	v := validators.MakeSadMapValidator(t)
	cmd := newCmdRemoveLabel(fakeFS, v.ValidatorArray)
	args := []string{"label1", "label2"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "labels must be comma-separated, with no spaces" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
