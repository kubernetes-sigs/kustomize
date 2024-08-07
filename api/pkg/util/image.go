// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"sigs.k8s.io/kustomize/api/internal/image"
)

// Splits image string name into name, tag and digest
func SplitImageName(imageName string) (name string, tag string, digest string) {
	return image.Split(imageName)
}
