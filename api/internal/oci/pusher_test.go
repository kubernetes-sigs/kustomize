// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping because docker binary wasn't found in PATH")
	}
}

// Set up the registry.
func registry(t *testing.T) (*exec.Cmd, int, error) {
	skipIfNoDocker(t)

	container_name := fmt.Sprintf("%s_%d", t.Name(), 15)
	internal_port := 5000

	registry := exec.Command("docker", "run",
		"--rm",
		"-p", fmt.Sprintf("0:%d", internal_port),
		"--name", container_name,
		"docker.io/library/registry:3.0.0",
	)

	stdout, err := registry.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := registry.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { registry.Process.Signal(syscall.SIGTERM) })

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, fmt.Sprintf("listening on [::]:%d", internal_port)) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	retrieve_registry_port := exec.Command("docker", "port", container_name, fmt.Sprintf("%d", internal_port))
	out, err := retrieve_registry_port.Output()
	if err != nil {
		t.Fatal(err)
	}
	_, portString, err := net.SplitHostPort(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatal(err)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatal(err)
	}

	return registry, port, nil
}

func TestFnContainerTransformerWithConfig(t *testing.T) {
	registry, port, err := registry(t)
	require.NoError(t, err)
	// t.Cleanup(func() {registry.})
	require.NotNil(t, registry)
	t.Setenv("asdfsd", "asdfadsf")

	require.Equal(t, port, 5)

}
