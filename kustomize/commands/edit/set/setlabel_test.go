// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"reflect"
	"testing"

	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func makeKustomization(t *testing.T) *types.Kustomization {
	t.Helper()
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

func TestRunSetLabels(t *testing.T) {
	var o setLabelOptions
	o.includeSelectors = true
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	err := o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}

	// assert content
	expectedContent := map[string]string{"app": "helloworld", "owls": "cute", "otters": "adorable"}
	if !reflect.DeepEqual(m.CommonLabels, expectedContent) {
		t.Log("m.CommonLabels", m.CommonLabels)
		t.Log("expectedContent", expectedContent)
		t.Errorf("commonLabels does not contain expected content")
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

func TestRunSetLabelsNoSelector(t *testing.T) {
	var o setLabelOptions
	o.includeSelectors = false
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	err := o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}

	// assert content
	expectedContent := make([]types.Label, 2)
	expectedContent[0] = types.Label{Pairs: map[string]string{"owls": "cute"}, IncludeSelectors: false}
	expectedContent[1] = types.Label{Pairs: map[string]string{"otters": "adorable"}, IncludeSelectors: false}
	if !reflect.DeepEqual(m.Labels, expectedContent) {
		t.Log("m.Labels", m.Labels)
		t.Log("expectedContent", expectedContent)
		t.Errorf("labels does not contain expected content")
	}

	// adding the same test input should work
	err = o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}

	// assert content
	expectedContent2 := make([]types.Label, 2)
	expectedContent2[0] = types.Label{Pairs: map[string]string{"owls": "cute"}, IncludeSelectors: false}
	expectedContent2[1] = types.Label{Pairs: map[string]string{"otters": "adorable"}, IncludeSelectors: false}
	if !reflect.DeepEqual(m.Labels, expectedContent2) {
		t.Log("m.Labels", m.Labels)
		t.Log("expectedContent", expectedContent2)
		t.Errorf("labels does not contain expected content")
	}

	// adding new labels should work
	o.metadata = map[string]string{"new": "label", "owls": "not cute"}
	err = o.setLabels(m)
	if err != nil {
		t.Errorf("unexpected error: could not write to kustomization file")
	}

	// assert content
	expectedContent3 := make([]types.Label, 3)
	expectedContent3[0] = types.Label{Pairs: map[string]string{"owls": "not cute"}, IncludeSelectors: false}
	expectedContent3[1] = types.Label{Pairs: map[string]string{"otters": "adorable"}, IncludeSelectors: false}
	expectedContent3[2] = types.Label{Pairs: map[string]string{"new": "label"}, IncludeSelectors: false}
	if !reflect.DeepEqual(m.Labels, expectedContent3) {
		t.Log("m.Labels", m.Labels)
		t.Log("expectedContent", expectedContent3)
		t.Errorf("labels does not contain expected content")
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
