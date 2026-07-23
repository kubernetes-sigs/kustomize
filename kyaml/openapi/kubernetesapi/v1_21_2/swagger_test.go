// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v1_21_2 //nolint:revive // The package name is part of the public API.

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetCompatibility(t *testing.T) {
	b, err := Asset(assetName)
	require.NoError(t, err)
	digest := sha256.Sum256(b)
	require.Equal(t, "5d171b55e9601912807a870d73ffe70bb306f5889a00e76986042a0f2d7b6bc2",
		hex.EncodeToString(digest[:]))
	require.Equal(t, []string{assetName}, AssetNames())
	info, err := AssetInfo(assetName)
	require.NoError(t, err)
	require.Equal(t, int64(3469475), info.Size())
	require.Equal(t, assetName, info.Name())

	backslashName := "kubernetesapi\\v1_21_2\\swagger.pb"
	_, err = Asset(backslashName)
	require.NoError(t, err)
	children, err := AssetDir("kubernetesapi/v1_21_2")
	require.NoError(t, err)
	require.Equal(t, []string{"swagger.pb"}, children)
	_, err = AssetDir(assetName)
	require.Error(t, err)

	b[0] ^= 0xff
	again := MustAsset(assetName)
	require.NotEqual(t, b[0], again[0], "Asset must return a fresh byte slice")
	_, err = Asset("missing")
	require.Error(t, err)
}
