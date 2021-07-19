// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// BuildOptions is a group of booleans used to toggle different build options
type BuildOptions struct {
	// option to retain resource origin data as an annotation
	AnnoOrigin bool `json:"addAnnoOrigin" yaml:"addAnnoOrigin"`
}
