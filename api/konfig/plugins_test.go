// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package konfig

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/types"
)

func TestDefaultAbsPluginHome_NoKustomizePluginHomeEnv(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(KustomizePluginHomeEnv)
	if isSet {
		_ = os.Unsetenv(KustomizePluginHomeEnv)
	}
	_, err := DefaultAbsPluginHome(fSys)
	if isSet {
		os.Setenv(KustomizePluginHomeEnv, keep)
	}
	if err == nil {
		t.Fatalf("expected err")
	}
	if !types.IsErrUnableToFind(err) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDefaultAbsPluginHome_WithKustomizePluginHomeEnv(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(KustomizePluginHomeEnv)
	if !isSet {
		keep = "whatever"
		os.Setenv(KustomizePluginHomeEnv, keep)
	}
	fSys.Mkdir(keep)
	h, err := DefaultAbsPluginHome(fSys)
	if !isSet {
		_ = os.Unsetenv(KustomizePluginHomeEnv)
	}
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if h != keep {
		t.Fatalf("unexpected config dir: %s", h)
	}
}

func TestDefaultAbsPluginHomeWithXdg(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if !isSet {
		keep = "whatever"
		os.Setenv(XdgConfigHomeEnv, keep)
	}
	configDir := filepath.Join(keep, ProgramName, RelPluginHome)
	fSys.Mkdir(configDir)
	h, err := DefaultAbsPluginHome(fSys)
	if !isSet {
		_ = os.Unsetenv(XdgConfigHomeEnv)
	}
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if h != configDir {
		t.Fatalf("unexpected config dir: %s", h)
	}
}

func TestDefaultAbsPluginHomeNoConfig(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if isSet {
		_ = os.Unsetenv(XdgConfigHomeEnv)
	}
	_, err := DefaultAbsPluginHome(fSys)
	if isSet {
		os.Setenv(XdgConfigHomeEnv, keep)
	}
	if err == nil {
		t.Fatalf("expected err")
	}
	if !types.IsErrUnableToFind(err) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDefaultAbsPluginHomeNoXdgWithDotConfig(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	configDir := filepath.Join(
		HomeDir(), XdgConfigHomeEnvDefault, ProgramName, RelPluginHome)
	fSys.Mkdir(configDir)
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if isSet {
		_ = os.Unsetenv(XdgConfigHomeEnv)
	}
	s, _ := DefaultAbsPluginHome(fSys)
	if isSet {
		os.Setenv(XdgConfigHomeEnv, keep)
	}
	if s != configDir {
		t.Fatalf("unexpected config dir: %s", s)
	}
}

func TestDefaultAbsPluginHomeNoXdgJustHomeDir(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	configDir := filepath.Join(
		HomeDir(), ProgramName, RelPluginHome)
	fSys.Mkdir(configDir)
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if isSet {
		_ = os.Unsetenv(XdgConfigHomeEnv)
	}
	s, _ := DefaultAbsPluginHome(fSys)
	if isSet {
		os.Setenv(XdgConfigHomeEnv, keep)
	}
	if s != configDir {
		t.Fatalf("unexpected config dir: %s", s)
	}
}
