// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	depproviderapi "sigs.k8s.io/kustomize/api/internal/provider"
)

// NewDefaultDepProvider provides a new Dependency Provider
// that contains a new hasher and validator.
func NewDefaultDepProvider() *depproviderapi.DepProvider {
	return depproviderapi.NewDepProvider()
}
