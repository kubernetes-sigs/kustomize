// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
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

// Compile reads ${srcRoot}/${g}/${v}/${k}.go
//    and writes ${objRoot}/${g}/${v}/${k}.so
func (b *Compiler) Compile(g, v, k string) error {
	lowK := strings.ToLower(k)
	objDir := filepath.Join(b.objRoot, g, v, lowK)
	objFile := filepath.Join(objDir, k) + ".so"
	if FileYoungerThan(objFile, time.Minute) {
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(
			err, "cannot compile %s:\nSTDERR\n%s\n", srcFile, stderr.String())
	}
	return nil
}
