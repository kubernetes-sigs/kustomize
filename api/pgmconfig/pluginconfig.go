// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package pgmconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/pflag"
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

	flagEnablePluginsName = "enable_alpha_plugins"
	flagEnablePluginsHelp = `enable plugins, an alpha feature.
See https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins/README.md
`
	flagErrorFmt = `
unable to load external plugin %s because plugins disabled
specify the flag
  --%s
to %s`
)

func ActivePluginConfig() *types.PluginConfig {
	pc := DefaultPluginConfig()
	pc.Enabled = true
	return pc
}

func DefaultPluginConfig() *types.PluginConfig {
	return &types.PluginConfig{
		Enabled:       false,
		DirectoryPath: filepath.Join(configRoot(), PluginRoot),
	}
}

func NotEnabledErr(name string) error {
	return fmt.Errorf(
		flagErrorFmt,
		name,
		flagEnablePluginsName,
		flagEnablePluginsHelp)
}

func AddFlagEnablePlugins(set *pflag.FlagSet, v *bool) {
	set.BoolVar(
		v, flagEnablePluginsName,
		false, flagEnablePluginsHelp)
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
