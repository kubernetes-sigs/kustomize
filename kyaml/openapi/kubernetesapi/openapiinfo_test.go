// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetesapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestLegacyMetadataMatchesBuiltinBundle(t *testing.T) {
	require.Equal(t, openapi.DefaultOpenAPI, DefaultOpenAPI)
	require.Equal(t, openapi.BuiltinSchemaInfo, Info)
	require.Contains(t, OpenAPIMustAsset, DefaultOpenAPI)
}
