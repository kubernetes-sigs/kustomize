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
// It sets/resets XDG_CONFIG_HOME, makes/removes a temp objRoot,
// manages a plugin compiler, etc.
type PluginTestEnv struct {
	t        *testing.T
	compiler *compiler.Compiler
	workDir  string
	oldXdg   string
	wasSet   bool
}

func NewPluginTestEnv(t *testing.T) *PluginTestEnv {
	return &PluginTestEnv{t: t}
}

func (x *PluginTestEnv) Set() *PluginTestEnv {
	x.createWorkDir()
	x.compiler = x.makeCompiler()
	x.setEnv()
	return x
}

func (x *PluginTestEnv) Reset() {
	x.resetEnv()
	x.removeWorkDir()
}

func (x *PluginTestEnv) BuildGoPlugin(g, v, k string) {
	err := x.compiler.Compile(g, v, k)
	if err != nil {
		x.t.Errorf("compile failed: %v", err)
	}
}

func (x *PluginTestEnv) BuildExecPlugin(g, v, k string) {
	lowK := strings.ToLower(k)
	obj := filepath.Join(x.compiler.ObjRoot(), g, v, lowK, k)
	src := filepath.Join(x.compiler.SrcRoot(), g, v, lowK, k)
	if err := os.MkdirAll(filepath.Dir(obj), 0755); err != nil {
		x.t.Errorf("error making directory: %s", filepath.Dir(obj))
	}
	cmd := exec.Command("cp", src, obj)
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		x.t.Errorf("error copying %s to %s: %v", src, obj, err)
	}
}

func (x *PluginTestEnv) makeCompiler() *compiler.Compiler {
	// The plugin loader wants to find object code under
	//    $XDG_CONFIG_HOME/kustomize/plugins
	// and the compiler writes object code to
	//    $objRoot
	// so set things up accordingly.
	objRoot := filepath.Join(
		x.workDir, konfig.ProgramName, konfig.RelPluginHome)
	err := os.MkdirAll(objRoot, os.ModePerm)
	if err != nil {
		x.t.Error(err)
	}
	srcRoot, err := compiler.DefaultSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		x.t.Error(err)
	}
	return compiler.NewCompiler(srcRoot, objRoot)
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
	x.oldXdg, x.wasSet = os.LookupEnv(konfig.XdgConfigHomeEnv)
	os.Setenv(konfig.XdgConfigHomeEnv, x.workDir)
}

func (x *PluginTestEnv) resetEnv() {
	if x.wasSet {
		os.Setenv(konfig.XdgConfigHomeEnv, x.oldXdg)
	} else {
		os.Unsetenv(konfig.XdgConfigHomeEnv)
	}
}
