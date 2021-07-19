// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource

// Origin retains information about where resources in the output
// of `kustomize build` originated from
type Origin struct {
	// Path is the path to the resource, rooted from the directory upon
	// which `kustomize build` was invoked
	Path string

	// Repo is the remote repository that the resource originated from if it is
	// not from a local file
	Repo string

	// Ref is the ref of the remote repository that the resource originated from
	// if it is not from a local file
	Ref string
}

// Copy returns a copy of origin
func (origin *Origin) Copy() Origin {
	return *origin
}
