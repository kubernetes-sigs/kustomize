// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
)

func copyFileIntoContainer(ctx context.Context, cli *client.Client, containerID, containerPath string, data []byte) error {
	// Create tar archive in memory
	buffer := new(bytes.Buffer)
	writer := tar.NewWriter(buffer)
	header := &tar.Header{
		Name: containerPath,
		Mode: 0644,
		Size: int64(len(data)),
	}

	if err := writer.WriteHeader(header); err != nil {
		return err
	}

	if _, err := writer.Write(data); err != nil {
		return err
	}

	writer.Close()

	_, err := cli.CopyToContainer(ctx, containerID, client.CopyToContainerOptions{
		DestinationPath: "/",
		Content:         buffer,
	})

	return err
}

// Set up the registry.

func registry(t *testing.T, certificate []byte, key []byte) (containerId string, port int, err error) {
	t.Helper()

	const container_cert_path = "/certs/cert.pem"
	const container_key_path = "/certs/key.pem"

	container_name := t.Name()
	internal_port := 5000

	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { apiClient.Close() })

	portMap := network.MustParsePort(fmt.Sprintf("%d/tcp", internal_port))

	container, err := apiClient.ContainerCreate(t.Context(), client.ContainerCreateOptions{
		Name: container_name,
		Config: &container.Config{
			Image:        "docker.io/library/registry:3.0.0",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			Env: []string{
				"REGISTRY_HTTP_TLS_CERTIFICATE=" + container_cert_path,
				"REGISTRY_HTTP_TLS_KEY=" + container_key_path,
			},
		},
		HostConfig: &container.HostConfig{
			AutoRemove: true,
			PortBindings: network.PortMap{
				portMap: []network.PortBinding{
					{
						HostPort: "",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { apiClient.ContainerStop(context.Background(), container.ID, client.ContainerStopOptions{}) })

	if err = copyFileIntoContainer(t.Context(), apiClient, container.ID, container_cert_path, []byte(certificate)); err != nil {
		t.Fatal(err)
	}
	if err = copyFileIntoContainer(t.Context(), apiClient, container.ID, container_key_path, []byte(key)); err != nil {
		t.Fatal(err)
	}

	if _, err = apiClient.ContainerStart(t.Context(), container.ID, client.ContainerStartOptions{}); err != nil {
		t.Fatal(err)
	}

	reader, err := apiClient.ContainerLogs(t.Context(), container.ID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, fmt.Sprintf("listening on [::]:%d", internal_port)) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	inspect, err := apiClient.ContainerInspect(t.Context(), container.ID, client.ContainerInspectOptions{})
	if err != nil {
		t.Fatal(err)
	}

	bindings := inspect.Container.NetworkSettings.Ports[portMap]
	if len(bindings) == 0 {
		t.Fatal("No ports bound")
	}

	port, err = strconv.Atoi(bindings[0].HostPort)
	if err != nil {
		t.Fatal(err)
	}

	return container.ID, port, nil
}

func generateSelfSignedCert(t *testing.T) (certificate []byte, key []byte) {
	t.Helper()

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Certificate template
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Self-sign the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		}),
		pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
}

func TestFnContainerTransformerWithConfig(t *testing.T) {
	certificate, key := generateSelfSignedCert(t)

	kustomization := map[string]string{
		"src/README.md": `# NO VALID FILE
`,
	}
	// clock := NewFakePassiveClock(time.Date(int(2025), time.July, int(28), int(20), int(56), int(0), int(0), time.UTC))

	_, _, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	registry, port, err := registry(t, certificate, key)
	require.NoError(t, err)

	// t.Cleanup(func() {registry.})
	require.NotNil(t, registry)
	t.Setenv("asdfsd", "asdfadsf")

	require.Equal(t, port, 7)
}
