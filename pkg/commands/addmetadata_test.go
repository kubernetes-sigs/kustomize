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
)

func TestRunAddAnnotation(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	var o addMetadataOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	err := o.RunAddAnnotation(fakeFS, annotation)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding the same test input should not work
	err = o.RunAddAnnotation(fakeFS, annotation)
	if err == nil {
		t.Errorf("expected already in kustomization file error")
	}
	// adding new annotations should work
	o.metadata = map[string]string{"new": "annotation"}
	err = o.RunAddAnnotation(fakeFS, annotation)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
}

func TestAddAnnotationNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdAddAnnotation(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but error is %v", err)
	}
	if err != nil && err.Error() != "must specify annotation" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationInvalidFormat(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdAddAnnotation(fakeFS)
	args := []string{"exclamation!:point"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is %v", err)
	}
	if err != nil && err.Error() != "invalid annotation format: exclamation!:point" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationNoKey(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdAddAnnotation(fakeFS)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is %v", err)
	}
	if err != nil && err.Error() != "invalid annotation format: :nokey" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationNoValue(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddAnnotation(fakeFS)
	args := []string{"no:,value"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationMultipleArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddAnnotation(fakeFS)
	args := []string{"this:annotation", "has:spaces"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is %v", err)
	}
	if err != nil && err.Error() != "annotations must be comma-separated, with no spaces. See help text for example" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRunAddLabel(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	var o addMetadataOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	err := o.RunAddLabel(fakeFS, label)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding the same test input should not work
	err = o.RunAddLabel(fakeFS, label)
	if err == nil {
		t.Errorf("expected already in kustomization file error")
	}
	// adding new labels should work
	o.metadata = map[string]string{"new": "label"}
	err = o.RunAddLabel(fakeFS, label)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
}

func TestAddLabelNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdAddLabel(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but error is: %v", err)
	}
	if err != nil && err.Error() != "must specify label" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelInvalidFormat(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdAddLabel(fakeFS)
	args := []string{"exclamation!:point"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is: %v", err)
	}
	if err != nil && err.Error() != "invalid label format: exclamation!:point" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelNoKey(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := newCmdAddLabel(fakeFS)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is: %v", err)
	}
	if err != nil && err.Error() != "invalid label format: :nokey" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelNoValue(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddLabel(fakeFS)
	args := []string{"no,value:"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddLabelMultipleArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddLabel(fakeFS)
	args := []string{"this:input", "has:spaces"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is: %v", err)
	}
	if err != nil && err.Error() != "labels must be comma-separated, with no spaces. See help text for example" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
