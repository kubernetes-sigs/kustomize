// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// Image contains an image name, a new name, a new tag or digest,
// which will replace the original name and tag.
type Image struct {
	// Name is a tag-less image name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// NewName is the value used to replace the original name.
	NewName string `json:"newName,omitempty" yaml:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`

	// RegexpName is a regexp string match image name.
	RegexpName string `json:"regexpName,omitempty" yaml:"regexpName,omitempty"`

	// NewRegexpName is the value used to regexp replace the original name.
	NewRegexpName string `json:"newRegexpName,omitempty" yaml:"newRegexpName,omitempty"`
}
