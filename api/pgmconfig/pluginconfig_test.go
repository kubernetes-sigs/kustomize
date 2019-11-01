// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package pgmconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDirNoXdg(t *testing.T) {
	xdg, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if isSet {
		os.Unsetenv(XdgConfigHomeEnv)
	}
	s := configRoot()
	if isSet {
		os.Setenv(XdgConfigHomeEnv, xdg)
	}
	if !strings.HasSuffix(
		s,
		rootedPath(XdgConfigHomeEnvDefault, ProgramName)) {
		t.Fatalf("unexpected config dir: %s", s)
	}
}

func rootedPath(elem ...string) string {
	return string(filepath.Separator) + filepath.Join(elem...)
}

func TestConfigDirWithXdg(t *testing.T) {
	xdg, isSet := os.LookupEnv(XdgConfigHomeEnv)
	os.Setenv(XdgConfigHomeEnv, rootedPath("blah"))
	s := configRoot()
	if isSet {
		os.Setenv(XdgConfigHomeEnv, xdg)
	}
	if s != rootedPath("blah", ProgramName) {
		t.Fatalf("unexpected config dir: %s", s)
	}
}
