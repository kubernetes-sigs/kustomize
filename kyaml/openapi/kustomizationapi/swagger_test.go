// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kustomizationapi

import (
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetCompatibility(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("swagger.json line endings may differ after a Windows checkout")
	}

	b, err := Asset(assetName)
	require.NoError(t, err)
	digest := sha256.Sum256(b)
	require.Equal(t, "65c5ffdbaa33a1a6db954c0e9524d5ae3ec5ef5af9d2ca17da0577868d887531",
		hex.EncodeToString(digest[:]))
	require.Equal(t, []string{assetName}, AssetNames())
	info, err := AssetInfo(assetName)
	require.NoError(t, err)
	require.Equal(t, int64(3182), info.Size())
	require.Equal(t, int64(len(b)), info.Size())
	require.Equal(t, assetName, info.Name())

	_, err = Asset("kustomizationapi\\swagger.json")
	require.NoError(t, err)
	children, err := AssetDir("kustomizationapi")
	require.NoError(t, err)
	require.Equal(t, []string{"swagger.json"}, children)
	_, err = AssetDir(assetName)
	require.Error(t, err)

	b[0] ^= 0xff
	again := MustAsset(assetName)
	require.NotEqual(t, b[0], again[0], "Asset must return a fresh byte slice")
	_, err = Asset("missing")
	require.Error(t, err)
}
