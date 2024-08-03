// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"testing"

	"github.com/stretchr/testify/require"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
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

func TestSetLabelWithoutSelector(t *testing.T) {
	var o setLabelOptions
	o.metadata = map[string]string{"key1": "foo", "key2": "bar"}
	o.labelsWithoutSelector = true

	m := makeKustomization(t)
	require.NoError(t, o.setLabels(m))
	require.Equal(t, m.Labels[0], types.Label{Pairs: map[string]string{"key1": "foo", "key2": "bar"}})
}

func TestSetLabelWithoutSelectorWithExistingLabels(t *testing.T) {
	var o setLabelOptions
	o.metadata = map[string]string{"key1": "foo", "key2": "bar"}
	o.labelsWithoutSelector = true

	m := makeKustomization(t)
	require.NoError(t, o.setLabels(m))
	require.Equal(t, m.Labels[0], types.Label{Pairs: map[string]string{"key1": "foo", "key2": "bar"}})

	o.metadata = map[string]string{"key3": "foobar"}
	require.NoError(t, o.setLabels(m))
	require.Equal(t, m.Labels[0], types.Label{Pairs: map[string]string{"key1": "foo", "key2": "bar", "key3": "foobar"}})
}

func TestSetLabelWithoutSelectorWithDuplicateLabel(t *testing.T) {
	var o setLabelOptions
	o.metadata = map[string]string{"key1": "foo", "key2": "bar"}
	o.labelsWithoutSelector = true
	o.includeTemplates = true

	m := makeKustomization(t)
	require.NoError(t, o.setLabels(m))
	require.Equal(t, m.Labels[0], types.Label{Pairs: map[string]string{"key1": "foo", "key2": "bar"}, IncludeTemplates: true})

	o.metadata = map[string]string{"key2": "bar"}
	o.includeTemplates = false
	require.NoError(t, o.setLabels(m))
	require.Equal(t, m.Labels[0], types.Label{Pairs: map[string]string{"key1": "foo"}, IncludeTemplates: true})
	require.Equal(t, m.Labels[1], types.Label{Pairs: map[string]string{"key2": "bar"}})
}

func TestSetLabelWithoutSelectorWithCommonLabel(t *testing.T) {
	var o setLabelOptions
	o.metadata = map[string]string{"key1": "foo", "key2": "bar"}

	m := makeKustomization(t)
	require.NoError(t, o.setLabels(m))
	require.Empty(t, m.Labels)
	require.Equal(t, m.CommonLabels, map[string]string{"app": "helloworld", "key1": "foo", "key2": "bar"})

	o.metadata = map[string]string{"key2": "bar"}
	o.labelsWithoutSelector = true
	require.NoError(t, o.setLabels(m))
	require.Equal(t, m.CommonLabels, map[string]string{"app": "helloworld", "key1": "foo"})
	require.Equal(t, m.Labels[0], types.Label{Pairs: map[string]string{"key2": "bar"}})
}
