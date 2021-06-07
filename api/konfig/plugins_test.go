// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package konfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/types"
)

func TestDefaultAbsPluginHome_NoKustomizePluginHomeEnv(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(KustomizePluginHomeEnv)
	if isSet {
		unsetenv(t, KustomizePluginHomeEnv)
	}
	_, err := DefaultAbsPluginHome(fSys)
	if isSet {
		setenv(t, KustomizePluginHomeEnv, keep)
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
	setenv(t, KustomizePluginHomeEnv, "")

	_, err := DefaultAbsPluginHome(filesys.MakeFsInMemory())
	if !isSet {
		unsetenv(t, KustomizePluginHomeEnv)
	} else {
		setenv(t, KustomizePluginHomeEnv, keep)
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
		setenv(t, KustomizePluginHomeEnv, keep)
	}
	err := fSys.Mkdir(keep)
	require.NoError(t, err)
	h, err := DefaultAbsPluginHome(fSys)
	if !isSet {
		unsetenv(t, KustomizePluginHomeEnv)
	}
	require.NoError(t, err)
	if h != keep {
		t.Fatalf("unexpected config dir: %s", h)
	}
}

func TestDefaultAbsPluginHomeWithXdg(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if !isSet {
		keep = "whatever"
		setenv(t, XdgConfigHomeEnv, keep)
	}
	configDir := filepath.Join(keep, ProgramName, RelPluginHome)
	err := fSys.Mkdir(configDir)
	require.NoError(t, err)
	h, err := DefaultAbsPluginHome(fSys)
	if !isSet {
		unsetenv(t, XdgConfigHomeEnv)
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
		unsetenv(t, XdgConfigHomeEnv)
	}
	_, err := DefaultAbsPluginHome(fSys)
	if isSet {
		setenv(t, XdgConfigHomeEnv, keep)
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
	setenv(t, XdgConfigHomeEnv, "")
	if isSet {
		unsetenv(t, XdgConfigHomeEnv)
	}
	_, err := DefaultAbsPluginHome(filesys.MakeFsInMemory())
	if isSet {
		setenv(t, XdgConfigHomeEnv, keep)
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
	err := fSys.Mkdir(configDir)
	require.NoError(t, err)
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if isSet {
		unsetenv(t, XdgConfigHomeEnv)
	}
	s, err := DefaultAbsPluginHome(fSys)
	require.NoError(t, err)
	if isSet {
		setenv(t, XdgConfigHomeEnv, keep)
	}
	if s != configDir {
		t.Fatalf("unexpected config dir: %s", s)
	}
}

func TestDefaultAbsPluginHomeNoXdgJustHomeDir(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	configDir := filepath.Join(
		HomeDir(), ProgramName, RelPluginHome)
	err := fSys.Mkdir(configDir)
	require.NoError(t, err)
	keep, isSet := os.LookupEnv(XdgConfigHomeEnv)
	if isSet {
		unsetenv(t, XdgConfigHomeEnv)
	}
	s, err := DefaultAbsPluginHome(fSys)
	require.NoError(t, err)
	if isSet {
		setenv(t, XdgConfigHomeEnv, keep)
	}
	if s != configDir {
		t.Fatalf("unexpected config dir: %s", s)
	}
}

func setenv(t *testing.T, key, value string) {
	require.NoError(t, os.Setenv(key, value))
}

func unsetenv(t *testing.T, key string) {
	require.NoError(t, os.Unsetenv(key))
}
