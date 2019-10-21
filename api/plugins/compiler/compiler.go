// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/api/pgmconfig"
	"sigs.k8s.io/kustomize/api/plugins/config"
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
// has her ${g}/${v}/$lower(${k})/${k}.go files.
func DefaultSrcRoot() (string, error) {
	var nope []string
	var root string

	root = filepath.Join(
		os.Getenv("GOPATH"), "src",
		config.DomainName, pgmconfig.ProgramName, config.PluginRoot)
	if FileExists(root) {
		return root, nil
	}
	nope = append(nope, root)

	root = config.DefaultPluginConfig().DirectoryPath
	if FileExists(root) {
		return root, nil
	}
	nope = append(nope, root)

	root = filepath.Join(
		config.HomeDir(), pgmconfig.ProgramName, config.PluginRoot)
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

// SrcRoot is where to find src.
func (b *Compiler) SrcRoot() string {
	return b.srcRoot
}

func goBin() string {
	return filepath.Join(runtime.GOROOT(), "bin", "go")
}

// Compile reads ${srcRoot}/${g}/${v}/${k}.go
//    and writes ${objRoot}/${g}/${v}/${k}.so
func (b *Compiler) Compile(g, v, k string) error {
	lowK := strings.ToLower(k)
	objDir := filepath.Join(b.objRoot, g, v, lowK)
	objFile := filepath.Join(objDir, k) + ".so"
	if RecentFileExists(objFile) {
		// Skip rebuilding it.
		return nil
	}
	err := os.MkdirAll(objDir, os.ModePerm)
	if err != nil {
		return err
	}
	srcFile := filepath.Join(b.srcRoot, g, v, lowK, k) + ".go"
	if !FileExists(srcFile) {
		// Handy for tests of lone plugins.
		s := k + ".go"
		if !FileExists(s) {
			return fmt.Errorf(
				"cannot find source at '%s' or '%s'", srcFile, s)

		}
		srcFile = s
	}
	commands := []string{
		"build",
		"-buildmode",
		"plugin",
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
