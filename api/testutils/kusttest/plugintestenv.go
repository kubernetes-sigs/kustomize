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

// PluginTestEnv manages the plugin test environment.
// It manages a Go plugin compiler,
// makes and removes a temporary working directory,
// and sets/resets shell env vars as needed.
type PluginTestEnv struct {
	t        *testing.T
	compiler *compiler.Compiler
	srcRoot  string
	workDir  string
	oldXdg   string
	wasSet   bool
}

// NewPluginTestEnv returns a new instance of PluginTestEnv.
func NewPluginTestEnv(t *testing.T) *PluginTestEnv {
	return &PluginTestEnv{t: t}
}

// Set creates a test environment.
func (x *PluginTestEnv) Set() *PluginTestEnv {
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

// Reset restores the environment to pre-test state.
func (x *PluginTestEnv) Reset() {
	x.resetEnv()
	x.removeWorkDir()
}

// WorkDir allows inspection of the temp working directory.
func (x *PluginTestEnv) WorkDir() string {
	return x.workDir
}

// BuildGoPlugin compiles a Go plugin, leaving the newly
// created object code in the right place - a temporary
// working  directory pointed to by KustomizePluginHomeEnv.
// This avoids overwriting anything the user/developer has
// otherwise created.
func (x *PluginTestEnv) BuildGoPlugin(g, v, k string) {
	err := x.compiler.Compile(g, v, k)
	if err != nil {
		x.t.Errorf("compile failed: %v", err)
	}
}

// PrepExecPlugin copies an exec plugin from it's
// home in the discovered srcRoot to the same temp
// directory where Go plugin object code is placed.
// Kustomize (and its tests) expect to find plugins
// (Go or Exec) in the same spot, and since the test
// framework is compiling Go plugins to a temp dir,
// it must likewise copy Exec plugins to that same
// temp dir.
func (x *PluginTestEnv) PrepExecPlugin(g, v, k string) {
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

func (x *PluginTestEnv) createWorkDir() {
	var err error
	x.workDir, err = ioutil.TempDir("", "kustomize-plugin-tests")
	if err != nil {
		x.t.Errorf("failed to make work dir: %v", err)
	}
}

func (x *PluginTestEnv) removeWorkDir() {
	err := os.RemoveAll(x.workDir)
	if err != nil {
		x.t.Errorf(
			"removing work dir: %s %v", x.workDir, err)
	}
}

func (x *PluginTestEnv) setEnv() {
	x.oldXdg, x.wasSet = os.LookupEnv(konfig.KustomizePluginHomeEnv)
	os.Setenv(konfig.KustomizePluginHomeEnv, x.workDir)
}

func (x *PluginTestEnv) resetEnv() {
	if x.wasSet {
		os.Setenv(konfig.KustomizePluginHomeEnv, x.oldXdg)
	} else {
		os.Unsetenv(konfig.KustomizePluginHomeEnv)
	}
}
