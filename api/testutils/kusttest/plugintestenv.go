// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"os"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/plugins/compiler"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
	"sigs.k8s.io/kustomize/api/konfig"
)

// pluginTestEnv manages compiling plugins for tests.
// It manages a Go plugin compiler, and sets/resets shell env vars as needed.
type pluginTestEnv struct {
	t          *testing.T
	compiler   *compiler.Compiler
	pluginRoot string
	oldXdg     string
	wasSet     bool
}

// newPluginTestEnv returns a new instance of pluginTestEnv.
func newPluginTestEnv(t *testing.T) *pluginTestEnv {
	return &pluginTestEnv{t: t}
}

// set creates a test environment.
// Uses a filesystem on disk for compilation (or copying) of
// plugin code - this FileSystem has nothing to do with
// the FileSystem used for loading config yaml in the tests.
func (x *pluginTestEnv) set() *pluginTestEnv {
	var err error
	x.pluginRoot, err = utils.DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		x.t.Error(err)
	}
	x.compiler = compiler.NewCompiler(x.pluginRoot)
	x.setEnv()
	return x
}

// reset restores the environment to pre-test state.
func (x *pluginTestEnv) reset() {
	// Calling Cleanup forces recompilation in a test file with multiple
	// calls to MakeEnhancedHarness - so leaving it out.  Your .gitignore
	// should ignore .so files anyway.
	// x.compiler.Cleanup()
	x.resetEnv()
}

// prepareGoPlugin compiles a Go plugin, leaving the newly
// created object code alongside the src code.
func (x *pluginTestEnv) prepareGoPlugin(g, v, k string) {
	x.compiler.SetGVK(g, v, k)
	err := x.compiler.Compile()
	if err != nil {
		x.t.Errorf("compile failed: %v", err)
	}
}

func (x *pluginTestEnv) prepareExecPlugin(_, _, _ string) {
	// Do nothing.  At one point this method
	// copied the exec plugin directory to a temp dir
	// and ran it from there.  Left as a hook.
}

func (x *pluginTestEnv) setEnv() {
	x.oldXdg, x.wasSet = os.LookupEnv(konfig.KustomizePluginHomeEnv)
	os.Setenv(konfig.KustomizePluginHomeEnv, x.pluginRoot)
}

func (x *pluginTestEnv) resetEnv() {
	if x.wasSet {
		os.Setenv(konfig.KustomizePluginHomeEnv, x.oldXdg)
	} else {
		os.Unsetenv(konfig.KustomizePluginHomeEnv)
	}
}
