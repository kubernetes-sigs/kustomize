// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestAddConfiguration(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddConfiguration(fSys)

	if cmd == nil {
		t.Fatal("Expected cmd to not be nil")
	}

	if cmd.Use != "configuration" {
		t.Fatalf("Expected Use to be 'configuration', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("Expected Short to not be empty")
	}
}

func TestAddConfigurationHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile("config1.yaml", []byte("apiVersion: v1\nkind: Config"))
	require.NoError(t, err)
	err = fSys.WriteFile("config2.yaml", []byte("apiVersion: v1\nkind: Config"))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddConfiguration(fSys)
	args := []string{"config1.yaml", "config2.yaml"}
	require.NoError(t, cmd.RunE(cmd, args))

	content, err := testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)
	assert.Contains(t, string(content), "config1.yaml")
	assert.Contains(t, string(content), "config2.yaml")
}

func TestAddConfigurationDuplicate(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	err := fSys.WriteFile("config.yaml", []byte("apiVersion: v1\nkind: Config"))
	require.NoError(t, err)
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddConfiguration(fSys)

	// First addition
	args := []string{"config.yaml"}
	require.NoError(t, cmd.RunE(cmd, args))

	content, err := testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)
	assert.Contains(t, string(content), "config.yaml")

	// Second addition (should skip duplicate)
	require.NoError(t, cmd.RunE(cmd, args))

	content, err = testutils_test.ReadTestKustomization(fSys)
	require.NoError(t, err)

	// Count occurrences - should only appear once
	count := 0
	data := string(content)
	for i := 0; i < len(data); {
		if idx := len(data[i:]); idx > 0 {
			if len(data) >= i+11 {
				if data[i:i+11] == "config.yaml" {
					count++
					i += 11
				} else {
					i++
				}
			} else {
				break
			}
		}
	}
	assert.Equal(t, 1, count, "config.yaml should appear exactly once")
}
