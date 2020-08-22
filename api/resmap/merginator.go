// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmap

import "sigs.k8s.io/kustomize/api/resource"

// Merginator merges resources.
type Merginator interface {
	// Merge creates a new ResMap by merging incoming resources.
	// Error if conflict found.
	Merge([]*resource.Resource) (ResMap, error)
}
