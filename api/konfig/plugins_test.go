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

const tempDir = "whatever"

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
	if !isSet || keep == "" {
		keep = tempDir
		os.Setenv(KustomizePluginHomeEnv, keep)
	}
	dirs := filepath.SplitList(keep)
	for _, dir := range dirs {
		fSys.Mkdir(dir)
	}
	h, err := DefaultAbsPluginHome(fSys)
	if !isSet {
		unsetenv(t, KustomizePluginHomeEnv)
	}
	require.NoError(t, err)
	for i, dir := range h {
		for _, pDir := range dirs {
			if dir != pDir && i == len(h)-1 {
				t.Fatalf("unexpected config dirs: %v", h)
			} else if dir == pDir {
				return
			}
		}
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
	for i, dir := range h {
		if dir != configDir && i == len(h)-1 {
			t.Fatalf("unexpected config dirs: %v", h)
		}
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
	for i, dir := range s {
		if dir != configDir && i == len(s)-1 {
			t.Fatalf("unexpected config dirs: %v", s)
		}
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
	for i, dir := range s {
		if dir != configDir && i == len(s)-1 {
			t.Fatalf("unexpected config dirs: %v", s)
		}
	}
}

func setenv(t *testing.T, key, value string) {
	require.NoError(t, os.Setenv(key, value))
}

func unsetenv(t *testing.T, key string) {
	require.NoError(t, os.Unsetenv(key))
}

func TestDefaultAbsPluginHomeWithXdgConfigDirs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(XdgConfigDirs)
	if !isSet || keep == "" {
		keep = tempDir
		os.Setenv(XdgConfigDirs, keep)
	}
	dirs := filepath.SplitList(keep)
	var configDirs []string
	for _, dir := range dirs {
		configDir := filepath.Join(dir, ProgramName, RelPluginHome)
		fSys.Mkdir(configDir)
		configDirs = append(configDirs, configDir)
	}
	h, err := DefaultAbsPluginHome(fSys)
	if !isSet {
		_ = os.Unsetenv(XdgConfigDirs)
	}
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	for i, dir := range h {
		for _, cDir := range configDirs {
			if dir != cDir && i == len(h)-1 {
				t.Fatalf("unexpected config dirs: %v", h)
			} else if dir == cDir {
				return
			}
		}
	}
}

func TestDefaultAbsPluginHomeNoXdgConfigDirs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	keep, isSet := os.LookupEnv(XdgConfigDirs)
	if isSet {
		_ = os.Unsetenv(XdgConfigDirs)
	}
	_, err := DefaultAbsPluginHome(fSys)
	if isSet {
		os.Setenv(XdgConfigDirs, keep)
	}
	if err == nil {
		t.Fatalf("expected err")
	}
	if !types.IsErrUnableToFind(err) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDefaultAbsPluginHomeEmptyXdgConfigDirs(t *testing.T) {
	keep, isSet := os.LookupEnv(XdgConfigDirs)
	os.Setenv(XdgConfigDirs, "")
	if isSet {
		_ = os.Unsetenv(XdgConfigDirs)
	}
	_, err := DefaultAbsPluginHome(filesys.MakeFsInMemory())
	if isSet {
		os.Setenv(XdgConfigDirs, keep)
	}
	if err == nil {
		t.Fatalf("expected err")
	}
	if !types.IsErrUnableToFind(err) {
		t.Fatalf("unexpected err: %v", err)
	}
	assert.Contains(t, err.Error(), "('<no value>'; homed in $XDG_CONFIG_DIRS)")
}
