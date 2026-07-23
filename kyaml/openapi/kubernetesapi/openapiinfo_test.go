// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetesapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyMetadataRemainsStable(t *testing.T) {
	require.Equal(t, "v1.21.2", DefaultOpenAPI)
	require.Equal(t, "{title:Kubernetes,version:v1.21.2}", Info)
	require.Contains(t, OpenAPIMustAsset, DefaultOpenAPI)
}
