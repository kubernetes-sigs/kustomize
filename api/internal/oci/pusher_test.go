// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
)

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping because docker binary wasn't found in PATH")
	}
}

// run calls Cmd.Run and wraps the error to include the output to make debugging
// easier. Not safe for real code, but fine for tests.
func run(cmd *exec.Cmd) error {
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n--- COMMAND OUTPUT ---\n%s", err, string(out))
	}
	return nil
}

// Set up the registry.
func registry(t *testing.T, certificate_path string, key_path string) (*exec.Cmd, int, error) {
	skipIfNoDocker(t)
	t.Helper()

	const container_cert_path = "/certs/cert.pem"
	const container_key_path = "/certs/key.pem"

	container_name := t.Name()
	internal_port := 5000

	create := exec.Command("docker", "create",
		"--rm",
		"--name", container_name,
		"--publish", fmt.Sprintf("0:%d", internal_port),
		"--env", "REGISTRY_HTTP_TLS_CERTIFICATE="+container_cert_path,
		"--env", "REGISTRY_HTTP_TLS_KEY="+container_key_path,
		"docker.io/library/registry:3.0.0",
	)
	require.NoError(t, run(create))

	cert_upload := exec.Command("docker", "cp", certificate_path, container_name+":"+container_cert_path)
	require.NoError(t, run(cert_upload))
	key_upload := exec.Command("docker", "cp", key_path, container_name+":"+container_key_path)
	require.NoError(t, run(key_upload))

	start := exec.Command("docker", "start",
		"--attach", "--interactive",
		container_name,
	)

	stdout, err := start.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := start.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { start.Process.Signal(syscall.SIGTERM) })

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

	return start, port, nil
}

func generateSelfSignedCert(t *testing.T) (certificate string, key string) {
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

	return string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		})),
		string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}))
}

func TestFnContainerTransformerWithConfig(t *testing.T) {
	certificate, key := generateSelfSignedCert(t)
	cert_path := "certs/cert.pem"
	key_path := "certs/key.pem"

	kustomization := map[string]string{
		"src/README.md": `# NO VALID FILE
`,
		cert_path: certificate,
		key_path:  key,
	}
	// clock := NewFakePassiveClock(time.Date(int(2025), time.July, int(28), int(20), int(56), int(0), int(0), time.UTC))

	_, _, target := loctest.PrepareFs(t, []string{"src", "certs"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	registry, port, err := registry(t, target.Join(cert_path), target.Join(key_path))
	require.NoError(t, err)

	// t.Cleanup(func() {registry.})
	require.NotNil(t, registry)
	t.Setenv("asdfsd", "asdfadsf")

	require.Equal(t, port, 7)

}
