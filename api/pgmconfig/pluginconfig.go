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

	// Name of environment variable used to set AbsPluginHome.
	// See that variable for an explanation.
	KustomizePluginHomeEnv = "KUSTOMIZE_PLUGIN_HOME"

	// Relative path below XDG_CONFIG_HOME/kustomize to find plugins.
	// e.g. AbsPluginHome = XDG_CONFIG_HOME/kustomize/plugin
	RelPluginHome = "plugin"

	// Location of builtin plugins below AbsPluginHome.
	BuiltinPluginPackage = "builtin"

	// The value of kubernetes ApiVersion to use in configuration
	// files for builtin plugins.
	// The value for non-builtins can be anything.
	BuiltinPluginApiVersion = BuiltinPluginPackage

	// Domain from which kustomize code is imported, for locating
	// plugin source code under $GOPATH when GOPATH is defined.
	DomainName = "sigs.k8s.io"
)

func EnabledPluginConfig() *types.PluginConfig {
	return MakePluginConfig(
		types.PluginRestrictionsNone, DefaultAbsPluginHome())
}

func DisabledPluginConfig() *types.PluginConfig {
	return MakePluginConfig(
		types.PluginRestrictionsBuiltinsOnly, NoPluginHomeSentinal)
}

func MakePluginConfig(
	pr types.PluginRestrictions, home string) *types.PluginConfig {
	return &types.PluginConfig{
		PluginRestrictions: pr,
		AbsPluginHome:      home,
	}
}

// Use an obviously erroneous path, in case it's accidentally used.
const NoPluginHomeSentinal = "/no/non-builtin/plugins!"

func DefaultAbsPluginHome() string {
	return filepath.Join(configRoot(), RelPluginHome)
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

func CurrentWorkingDir() string {
	// Try for full path first to be explicit.
	pwd := os.Getenv(pwdEnv())
	if len(pwd) > 0 {
		return pwd
	}
	return "."
}

func pwdEnv() string {
	if runtime.GOOS == "windows" {
		return "CD"
	}
	return "PWD"
}
