// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const stubOutputEnv = "KUSTOMIZE_TEST_KUBECTL_OUTPUT"

func stubKubectl(t *testing.T, output string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub kubectl is a shell script")
	}
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s' \"$" + stubOutputEnv + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "kubectl"), []byte(script), 0o700)) //nolint:gosec
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv(stubOutputEnv, output)
}

func runFetch(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := NewCmdFetch(&buf)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestFetchInvalidJSON(t *testing.T) {
	stubKubectl(t, "<html><body>login required</body></html>")

	out, err := runFetch(t)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse the schema returned by kubectl")
	require.Empty(t, out)
}

func TestFetchValidJSON(t *testing.T) {
	schema := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "Kubernetes", "version": "v1.30.0"},
	}
	raw, err := json.Marshal(schema)
	require.NoError(t, err)
	stubKubectl(t, string(raw))

	out, err := runFetch(t)
	require.NoError(t, err)
	expected, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)
	require.Equal(t, string(expected)+"\n", out)
}

func TestFetchValidJSONAsYAML(t *testing.T) {
	schema := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "Kubernetes", "version": "v1.30.0"},
	}
	raw, err := json.Marshal(schema)
	require.NoError(t, err)
	stubKubectl(t, string(raw))

	out, err := runFetch(t, "--format=yaml")
	require.NoError(t, err)
	var parsed map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(out), &parsed))
	require.Equal(t, schema, parsed)
}
