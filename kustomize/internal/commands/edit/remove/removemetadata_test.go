// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

func makeKustomizationFS() filesys.FileSystem {
	fSys := filesys.MakeFsInMemory()
	commonLabels := []string{"label1: val1", "label2: val2"}
	commonAnnotations := []string{"annotation1: val1", "annotation2: val2"}

	testutils_test.WriteTestKustomizationWith(fSys, []byte(
		fmt.Sprintf("commonLabels:\n  %s\ncommonAnnotations:\n  %s",
			strings.Join(commonLabels, "\n  "), strings.Join(commonAnnotations, "\n  "))))
	return fSys
}

func readKustomizationFS(t *testing.T, fSys filesys.FileSystem) *types.Kustomization {
	kf, err := kustfile.NewKustomizationFile(fSys)
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
	fSys := makeKustomizationFS()
	return readKustomizationFS(t, fSys)
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
	fSys := makeKustomizationFS()

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fSys, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"annotation3"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveAnnotationNoDefinition(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, []byte(""))

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fSys, v.ValidatorArray)
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
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, []byte(""))

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"annotation1,annotation2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveAnnotationNoArgs(t *testing.T) {
	fSys := makeKustomizationFS()

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fSys, v.ValidatorArray)
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
	fSys := makeKustomizationFS()

	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdRemoveAnnotation(fSys, v.ValidatorArray)
	args := []string{"nospecialchars%^=@"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != valtest_test.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveAnnotationMultipleArgs(t *testing.T) {
	fSys := makeKustomizationFS()

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveAnnotation(fSys, v.ValidatorArray)
	args := []string{"annotation1,annotation2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fSys)
	splitArgs := strings.Split(args[0], ",")
	for _, k := range splitArgs {
		if _, exist := m.CommonAnnotations[k]; exist {
			t.Errorf("%s must be deleted", k)
		}
	}
}

func TestRemoveAnnotationMultipleArgsInvalidFormat(t *testing.T) {
	fSys := makeKustomizationFS()

	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdRemoveAnnotation(fSys, v.ValidatorArray)
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
	fSys := makeKustomizationFS()

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"label3"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveLabelNoDefinition(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, []byte(""))

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
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
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, []byte(""))

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
	cmd.Flag("ignore-non-existence").Value.Set("true")
	args := []string{"label1,label2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestRemoveLabelNoArgs(t *testing.T) {
	fSys := makeKustomizationFS()

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
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
	fSys := makeKustomizationFS()

	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
	args := []string{"exclamation!"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != valtest_test.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestRemoveLabelMultipleArgs(t *testing.T) {
	fSys := makeKustomizationFS()

	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
	args := []string{"label1,label2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	m := readKustomizationFS(t, fSys)
	splitArgs := strings.Split(args[0], ",")
	for _, k := range splitArgs {
		if _, exist := m.CommonLabels[k]; exist {
			t.Errorf("%s must be deleted", k)
		}
	}
}

func TestRemoveLabelMultipleArgsInvalidFormat(t *testing.T) {
	fSys := makeKustomizationFS()

	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdRemoveLabel(fSys, v.ValidatorArray)
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
