// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	require.NoError(t, err)
	m, err := kf.Read()
	require.NoError(t, err)
	return m
}

func TestRunAddAnnotation(t *testing.T) {
	var o addMetadataOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	require.NoError(t, o.addAnnotations(m))
	// adding the same test input should not work
	require.Error(t, o.addAnnotations(m))

	// adding new annotations should work
	o.metadata = map[string]string{"new": "annotation"}
	require.NoError(t, o.addAnnotations(m))
}

func TestAddAnnotationNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	err := cmd.Execute()
	v.VerifyNoCall()
	require.Error(t, err)
	assert.Equal(t, "must specify annotation", err.Error())
}

func TestAddAnnotationInvalidFormat(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"whatever:whatever"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	require.Error(t, err)
	assert.Equal(t, valtest_test.SAD, err.Error())
}

func TestAddAnnotationManyArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"k1:v1,k2:v2,k3:v3,k4:v5"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddAnnotationValueQuoted(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"k1:\"v1\""}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddAnnotationValueWithColon(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"k1:\"v1:v2\""}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddAnnotationValueWithComma(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	value := "{\"k1\":\"v1\",\"k2\":\"v2\"}"
	args := []string{"test:" + value}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
	b, err := fSys.ReadFile("/kustomization.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(b), value)
}

func TestAddAnnotationNoKey(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{":nokey"}
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	require.Error(t, err)
	assert.Equal(t, "invalid annotation: ':nokey' (need k:v pair where v may be quoted)", err.Error())
}

func TestAddAnnotationTooManyColons(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"key:v1:v2"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddAnnotationNoValue(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"no:,value"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddAnnotationMultipleArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"this:annotation", "has:spaces"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddAnnotationForce(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddAnnotation(fSys, v.Validator)
	args := []string{"key:foo"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
	// trying to add the same annotation again should not work
	args = []string{"key:bar"}
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddAnnotation(fSys, v.Validator)
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	require.Error(t, err)
	assert.Equal(t, "annotation key already in kustomization file. Use --force to override.", err.Error())
	// but trying to add it with --force should
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddAnnotation(fSys, v.Validator)
	err = cmd.Flag("force").Value.Set("true")
	require.NoError(t, err)
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestRunAddLabel(t *testing.T) {
	var o addMetadataOptions
	o.metadata = map[string]string{"owls": "cute", "otters": "adorable"}

	m := makeKustomization(t)
	require.NoError(t, o.addLabels(m))
	// adding the same test input should not work
	require.Error(t, o.addLabels(m))
	// adding new labels should work
	o.metadata = map[string]string{"new": "label"}
	require.NoError(t, o.addLabels(m))
}

func TestAddLabelNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	err := cmd.Execute()
	v.VerifyNoCall()
	require.Error(t, err)
	assert.Equal(t, "must specify label", err.Error())
}

func TestAddLabelInvalidFormat(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeSadMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"exclamation!:point"}
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddLabelNoValue(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"no,value:"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddLabelMultipleArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"this:input", "has:spaces"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddLabelForce(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"key:foo"}
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
	// trying to add the same label again should not work
	args = []string{"key:bar"}
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddLabel(fSys, v.Validator)
	err := cmd.RunE(cmd, args)
	v.VerifyCall()
	require.Error(t, err)
	assert.Equal(t, "label key already in kustomization file. Use --force to override.", err.Error())
	// but trying to add it with --force should
	v = valtest_test.MakeHappyMapValidator(t)
	cmd = newCmdAddLabel(fSys, v.Validator)
	err = cmd.Flag("force").Value.Set("true")
	require.NoError(t, err)
	require.NoError(t, cmd.RunE(cmd, args))
	v.VerifyCall()
}

func TestAddLabelIncludeTemplatesWithoutRequiredFlag(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	v := valtest_test.MakeHappyMapValidator(t)
	cmd := newCmdAddLabel(fSys, v.Validator)
	args := []string{"new:label"}
	_ = cmd.Flag("include-templates").Value.Set("true")
	_ = cmd.Flag("without-selector").Value.Set("false")
	err := cmd.RunE(cmd, args)
	v.VerifyNoCall()
	require.Error(t, err)
	require.Containsf(t, err.Error(), "--without-selector flag must be specified for --include-templates to work", "incorrect error: %s", err.Error())
}

func TestAddLabelWithoutSelector(t *testing.T) {
	tests := []struct {
		name        string
		baseLabels  []types.Label      // original labels
		options     addMetadataOptions // add labels
		metadata    map[string]string  // added labels
		expected    []types.Label      // expected labels
		expectedErr string
	}{
		{
			name:       "add to empty labels",
			baseLabels: []types.Label{},
			options: addMetadataOptions{
				labelsWithoutSelector: true,
				metadata:              map[string]string{"new": "label"},
			},
			expected: []types.Label{
				{
					Pairs: map[string]string{"new": "label"},
				},
			},
		},
		{
			name:       "add to empty labels with IncludeTemplates",
			baseLabels: []types.Label{},
			options: addMetadataOptions{
				labelsWithoutSelector: true,
				includeTemplates:      true,
				metadata:              map[string]string{"new": "label"},
			},
			expected: []types.Label{
				{
					Pairs:            map[string]string{"new": "label"},
					IncludeTemplates: true,
				},
			},
		},
		{
			name: "overwrite label requires force",
			baseLabels: []types.Label{
				{
					Pairs: map[string]string{"key1": "old-value1"},
				},
			},
			options: addMetadataOptions{
				labelsWithoutSelector: true,
				force:                 false,
				metadata:              map[string]string{"key1": "new-value1"},
			},
			expectedErr: "label key1 already in kustomization file",
		},
		{
			name: "overwrite label",
			baseLabels: []types.Label{
				{
					Pairs: map[string]string{"key1": "old-value1"},
				},
			},
			options: addMetadataOptions{
				labelsWithoutSelector: true,
				force:                 true,
				metadata:              map[string]string{"key1": "new-value1"},
			},
			expected: []types.Label{
				{
					Pairs: map[string]string{"key1": "new-value1"},
				},
			},
		},
		{
			name: "overwrite and add label",
			baseLabels: []types.Label{
				{
					Pairs: map[string]string{"key1": "old-value1"},
				},
			},
			options: addMetadataOptions{
				labelsWithoutSelector: true,
				force:                 true,
				metadata:              map[string]string{"key1": "new-value1", "key2": "value2"},
			},
			expected: []types.Label{
				{
					Pairs: map[string]string{"key1": "new-value1", "key2": "value2"},
				},
			},
		},
		{
			name: "overwrite label with same settings",
			baseLabels: []types.Label{
				{
					Pairs: map[string]string{"key": "old-value"},
				},
				{
					IncludeTemplates: true,
					Pairs:            map[string]string{"key": "old-value"},
				},
			},
			options: addMetadataOptions{
				labelsWithoutSelector: true,
				force:                 true,
				metadata:              map[string]string{"key": "new-value"},
				includeTemplates:      true,
			},
			expected: []types.Label{
				{
					Pairs: map[string]string{"key": "old-value"},
				},
				{
					IncludeTemplates: true,
					Pairs:            map[string]string{"key": "new-value"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeKustomization(t)
			m.Labels = tt.baseLabels
			err := tt.options.addLabels(m)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, m.Labels)
			}
		})
	}
}
