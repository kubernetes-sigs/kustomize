// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package konfig

import (
	"os"
	"path/filepath"
	"runtime"

	"sigs.k8s.io/kustomize/api/filesys"
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

	// Injected into plugin paths when plugins are disabled.
	// Provides a clue in flows that shouldn't happen.
	NoPluginHomeSentinal = "/No/non-builtin/plugins!"
)

type NotedFunc struct {
	Note string
	F    func() []string
}

// DefaultAbsPluginHome returns the absolute path in the given file
// system to first directory that looks like a good candidate for
// the home of kustomize plugins.
func DefaultAbsPluginHome(fSys filesys.FileSystem) ([]string, error) {
	return FirstDirThatExistsElseError(
		"plugin root", fSys, []NotedFunc{
			{
				Note: "homed in $" + KustomizePluginHomeEnv,
				F: func() []string {
					if os.Getenv(KustomizePluginHomeEnv) != "" {
						return filepath.SplitList(os.Getenv(KustomizePluginHomeEnv))
					}
					return []string{""}
				},
			},
			{
				Note: "homed in $" + XdgConfigHomeEnv,
				F: func() []string {
					if root := os.Getenv(XdgConfigHomeEnv); root != "" {
						return []string{filepath.Join(root, ProgramName, RelPluginHome)}
					}
					// do not look in "kustomize/plugin" if XdgConfigHomeEnv is unset
					return []string{""}
				},
			},
			{
				Note: "homed in default value of $" + XdgConfigHomeEnv,
				F: func() []string {
					return []string{filepath.Join(
						HomeDir(), XdgConfigHomeEnvDefault,
						ProgramName, RelPluginHome)}
				},
			},
			{
				Note: "homed in home directory",
				F: func() []string {
					return []string{filepath.Join(
						HomeDir(), ProgramName, RelPluginHome)}
				},
			},
			{
				Note: "homed in $" + XdgConfigDirs,
				F: func() []string {
					if os.Getenv(XdgConfigDirs) != "" {
						root := filepath.SplitList(os.Getenv(XdgConfigDirs))
						for i, dir := range root {
							root[i] = filepath.Join(dir, ProgramName, RelPluginHome)
						}
						return root
					}
					return []string{""}
				},
			},
			{
				Note: "homed in default value of $" + XdgConfigDirs,
				F: func() []string {
					root := filepath.SplitList(XdgConfigDirsDefault)
					for i, dir := range root {
						root[i] = filepath.Join(dir, ProgramName, RelPluginHome)
					}
					return root
				},
			},
		})
}

// FirstDirThatExistsElseError tests different path functions for
// existence, returning the first that works, else error if all fail.
func FirstDirThatExistsElseError(
	what string,
	fSys filesys.FileSystem,
	pathFuncs []NotedFunc) ([]string, error) {
	var result []string
	var nope []types.Pair
	for _, dt := range pathFuncs {
		for _, dir := range dt.F() {
			if dir != "" {
				if fSys.Exists(dir) {
					result = append(result, dir)
				}
				nope = append(nope, types.Pair{Key: dt.Note, Value: dir})
			} else {
				nope = append(nope, types.Pair{Key: dt.Note, Value: "<no value>"})
			}
		}
	}
	if result != nil {
		return result, nil
	}
	return result, types.NewErrUnableToFind(what, nope)
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
	return filesys.SelfDir
}

func pwdEnv() string {
	if runtime.GOOS == "windows" {
		return "CD"
	}
	return "PWD"
}
