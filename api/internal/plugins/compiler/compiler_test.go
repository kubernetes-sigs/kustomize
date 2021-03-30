// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler_test

import (
	"fmt"
	"os"

	// "os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	. "sigs.k8s.io/kustomize/api/internal/plugins/compiler"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
)

// Regression coverage over compiler behavior.
func TestCompiler(t *testing.T) {
	srcRoot, err := utils.DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	for i, obj := range srcRoot {
		c := NewCompiler(obj)
		c.SetGVK("someteam.example.com", "v1", "DatePrefixer")
		err := matchOrError(c, obj, i, len(srcRoot), "someteam.example.com", "v1", "dateprefixer", "DatePrefixer.so")
		if err != nil {
			t.Error(err)
		}
		c.SetGVK("builtin", "", "SecretGenerator")
		err = matchOrError(c, obj, i, len(srcRoot), "builtin", "", "secretgenerator", "SecretGenerator.so")
		if err != nil {
			t.Error(err)
		}
	}
}

func matchOrError(c *Compiler, obj string, idx, lenOfSrcRoot int, gvk ...string) error {
	expectObj := filepath.Join(obj, strings.Join(gvk, string(os.PathSeparator)))
	if !strings.Contains(c.ObjPath(), expectObj) && idx == (lenOfSrcRoot-1) {
		return fmt.Errorf("Expected '%s', got '%s'", expectObj, c.ObjPath())
	}
	err := c.Compile()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if !utils.FileExists(expectObj) {
		return fmt.Errorf("didn't find expected obj file %s", expectObj)
	}
	return nil
}
