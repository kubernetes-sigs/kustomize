// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
)

// Compiler creates Go plugin object files.
type Compiler struct {
	// pluginRoot is where the user
	// has her ${g}/${v}/$lower(${k})/${k}.go files.
	pluginRoot string
	// Where compilation happens.
	workDir string
	// Used as the root file name for src and object.
	rawKind string
	// Capture compiler output.
	stderr bytes.Buffer
	// Capture compiler output.
	stdout bytes.Buffer
}

// NewCompiler returns a new compiler instance.
func NewCompiler(root string) *Compiler {
	return &Compiler{pluginRoot: root}
}

// Set GVK converts g,v,k tuples to file path components.
func (b *Compiler) SetGVK(g, v, k string) {
	b.rawKind = k
	b.workDir = filepath.Join(b.pluginRoot, g, v, strings.ToLower(k))
}

func (b *Compiler) srcPath() string {
	return filepath.Join(b.workDir, b.rawKind+".go")
}

func (b *Compiler) objFile() string {
	return b.rawKind + ".so"
}

// Absolute path to the compiler output (the .so file).
func (b *Compiler) ObjPath() string {
	return filepath.Join(b.workDir, b.objFile())
}

// Compile changes its working directory to
// ${pluginRoot}/${g}/${v}/$lower(${k} and places
// object code next to source code.
func (b *Compiler) Compile() error {
	if !utils.FileExists(b.srcPath()) {
		return fmt.Errorf("cannot find source at '%s'", b.srcPath())
	}
	// If you use an IDE, make sure it's go build and test flags
	// match those used below.  Same goes for Makefile targets.
	commands := []string{
		"build",
		// "-trimpath",  This flag used to make it better, now it makes it worse,
		//               see https://github.com/golang/go/issues/31354
		"-buildmode",
		"plugin",
		"-o", b.objFile(),
	}
	goBin := utils.GoBin()
	if !utils.FileExists(goBin) {
		return fmt.Errorf(
			"cannot find go compiler %s", goBin)
	}
	cmd := exec.Command(goBin, commands...)
	b.stderr.Reset()
	cmd.Stderr = &b.stderr
	b.stdout.Reset()
	cmd.Stdout = &b.stdout
	cmd.Env = os.Environ()
	cmd.Dir = b.workDir
	if err := cmd.Run(); err != nil {
		b.report()
		return errors.Wrapf(
			err, "cannot compile %s:\nSTDERR\n%s\n",
			b.srcPath(), b.stderr.String())
	}
	result := filepath.Join(b.workDir, b.objFile())
	if utils.FileExists(result) {
		log.Printf("compiler created: %s", result)
		return nil
	}
	return fmt.Errorf("post compile, cannot find '%s'", result)
}

func (b *Compiler) report() {
	log.Println("stdout:  -------")
	log.Println(b.stdout.String())
	log.Println("----------------")
	log.Println("stderr:  -------")
	log.Println(b.stderr.String())
	log.Println("----------------")
}
