// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// Some plugin classes
// - builtin: plugins defined in the kustomize repo.
//   May be freely used and re-configured.
// - local: plugins that aren't builtin but are
//   locally defined (presumably by the user), meaning
//   the kustomization refers to them via a relative
//   file path, not a URL.
// - remote: require a build-time download to obtain.
//   Unadvised, unless one controls the
//   serving site.
//
//go:generate stringer -type=PluginRestrictions
type PluginRestrictions int

const (
	PluginRestrictionsUnknown PluginRestrictions = iota

	// Non-builtin plugins completely disabled.
	PluginRestrictionsBuiltinsOnly

	// No restrictions, do whatever you want.
	PluginRestrictionsNone
)

// BuiltinPluginLoadingOptions distinguish ways in which builtin plugins are used.
//go:generate stringer -type=BuiltinPluginLoadingOptions
type BuiltinPluginLoadingOptions int

const (
	BploUndefined BuiltinPluginLoadingOptions = iota

	// Desired in production use for performance.
	BploUseStaticallyLinked

	// Desired in testing and development cycles where it's undesirable
	// to generate static code.
	BploLoadFromFileSys
)
