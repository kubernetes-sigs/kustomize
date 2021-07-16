// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"testing"

	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func makeAnnotationKustomization(t *testing.T) *types.Kustomization {
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

func TestRunSetAnnotation(t *testing.T) {
	var o setAnnotationOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeAnnotationKustomization(t)
	err := o.setAnnotations(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding the same test input should work
	err = o.setAnnotations(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
	// adding new annotations should work
	o.metadata = map[string]string{"new": "annotation", "owls": "not cute"}
	err = o.setAnnotations(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}
}

func TestSetAnnotationNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
	err := cmd.Execute()
	v.VerifyNoCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "must specify annotation" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestSetAnnotationInvalidFormat(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
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

func TestSetAnnotationNoKey(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
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

func TestSetAnnotationTooManyColons(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
	args := []string{"key:v1:v2"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestSetAnnotationNoValue(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
	args := []string{"no,value:"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid annotation key: see the syntax and character set rules at https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestSetAnnotationMultipleArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
	args := []string{"this:input", "has:spaces"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestSetAnnotationExisting(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdSetAnnotation(fSys, v.Validator)
	args := []string{"key:foo"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdSetAnnotation(fSys, v.Validator)
	err = cmd.RunE(cmd, args)
	v.VerifyCall()
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}
