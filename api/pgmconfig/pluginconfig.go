// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package pgmconfig

import (
	"os"
	"path/filepath"
	"runtime"

	"sigs.k8s.io/kustomize/api/types"
)

const (
	// Symbol that must be used inside Go plugins.
	PluginSymbol = "KustomizePlugin"

	// Location of builtins.
	BuiltinPluginPackage = "builtin"

	// The value of kubernetes ApiVersion to use in configuration
	// files for builtin plugins.
	// The value for non-builtins can be anything.
	BuiltinPluginApiVersion = BuiltinPluginPackage

	// Domain from which kustomize code is imported, for locating
	// plugin source code under $GOPATH when GOPATH is defined.
	DomainName = "sigs.k8s.io"

	// Name of directory housing all plugins.
	PluginRoot = "plugin"
)

func ActivePluginConfig() *types.PluginConfig {
	pc := DefaultPluginConfig()
	pc.PluginRestrictions = types.PluginRestrictionsNone
	return pc
}

func DefaultPluginConfig() *types.PluginConfig {
	return &types.PluginConfig{
		PluginRestrictions: types.PluginRestrictionsBuiltinsOnly,
		DirectoryPath:      filepath.Join(configRoot(), PluginRoot),
	}
}

// Use https://github.com/kirsle/configdir instead?
func configRoot() string {
	dir := os.Getenv(XdgConfigHomeEnv)
	if len(dir) == 0 {
		dir = filepath.Join(
			HomeDir(), XdgConfigHomeEnvDefault)
	}
	return filepath.Join(dir, ProgramName)
}

func HomeDir() string {
	home := os.Getenv(homeEnv())
	if len(home) > 0 {
		return home
	}
	return "~"
}

func homeEnv() string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE"
	}
	return "HOME"
}
