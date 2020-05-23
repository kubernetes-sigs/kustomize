// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
)

func TestDeterminePluginSrcRoot(t *testing.T) {
	actual, err := DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	if !filepath.IsAbs(actual) {
		t.Errorf("expected absolute path, but got '%s'", actual)
	}
	if !strings.HasSuffix(actual, konfig.RelPluginHome) {
		t.Errorf("expected suffix '%s' in '%s'", konfig.RelPluginHome, actual)
	}
}
