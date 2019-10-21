// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "sigs.k8s.io/kustomize/api/plugins/compiler"
)

// Regression coverage over compiler behavior.
func TestCompiler(t *testing.T) {
	configRoot, err := ioutil.TempDir("", "kustomize-compiler-test")
	if err != nil {
		t.Errorf("failed to make temp dir: %v", err)
	}
	srcRoot, err := DefaultSrcRoot()
	if err != nil {
		t.Error(err)
	}
	c := NewCompiler(srcRoot, configRoot)
	if configRoot != c.ObjRoot() {
		t.Errorf("unexpected objRoot %s", c.ObjRoot())
	}

	expectObj := filepath.Join(
		c.ObjRoot(),
		"someteam.example.com", "v1", "dateprefixer", "DatePrefixer.so")
	if FileExists(expectObj) {
		t.Errorf("obj file should not exist yet: %s", expectObj)
	}
	err = c.Compile("someteam.example.com", "v1", "DatePrefixer")
	if err != nil {
		t.Error(err)
	}
	if !RecentFileExists(expectObj) {
		t.Errorf("didn't find expected obj file %s", expectObj)
	}

	expectObj = filepath.Join(
		c.ObjRoot(),
		"builtin", "", "secretgenerator", "SecretGenerator.so")
	if FileExists(expectObj) {
		t.Errorf("obj file should not exist yet: %s", expectObj)
	}
	err = c.Compile("builtin", "", "SecretGenerator")
	if err != nil {
		t.Error(err)
	}
	if !RecentFileExists(expectObj) {
		t.Errorf("didn't find expected obj file %s", expectObj)
	}

	err = os.RemoveAll(c.ObjRoot())
	if err != nil {
		t.Errorf(
			"removing temp dir: %s %v", c.ObjRoot(), err)
	}
	if FileExists(expectObj) {
		t.Errorf("cleanup failed; still see: %s", expectObj)
	}
}
