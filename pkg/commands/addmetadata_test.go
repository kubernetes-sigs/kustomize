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

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/kubernetes-sigs/kustomize/pkg/validators"
)

func makeKustomization(t *testing.T) *types.Kustomization {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	kf, err := newKustomizationFile(constants.KustomizationFileName, fakeFS)
	if err != nil {
		t.Errorf("unexpected new error %v", err)
	}
	m, err := kf.read()
	if err != nil {
		t.Errorf("unexpected read error %v", err)
	}
	return m
}

func TestRunAddAnnotation(t *testing.T) {
	var o addMetadataOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	err := o.addAnnotations(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding the same test input should not work
	err = o.addAnnotations(m)
	if err == nil {
		t.Errorf("expected already in kustomization file error")
	}
	// adding new annotations should work
	o.metadata = map[string]string{"new": "annotation"}
	err = o.addAnnotations(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
}

func TestAddAnnotationNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	err := cmd.Execute()
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify annotation" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationInvalidFormat(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeSadMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	args := []string{"whatever:whatever"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != validators.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationManyArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	args := []string{"k1:v1,k2:v2,k3:v3,k4:v5"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationNoKey(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid annotation: :nokey (empty key)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationTooManyColons(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	args := []string{"key:v1:v2"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid annotation: key:v1:v2 (too many colons)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationNoValue(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	args := []string{"no:,value"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationMultipleArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fakeFS, v.Validator)
	args := []string{"this:annotation", "has:spaces"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "annotations must be comma-separated, with no spaces" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRunAddLabel(t *testing.T) {
	var o addMetadataOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	err := o.addLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding the same test input should not work
	err = o.addLabels(m)
	if err == nil {
		t.Errorf("expected already in kustomization file error")
	}
	// adding new labels should work
	o.metadata = map[string]string{"new": "label"}
	err = o.addLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
}

func TestAddLabelNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fakeFS, v.Validator)
	err := cmd.Execute()
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify label" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelInvalidFormat(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeSadMapValidator(t)
	cmd := newCmdAddLabel(fakeFS, v.Validator)
	args := []string{"exclamation!:point"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != validators.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelNoKey(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fakeFS, v.Validator)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid label: :nokey (empty key)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelTooManyColons(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fakeFS, v.Validator)
	args := []string{"key:v1:v2"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid label: key:v1:v2 (too many colons)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelNoValue(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fakeFS, v.Validator)
	args := []string{"no,value:"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddLabelMultipleArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	v := validators.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fakeFS, v.Validator)
	args := []string{"this:input", "has:spaces"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "labels must be comma-separated, with no spaces" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
