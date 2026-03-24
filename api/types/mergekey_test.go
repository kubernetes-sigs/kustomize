// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

func TestMergeKeySpecRoundTrip(t *testing.T) {
	y := `mergeKeys:
- kind: HelmRelease
  group: helm.toolkit.fluxcd.io
  path: spec/values/myapp/env
  key: name
`
	k := &Kustomization{}
	require.NoError(t, yaml.Unmarshal([]byte(y), k))
	require.Len(t, k.MergeKeys, 1)
	assert.Equal(t, "HelmRelease", k.MergeKeys[0].Kind)
	assert.Equal(t, "helm.toolkit.fluxcd.io", k.MergeKeys[0].Group)
	assert.Equal(t, "spec/values/myapp/env", k.MergeKeys[0].Path)
	assert.Equal(t, "name", k.MergeKeys[0].Key)
}
