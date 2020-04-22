// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
)

func TestDeterminePluginSrcRoot(t *testing.T) {
	actual, err := DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	if !filepath.IsAbs(actual) {
		t.Errorf("expected absolute path, but got '%s'", actual)
	}
	expectedSuffix := filepath.Join("sigs.k8s.io", "kustomize", "plugin")
	if !strings.HasSuffix(actual, expectedSuffix) {
		t.Errorf("expected suffix '%s' in '%s'", expectedSuffix, actual)
	}
}
