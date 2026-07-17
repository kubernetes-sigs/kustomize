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

func TestReadInputWithLimit(t *testing.T) {
	const limit = int64(4)
	testDir := t.TempDir()

	write := func(t *testing.T, name string, contents []byte) string {
		t.Helper()
		path := filepath.Join(testDir, name)
		require.NoError(t, os.WriteFile(path, contents, 0o600))
		return path
	}
	gzipContents := func(t *testing.T, contents []byte) []byte {
		t.Helper()
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		_, err := writer.Write(contents)
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		return buffer.Bytes()
	}

	t.Run("raw at limit", func(t *testing.T) {
		path := write(t, "raw", []byte("data"))
		got, err := readInputWithLimit(path, limit)
		require.NoError(t, err)
		require.Equal(t, []byte("data"), got)
	})

	t.Run("raw exceeds limit", func(t *testing.T) {
		path := write(t, "raw-large", []byte("large"))
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "input")
		require.ErrorContains(t, err, "exceeds 4 bytes")
	})

	t.Run("gzip at limit", func(t *testing.T) {
		path := write(t, "gzip", gzipContents(t, []byte("data")))
		got, err := readInputWithLimit(path, limit)
		require.NoError(t, err)
		require.Equal(t, []byte("data"), got)
	})

	t.Run("gzip exceeds limit", func(t *testing.T) {
		path := write(t, "gzip-large", gzipContents(t, []byte("large")))
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "decompressed input")
		require.ErrorContains(t, err, "exceeds 4 bytes")
	})

	t.Run("invalid gzip header", func(t *testing.T) {
		path := write(t, "gzip-invalid-header", []byte{0x1f, 0x8b})
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "open gzip input")
	})

	t.Run("invalid gzip body", func(t *testing.T) {
		contents := gzipContents(t, []byte("data"))
		contents[len(contents)-1]++
		path := write(t, "gzip-invalid-body", contents)
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "read decompressed input")
	})
}
