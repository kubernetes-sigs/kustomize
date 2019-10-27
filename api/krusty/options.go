// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty

// Options holds high-level configuration options, e.g.
// are plugins enabled, should the loader be restricted to
// the kustomization root, etc.
type Options struct {
	// When true, sort the resources before emitting them,
	// per a particular sort order.  When false, don't do the
	// sort, and instead respect the depth-first resource input
	// order as specified by the kustomization file(s).
	DoLegacyResourceSort bool
	// When true, the files referenced by a kustomization file
	// must be in or under the directory holding the kustomization
	// file itself.  When false, the kustomization file may specify
	// absolute or relative paths to patch or resources files outside
	// its own tree.
	RestrictToRootOnly bool
}

// MakeDefaultOptions returns a default instance of Options.
func MakeDefaultOptions() *Options {
	return &Options{
		DoLegacyResourceSort: true,
		RestrictToRootOnly:   true,
	}
}
