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

package target_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
)

// TestEnvController manages the KustTarget test environment.
// It sets/resets XDG_CONFIG_HOME, makes/removes a temp objRoot.
type TestEnvController struct {
	t             *testing.T
	xdgConfigHome string
	oldXdg        string
	wasSet        bool
}

func NewTestEnvController(t *testing.T) *TestEnvController {
	return &TestEnvController{t: t}
}

func (x *TestEnvController) Set() *TestEnvController {
	x.makeTmpConfigHomeDir()
	x.makeObjectRootDir()
	x.setEnv()
	return x
}

func (x *TestEnvController) Reset() {
	x.resetEnv()
	x.removeTmpConfigHomeDir()
}

func (x *TestEnvController) fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (x *TestEnvController) recentFileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	age := time.Now().Sub(fi.ModTime())
	return age.Minutes() < 1
}

func (x *TestEnvController) BuildGoPlugin(plugin ...string) {
	obj := filepath.Join(
		append([]string{x.ObjectRoot()}, plugin...)...) + ".so"
	if x.recentFileExists(obj) {
		// Skip rebuilding it.
		return
	}
	src := filepath.Join(
		append([]string{x.SrcRoot()}, plugin...)...) + ".go"
	if !x.fileExists(src) {
		x.t.Errorf("cannot find go plugin source %s", src)
	}
	commands := []string{
		"build",
		"-buildmode",
		"plugin",
		"-tags=plugin",
		"-o", obj, src,
	}
	goBin := filepath.Join(os.Getenv("GOROOT"), "bin", "go")
	if !x.fileExists(src) {
		x.t.Errorf("cannot find go compiler %s", goBin)
	}
	cmd := exec.Command(goBin, commands...)
	cmd.Env = os.Environ()
	// cmd.Dir = filepath.Join(dir, "kustomize", "plugins")

	if err := cmd.Run(); err != nil {
		x.t.Errorf("compiler error building %s: %v", src, err)
	}
}

// ObjectRoot is the objRoot dir for plugin object files.
func (x *TestEnvController) ObjectRoot() string {
	return filepath.Join(
		x.xdgConfigHome, pgmconfig.PgmName, plugin.PluginRoot)
}

// SrcRoot is a objRoot directory for plugin source code
// used by tests.
//
// Plugin object code files have to be in a particular
// location to be found and loaded for security reasons,
// but placement of plugin source code is up to the user.
//
// This function returns a location for storing example
// plugins for tests.  And maybe builtins at some point.
func (x *TestEnvController) SrcRoot() string {
	dir := filepath.Join(
		os.Getenv("GOPATH"), "src",
		pgmconfig.Repo, pgmconfig.PgmName, plugin.PluginRoot)
	if _, err := os.Stat(dir); err != nil {
		x.t.Errorf("plugin source objRoot '%s' not found", dir)
	}
	return dir
}

func (x *TestEnvController) makeTmpConfigHomeDir() {
	var err error
	x.xdgConfigHome, err = ioutil.TempDir("", "kustomizetests")
	if err != nil {
		x.t.Errorf("failed to make temp objRoot: %v", err)
	}
}

func (x *TestEnvController) makeObjectRootDir() {
	err := os.MkdirAll(x.ObjectRoot(), os.ModePerm)
	if err != nil {
		x.t.Errorf(
			"making temp object objRoot %s: %v", x.ObjectRoot(), err)
	}
}

func (x *TestEnvController) removeTmpConfigHomeDir() {
	err := os.RemoveAll(x.xdgConfigHome)
	if err != nil {
		x.t.Errorf(
			"removing temp object objRoot: %s %v", x.xdgConfigHome, err)
	}
}

func (x *TestEnvController) setEnv() {
	x.oldXdg, x.wasSet = os.LookupEnv(pgmconfig.XDG_CONFIG_HOME)
	os.Setenv(pgmconfig.XDG_CONFIG_HOME, x.xdgConfigHome)
}

func (x *TestEnvController) resetEnv() {
	if x.wasSet {
		os.Setenv(pgmconfig.XDG_CONFIG_HOME, x.oldXdg)
	} else {
		os.Unsetenv(pgmconfig.XDG_CONFIG_HOME)
	}
}
