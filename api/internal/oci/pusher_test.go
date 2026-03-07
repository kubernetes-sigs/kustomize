// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/distribution/reference"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"sigs.k8s.io/kustomize/api/types"
)

type PusherTestSuite struct {
	suite.Suite
	registry              *testcontainers.DockerContainer
	address               string
	certificate_directory string
	ctx                   context.Context
}

func (suite *PusherTestSuite) SetupSuite() {
	certificate, key := generateSelfSignedCert(suite)

	const container_cert_path = "/certs/registry.pem"
	const container_key_path = "/certs/registry.key"
	registry_port := nat.Port("5000/tcp")

	suite.ctx = context.Background()
	container, err := testcontainers.Run(suite.ctx, "docker.io/library/registry:3.0.0",
		testcontainers.WithFiles(
			testcontainers.ContainerFile{
				HostFilePath:      certificate,
				ContainerFilePath: container_cert_path,
				FileMode:          0o644,
			},
			testcontainers.ContainerFile{
				HostFilePath:      key,
				ContainerFilePath: container_key_path,
				FileMode:          0o644,
			},
		),
		testcontainers.WithEnv(map[string]string{
			"REGISTRY_HTTP_TLS_CERTIFICATE": container_cert_path,
			"REGISTRY_HTTP_TLS_KEY":         container_key_path,
		}),
		testcontainers.WithExposedPorts(registry_port.Port()),
		testcontainers.WithWaitStrategy(
			wait.ForMappedPort(registry_port),
			wait.ForLog(fmt.Sprintf("listening on [::]:%s", registry_port.Port())),
		),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	ip, _ := container.Host(suite.ctx)
	port, _ := container.MappedPort(suite.ctx, registry_port)
	suite.address = fmt.Sprintf("https://%s:%s", ip, port.Port())
	suite.registry = container
}

func (suite *PusherTestSuite) TearDownSuite() {
	if suite.registry != nil {
		if err := suite.registry.Terminate(suite.ctx); err != nil {
			suite.FailNow("error terminating postgres container: %s", err)
		}
	}
	if suite.certificate_directory != "" {
		if err := os.RemoveAll(suite.certificate_directory); err != nil {
			suite.FailNow("error removing temp certificate directory: %s", err)
		}
	}
}

// Set up the registry.

func generateSelfSignedCert(suite *PusherTestSuite) (certificate string, key string) {
	directory, err := os.MkdirTemp("", "certs")
	if err != nil {
		suite.FailNow("failed to create temp directory: %s", err)
	}
	suite.certificate_directory = directory

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		suite.FailNow("failed to generate private key: %v", err)
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
		suite.FailNow("failed to create certificate: %v", err)
	}

	certificate_bytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	certificate_path := filepath.Join(directory, "registry.pem")
	os.WriteFile(certificate_path, certificate_bytes, 0o0640)

	key_bytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	key_path := filepath.Join(directory, "registry.key")
	os.WriteFile(key_path, key_bytes, 0o0640)

	return certificate_path, key_path
}

func AsNamedTagged(name string, tag string) reference.NamedTagged {
	named, _ := reference.WithName(name)
	tagged, _ := reference.WithTag(named, tag)
	return tagged
}

func (suite *PusherTestSuite) TestPushSetup() {
	t := suite.T()

	t.Setenv("asdfsd", "asdfadsf")

	require.Equal(t, suite.address, "something")
}

func TestPusherSuite(t *testing.T) {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	suite.Run(t, new(PusherTestSuite))
}

func TestPusherNeedsTargets(t *testing.T) {
	err := PushToOciRegistries(&PushOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "At least one target is required.")
}

func TestPusherNeedsNonNullKustomization(t *testing.T) {
	pushOptions := PushOptions{targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")}}

	err := PushToOciRegistries(&pushOptions)

	require.Error(t, err)
	require.Contains(t, err.Error(), "kustomization cannot be null")
}

func TestPusherNeedsNonEmptyKustomization(t *testing.T) {
	pushOptions := PushOptions{
		kustomization: &types.Kustomization{},
		targets:       []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kustomization.yaml is empty")
}

func TestPusherNeedsValidMetaIfSet(t *testing.T) {
	badData := map[string]types.TypeMeta{
		"nonempty_version": {
			APIVersion: "NonemptyVersion",
		},
		"invalid_kind": {
			Kind: "InvalidKind",
		},
		"invalid_version_for_kustomization_kind": {
			Kind:       types.KustomizationKind,
			APIVersion: "NonemptyVersion",
		},
		"invalid_version_for_compomenent_kind": {
			Kind:       types.ComponentKind,
			APIVersion: "NonemptyVersion",
		},
	}

	for name, testCase := range badData {
		t.Run(name, func(t *testing.T) {
			pushOptions := PushOptions{
				kustomization: &types.Kustomization{
					TypeMeta:  testCase,
					Namespace: "somethingnonempty",
				},
				targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
			}

			err := PushToOciRegistries(&pushOptions)
			require.Error(t, err)
			require.Contains(t, err.Error(), "kustomization has field errors")
		})
	}
}

func TestLogsDeprecatedFields(t *testing.T) {

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	pushOptions := PushOptions{
		kustomization: &types.Kustomization{
			Namespace:    "somethingnonempty",
			CommonLabels: map[string]string{"sdfsd": "sdfsf"},
			Vars:         []types.Var{{Name: "sdf"}},
		},
		targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Warning: 'commonLabels' is deprecated.")
	require.Contains(t, buf.String(), "Warning: 'vars' is deprecated.")
}

// func TestFnContainerTransformerWithConfig(t *testing.T) {

// 	kustomization := map[string]string{
// 		"src/README.md": `# NO VALID FILE
// `,
// 	}
// 	// clock := NewFakePassiveClock(time.Date(int(2025), time.July, int(28), int(20), int(56), int(0), int(0), time.UTC))

// 	_, _, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
// 	loctest.SetWorkingDir(t, target.Join("src"))

// 	registry, port, err := registry(t, certificate, key)
// 	require.NoError(t, err)

// 	// t.Cleanup(func() {registry.})
// 	require.NotNil(t, registry)
// 	t.Setenv("asdfsd", "asdfadsf")

// 	require.Equal(t, port, 7)
// }
