// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

func makeKustomization(t *testing.T) *types.Kustomization {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
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
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
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
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"whatever:whatever"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != valtest_test.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationManyArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"k1:v1,k2:v2,k3:v3,k4:v5"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationValueQuoted(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"k1:\"v1\""}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationValueWithColon(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"k1:\"v1:v2\""}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationValueWithComma(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	value := "{\"k1\":\"v1\",\"k2\":\"v2\"}"
	args := []string{"test:" + value}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
	b, err := fSys.ReadFile("/kustomization.yaml")
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
	if !strings.Contains(string(b), value) {
		t.Errorf(
			"Modified file doesn't contain expected string.\nExpected string:\n%s\nActual:\n%s",
			value, b)
	}
}

func TestAddAnnotationNoKey(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid annotation: ':nokey' (need k:v pair where v may be quoted)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationTooManyColons(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"key:v1:v2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationNoValue(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"no:,value"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationMultipleArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"this:annotation", "has:spaces"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddAnnotationForce(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"key:foo"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
	// trying to add the same annotation again should not work
	args = []string{"key:bar"}
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddAnnotation(fSys, v.Validator)
	err = cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "annotation key already in kustomization file" {
		t.Errorf("expected an error")
	}
	// but trying to add it with --force should
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddAnnotation(fSys, v.Validator)
	cmd.Flag("force").Value.Set("true")
	err = cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
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
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
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
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"exclamation!:point"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != valtest_test.SAD {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelNoKey(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid label: ':nokey' (need k:v pair where v may be quoted)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelTooManyColons(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"key:v1:v2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddLabelNoValue(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"no,value:"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddLabelMultipleArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"this:input", "has:spaces"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestAddLabelForce(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"key:foo"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
	// trying to add the same label again should not work
	args = []string{"key:bar"}
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddLabel(fSys, v.Validator)
	err = cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "label key already in kustomization file" {
		t.Errorf("expected an error")
	}
	// but trying to add it with --force should
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddLabel(fSys, v.Validator)
	cmd.Flag("force").Value.Set("true")
	err = cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}
