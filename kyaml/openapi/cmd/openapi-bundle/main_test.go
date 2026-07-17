// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
)

func TestGeneratedBundleIsCurrentAndDeterministic(t *testing.T) {
	source := filepath.Join("..", "..", "kubernetesapi", "v1_21_2", "swagger.pb.gz")
	checkedIn := filepath.Join("..", "..", "kubernetesapi", "data",
		"kubernetes-openapi-union-v1.21.2.bundle-v1.json.gz")
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "first.json.gz")
	second := filepath.Join(tempDir, "second.json.gz")
	legacy := filepath.Join(tempDir, "swagger.pb.gz")

	for i, output := range []string{first, second} {
		legacyOutput := ""
		if i == 0 {
			legacyOutput = legacy
		}
		require.NoError(t, run(options{
			input:             source,
			output:            output,
			legacyProtoOutput: legacyOutput,
			kubernetesVersion: "v1.21.2",
		}))
	}

	want, err := os.ReadFile(checkedIn)
	require.NoError(t, err)
	got, err := os.ReadFile(first)
	require.NoError(t, err)
	again, err := os.ReadFile(second)
	require.NoError(t, err)
	require.Equal(t, want, got, "checked-in bundle is stale")
	require.Equal(t, got, again, "bundle generation is not deterministic")
	sourceArchive, err := os.ReadFile(source)
	require.NoError(t, err)
	legacyArchive, err := os.ReadFile(legacy)
	require.NoError(t, err)
	require.Equal(t, sourceArchive, legacyArchive, "compiler input archive is not deterministic")

	reader, err := gzip.NewReader(bytes.NewReader(got))
	require.NoError(t, err)
	require.True(t, reader.ModTime.IsZero())
	require.Empty(t, reader.Name)
	require.Empty(t, reader.Comment)
	decoder := json.NewDecoder(reader)
	var bundle builtinopenapi.Bundle
	require.NoError(t, decoder.Decode(&bundle))
	var trailing interface{}
	require.ErrorIs(t, decoder.Decode(&trailing), io.EOF)
	require.NoError(t, reader.Close())
	require.NoError(t, bundle.Validate())
	require.Len(t, bundle.Definitions, 618)
	require.Len(t, bundle.Resources, 275)
	require.Equal(t, "5d171b55e9601912807a870d73ffe70bb306f5889a00e76986042a0f2d7b6bc2",
		bundle.Sources[0].SHA256)
}

func TestWriteGzipUsesStableHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.gz")
	require.NoError(t, writeGzip(path, []byte("data")))
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	reader, err := gzip.NewReader(bytes.NewReader(b))
	require.NoError(t, err)
	require.Equal(t, time.Time{}, reader.ModTime)
	require.Empty(t, reader.Name)
	require.Empty(t, reader.Comment)
	require.Equal(t, byte(255), reader.OS)
	require.NoError(t, reader.Close())
}
