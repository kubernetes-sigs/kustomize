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

package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
)

// Compiler creates Go plugin object files.
//
// Source code is read from
//   ${srcRoot}/${g}/${v}/${k}.go
//
// Object code is written to
//   ${objRoot}/${g}/${v}/${k}.so
type Compiler struct {
	srcRoot string
	objRoot string
}

// DefaultSrcRoot guesses where the user
// has her ${g}/${v}/${k}.go files.
func DefaultSrcRoot() (string, error) {
	var nope []string
	var root string

	root = filepath.Join(
		os.Getenv("GOPATH"), "src",
		pgmconfig.DomainName,
		pgmconfig.ProgramName, plugin.PluginRoot)
	if FileExists(root) {
		return root, nil
	}
	nope = append(nope, root)

	_, filename, _, _ := runtime.Caller(1)
	root = path.Join(path.Dir(filename), "../..", plugin.PluginRoot)
	if FileExists(root) {
		return root, nil
	}
	nope = append(nope, root)

	root = filepath.Join(
		pgmconfig.ConfigRoot(), plugin.PluginRoot)
	if FileExists(root) {
		return root, nil
	}
	nope = append(nope, root)

	root = filepath.Join(
		pgmconfig.HomeDir(),
		pgmconfig.ProgramName, plugin.PluginRoot)
	if FileExists(root) {
		return root, nil
	}
	nope = append(nope, root)

	return "", fmt.Errorf(
		"no default src root; tried %v", nope)
}

// NewCompiler returns a new compiler instance.
func NewCompiler(srcRoot, objRoot string) *Compiler {
	return &Compiler{srcRoot: srcRoot, objRoot: objRoot}
}

// ObjRoot is root of compilation target tree.
func (b *Compiler) ObjRoot() string {
	return b.objRoot
}

func goBin() string {
	return filepath.Join(os.Getenv("GOROOT"), "bin", "go")
}

// Compile reads ${srcRoot}/${g}.go
//    and writes ${objRoot}/${g}.so
func (b *Compiler) Compile(g string) error {
	objDir := filepath.Join(b.objRoot)
	objFile := filepath.Join(objDir, g) + ".so"
	if RecentFileExists(objFile) {
		// Skip rebuilding it.
		return nil
	}
	err := os.MkdirAll(objDir, os.ModePerm)
	if err != nil {
		return err
	}
	srcFile := filepath.Join(b.srcRoot, g) + ".go"
	if !FileExists(srcFile) {
		return fmt.Errorf(
			"cannot find source %s", srcFile)
	}
	commands := []string{
		"build",
		"-buildmode",
		"plugin",
		"-tags=plugin",
		"-o", objFile, srcFile,
	}
	goBin := goBin()
	if !FileExists(goBin) {
		return fmt.Errorf(
			"cannot find go compiler %s", goBin)
	}
	cmd := exec.Command(goBin, commands...)
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"compiler error building %s: %v", srcFile, err)
	}
	return nil
}

// True if file less than 3 minutes old, i.e. not
// accidentally left over from some earlier build.
func RecentFileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	age := time.Now().Sub(fi.ModTime())
	return age.Minutes() < 3
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
