// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package konfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
	for _, expectedMsg := range []string{
		"unable to find plugin root - tried:",
		"('<no value>'; homed in $KUSTOMIZE_PLUGIN_HOME)",
		"; homed in $XDG_CONFIG_HOME)",
		"/.config/kustomize/plugin'; homed in default value of $XDG_CONFIG_HOME)",
		"/kustomize/plugin'; homed in home directory)",
	} {
		assert.Contains(t, err.Error(), expectedMsg)
	}
}

func TestDefaultAbsPluginHome_EmptyKustomizePluginHomeEnv(t *testing.T) {
	keep, isSet := os.LookupEnv(KustomizePluginHomeEnv)
	os.Setenv(KustomizePluginHomeEnv, "")

	_, err := DefaultAbsPluginHome(filesys.MakeFsInMemory())
	if !isSet {
		_ = os.Unsetenv(KustomizePluginHomeEnv)
	} else {
		_ = os.Setenv(KustomizePluginHomeEnv, keep)
	}
	if err == nil {
		t.Fatalf("expected err")
	}
	if !types.IsErrUnableToFind(err) {
		t.Fatalf("unexpected err: %v", err)
	}
	assert.Contains(t, err.Error(), "('<no value>'; homed in $KUSTOMIZE_PLUGIN_HOME)")
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

func TestDefaultAbsPluginHomeEmptyXdgConfig(t *testing.T) {
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	os.Setenv(XdgConfigHomeEnv, "")
	if isSet {
		_ = os.Unsetenv(XdgConfigHomeEnv)
	}
	_, err := DefaultAbsPluginHome(filesys.MakeFsInMemory())
	if isSet {
		os.Setenv(XdgConfigHomeEnv, keep)
	}
	if err == nil {
		t.Fatalf("expected err")
	}
	if !types.IsErrUnableToFind(err) {
		t.Fatalf("unexpected err: %v", err)
	}
	assert.Contains(t, err.Error(), "('<no value>'; homed in $XDG_CONFIG_HOME)")
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
