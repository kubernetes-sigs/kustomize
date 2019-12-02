// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/plugins/compiler"
	"sigs.k8s.io/kustomize/api/konfig"
)

// pluginTestEnv manages plugins for tests.
// It manages a Go plugin compiler,
// makes and removes a temporary working directory,
// and sets/resets shell env vars as needed.
type pluginTestEnv struct {
	t        *testing.T
	compiler *compiler.Compiler
	srcRoot  string
	workDir  string
	oldXdg   string
	wasSet   bool
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
	x.createWorkDir()
	var err error
	x.srcRoot, err = compiler.DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		x.t.Error(err)
	}
	x.compiler = compiler.NewCompiler(x.srcRoot, x.workDir)
	x.setEnv()
	return x
}

// reset restores the environment to pre-test state.
func (x *pluginTestEnv) reset() {
	x.resetEnv()
	x.removeWorkDir()
}

// buildGoPlugin compiles a Go plugin, leaving the newly
// created object code in the right place - a temporary
// working  directory pointed to by KustomizePluginHomeEnv.
// This avoids overwriting anything the user/developer has
// otherwise created.
func (x *pluginTestEnv) buildGoPlugin(g, v, k string) {
	err := x.compiler.Compile(g, v, k)
	if err != nil {
		x.t.Errorf("compile failed: %v", err)
	}
}

// prepExecPlugin copies an exec plugin from it's
// home in the discovered srcRoot to the same temp
// directory where Go plugin object code is placed.
// Kustomize (and its tests) expect to find plugins
// (Go or Exec) in the same spot, and since the test
// framework is compiling Go plugins to a temp dir,
// it must likewise copy Exec plugins to that same
// temp dir.
func (x *pluginTestEnv) prepExecPlugin(g, v, k string) {
	lowK := strings.ToLower(k)
	src := filepath.Join(x.srcRoot, g, v, lowK, k)
	tmp := filepath.Join(x.workDir, g, v, lowK, k)
	if err := os.MkdirAll(filepath.Dir(tmp), 0755); err != nil {
		x.t.Errorf("error making directory: %s", filepath.Dir(tmp))
	}
	cmd := exec.Command("cp", src, tmp)
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		x.t.Errorf("error copying %s to %s: %v", src, tmp, err)
	}
}

func (x *pluginTestEnv) createWorkDir() {
	var err error
	x.workDir, err = ioutil.TempDir("", "kustomize-plugin-tests")
	if err != nil {
		x.t.Errorf("failed to make work dir: %v", err)
	}
}

func (x *pluginTestEnv) removeWorkDir() {
	err := os.RemoveAll(x.workDir)
	if err != nil {
		x.t.Errorf(
			"removing work dir: %s %v", x.workDir, err)
	}
}

func (x *pluginTestEnv) setEnv() {
	x.oldXdg, x.wasSet = os.LookupEnv(konfig.KustomizePluginHomeEnv)
	os.Setenv(konfig.KustomizePluginHomeEnv, x.workDir)
}

func (x *pluginTestEnv) resetEnv() {
	if x.wasSet {
		os.Setenv(konfig.KustomizePluginHomeEnv, x.oldXdg)
	} else {
		os.Unsetenv(konfig.KustomizePluginHomeEnv)
	}
}
