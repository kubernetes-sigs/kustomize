// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/pgmconfig"
)

func TestConfigDirNoXdg(t *testing.T) {
	xdg, isSet := os.LookupEnv(pgmconfig.XdgConfigHome)
	if isSet {
		os.Unsetenv(pgmconfig.XdgConfigHome)
	}
	s := configRoot()
	if isSet {
		os.Setenv(pgmconfig.XdgConfigHome, xdg)
	}
	if !strings.HasSuffix(
		s,
		rootedPath(pgmconfig.DefaultConfigSubdir, pgmconfig.ProgramName)) {
		t.Fatalf("unexpected config dir: %s", s)
	}
}

func rootedPath(elem ...string) string {
	return string(filepath.Separator) + filepath.Join(elem...)
}

func TestConfigDirWithXdg(t *testing.T) {
	xdg, isSet := os.LookupEnv(pgmconfig.XdgConfigHome)
	os.Setenv(pgmconfig.XdgConfigHome, rootedPath("blah"))
	s := configRoot()
	if isSet {
		os.Setenv(pgmconfig.XdgConfigHome, xdg)
	}
	if s != rootedPath("blah", pgmconfig.ProgramName) {
		t.Fatalf("unexpected config dir: %s", s)
	}
}
