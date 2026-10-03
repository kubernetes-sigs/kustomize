// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package kubernetesapi preserves access to the legacy built-in Kubernetes
// OpenAPI assets. New code should use the parent openapi package.
package kubernetesapi

import (
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi/v1_21_2"
)

// Info describes the Kubernetes release represented by the legacy asset map.
const Info = "{title:Kubernetes,version:v1.21.2}"

// OpenAPIMustAsset maps supported Kubernetes releases to their legacy asset
// loaders.
var OpenAPIMustAsset = map[string]func(string) []byte{ //nolint:gochecknoglobals // Retained for API compatibility.
	"v1.21.2": v1_21_2.MustAsset,
}

// DefaultOpenAPI is the default release exposed by the legacy asset map.
const DefaultOpenAPI = "v1.21.2"
