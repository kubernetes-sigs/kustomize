// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
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

func TestRunSetLabel(t *testing.T) {
	var o setLabelOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	err := o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding the same test input should work
	err = o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding new labels should work
	o.metadata = map[string]string{"new": "label", "owls": "not cute"}
	err = o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
}

func TestSetLabelNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
	err := cmd.Execute()
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify label" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestSetLabelInvalidFormat(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
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

func TestSetLabelNoKey(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
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

func TestSetLabelTooManyColons(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
	args := []string{"key:v1:v2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestSetLabelNoValue(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
	args := []string{"no,value:"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestSetLabelMultipleArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
	args := []string{"this:input", "has:spaces"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestSetLabelExisting(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetLabel(fSys, v.Validator)
	args := []string{"key:foo"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdSetLabel(fSys, v.Validator)
	err = cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}
