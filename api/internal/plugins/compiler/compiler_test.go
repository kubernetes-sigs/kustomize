// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler_test

import (
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/kustomize/api/filesys"
	. "sigs.k8s.io/kustomize/api/internal/plugins/compiler"
)

// Regression coverage over compiler behavior.
func TestCompiler(t *testing.T) {
	srcRoot, err := DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	c := NewCompiler(srcRoot)

	c.SetGVK("someteam.example.com", "v1", "DatePrefixer")
	expectObj := filepath.Join(
		srcRoot, "someteam.example.com", "v1", "dateprefixer", "DatePrefixer.so")
	if expectObj != c.ObjPath() {
		t.Errorf("Expected '%s', got '%s'", expectObj, c.ObjPath())
	}
	err = c.Compile()
	if err != nil {
		t.Error(err)
	}
	if !FileYoungerThan(expectObj, time.Second) {
		t.Errorf("didn't find expected obj file %s", expectObj)
	}
	c.Cleanup()
	if FileExists(expectObj) {
		t.Errorf("obj file '%s' should be gone", expectObj)
	}

	c.SetGVK("builtin", "", "SecretGenerator")
	expectObj = filepath.Join(
		srcRoot,
		"builtin", "", "secretgenerator", "SecretGenerator.so")
	if expectObj != c.ObjPath() {
		t.Errorf("Expected '%s', got '%s'", expectObj, c.ObjPath())
	}
	err = c.Compile()
	if err != nil {
		t.Error(err)
	}
	if !FileYoungerThan(expectObj, time.Second) {
		t.Errorf("didn't find expected obj file %s", expectObj)
	}
	c.Cleanup()
	if FileExists(expectObj) {
		t.Errorf("obj file '%s' should be gone", expectObj)
	}
}
