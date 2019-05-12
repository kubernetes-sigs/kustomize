/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugintest_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/plugins"
)

// PluginTestEnv manages the plugin test environment.
// It sets/resets XDG_CONFIG_HOME, makes/removes a temp objRoot.
type PluginTestEnv struct {
	t        *testing.T
	compiler *plugins.Compiler
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

func (x *PluginTestEnv) BuildExecPlugin(name ...string) {
	obj := filepath.Join(
		append([]string{x.workDir, pgmconfig.ProgramName, pgmconfig.PluginRoot}, name...)...)

	srcRoot, err := plugins.DefaultSrcRoot()
	if err != nil {
		x.t.Error(err)
	}

	src := filepath.Join(
		append([]string{srcRoot}, name...)...)

	if err := os.MkdirAll(filepath.Dir(obj), 0755); err != nil {
		x.t.Errorf("error making directory: %s", filepath.Dir(obj))
	}
	cmd := exec.Command("cp", src, obj)
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		x.t.Errorf("error copying %s: %v", src, err)
	}
}

func (x *PluginTestEnv) makeCompiler() *plugins.Compiler {
	// The plugin loader wants to find object code under
	//    $XDG_CONFIG_HOME/kustomize/plugins
	// and the compiler writes object code to
	//    $objRoot
	// so set things up accordingly.
	objRoot := filepath.Join(
		x.workDir, pgmconfig.ProgramName, pgmconfig.PluginRoot)
	err := os.MkdirAll(objRoot, os.ModePerm)
	if err != nil {
		x.t.Error(err)
	}
	srcRoot, err := plugins.DefaultSrcRoot()
	if err != nil {
		x.t.Error(err)
	}
	return plugins.NewCompiler(srcRoot, objRoot)
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
	x.oldXdg, x.wasSet = os.LookupEnv(pgmconfig.XDG_CONFIG_HOME)
	os.Setenv(pgmconfig.XDG_CONFIG_HOME, x.workDir)
}

func (x *PluginTestEnv) resetEnv() {
	if x.wasSet {
		os.Setenv(pgmconfig.XDG_CONFIG_HOME, x.oldXdg)
	} else {
		os.Unsetenv(pgmconfig.XDG_CONFIG_HOME)
	}
}
