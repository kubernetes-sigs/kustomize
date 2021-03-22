// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// PluginConfig holds plugin configuration.
type PluginConfig struct {
	// PluginRestrictions distinguishes plugin restrictions.
	PluginRestrictions PluginRestrictions

	// BpLoadingOptions distinguishes builtin plugin behaviors.
	BpLoadingOptions BuiltinPluginLoadingOptions

	// FnpLoadingOptions sets the way function-based plugin behaviors.
	FnpLoadingOptions FnPluginLoadingOptions
}
