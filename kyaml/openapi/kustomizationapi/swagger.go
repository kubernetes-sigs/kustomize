// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package kustomizationapi exposes the embedded Kustomization OpenAPI schema.
package kustomizationapi

import (
	_ "embed"
	"os"
	"time"

	"sigs.k8s.io/kustomize/kyaml/openapi/internal/embeddedasset"
)

const assetName = "kustomizationapi/swagger.json"

//go:embed swagger.json
var swaggerJSON []byte

var asset = embeddedasset.File{ //nolint:gochecknoglobals
	Name:    assetName,
	Data:    swaggerJSON,
	Size:    3182,
	Mode:    0o644,
	ModTime: time.Unix(1615228558, 0),
}

// Asset loads and returns the asset for the given name.
func Asset(name string) ([]byte, error) {
	return embeddedasset.Asset(asset, name) //nolint:wrapcheck // Preserve legacy error text.
}

// MustAsset is like Asset but panics when Asset would return an error.
func MustAsset(name string) []byte { return embeddedasset.MustAsset(asset, name) }

// AssetInfo loads and returns the asset info for the given name.
func AssetInfo(name string) (os.FileInfo, error) {
	return embeddedasset.AssetInfo(asset, name) //nolint:wrapcheck // Preserve legacy error text.
}

// AssetNames returns the names of the assets.
func AssetNames() []string { return []string{assetName} }

// AssetDir returns the file names below an embedded directory.
func AssetDir(name string) ([]string, error) {
	return embeddedasset.AssetDir(asset, name) //nolint:wrapcheck // Preserve legacy error text.
}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	return embeddedasset.RestoreAsset(asset, dir, name) //nolint:wrapcheck // Preserve the compatibility API.
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	return embeddedasset.RestoreAssets(asset, dir, name) //nolint:wrapcheck // Preserve the compatibility API.
}
