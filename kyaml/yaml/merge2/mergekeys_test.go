// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	. "sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

// TestMergeWithCustomMergeKey demonstrates that a CRD-like resource with no
// registered schema has its list replaced by default, but when MergeKeySpecs
// are provided the list items are merged instead.
func TestMergeWithCustomMergeKey(t *testing.T) {
	base := `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: BASE
    value: base`

	patch := `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: PATCH
    value: patch`

	got, err := MergeStrings(patch, base, false, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListPrepend,
		MergeKeySpecs: []yaml.MergeKeySpec{
			{Path: []string{"spec", "env"}, Key: "name"},
		},
	})
	require.NoError(t, err)
	assert.Contains(t, got, "BASE")
	assert.Contains(t, got, "PATCH")
}

// TestMergeKeyPathResetsInsideAssociativeList documents the path-scoping
// behaviour when a target list is nested inside another associative list.
// Because the walker does not propagate Path through associative-list element
// boundaries, the path in MergeKeySpec must be relative to the list-element
// level where the target list lives, NOT from the document root.
//
// Here the target is spec.volumes[*].configMap.items. Each volume element is
// matched by the outer associative list (key "name"), and inside a matched
// element the path resets to []. So the correct path to declare is
// ["configMap", "items"], not ["spec", "volumes", "configMap", "items"].
func TestMergeKeyPathResetsInsideAssociativeList(t *testing.T) {
	base := `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  volumes:
  - name: config
    configMap:
      items:
      - key: foo
        path: foo`

	patch := `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  volumes:
  - name: config
    configMap:
      items:
      - key: bar
        path: bar`

	// Correct: path is relative to inside a volume element.
	got, err := MergeStrings(patch, base, false, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListPrepend,
		MergeKeySpecs: []yaml.MergeKeySpec{
			{Path: []string{"configMap", "items"}, Key: "key"},
		},
	})
	require.NoError(t, err)
	assert.Contains(t, got, "key: foo", "base item should be present")
	assert.Contains(t, got, "key: bar", "patch item should be present")

	// Wrong: root-relative path does NOT match and items are replaced.
	gotWrong, err := MergeStrings(patch, base, false, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListPrepend,
		MergeKeySpecs: []yaml.MergeKeySpec{
			{Path: []string{"spec", "volumes", "configMap", "items"}, Key: "key"},
		},
	})
	require.NoError(t, err)
	assert.NotContains(t, gotWrong, "key: foo", "base item should be replaced when path is wrong")
}

// TestMergeWithoutCustomMergeKeyReplacesLists confirms that without
// MergeKeySpecs the list is replaced (existing behaviour for CRDs).
func TestMergeWithoutCustomMergeKeyReplacesLists(t *testing.T) {
	base := `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: BASE
    value: base`

	patch := `apiVersion: example.com/v1
kind: MyApp
metadata:
  name: app
spec:
  env:
  - name: PATCH
    value: patch`

	got, err := MergeStrings(patch, base, false, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListPrepend,
	})
	require.NoError(t, err)
	// Without custom merge key the patch list replaces the base list.
	assert.NotContains(t, got, "BASE")
	assert.Contains(t, got, "PATCH")
}
