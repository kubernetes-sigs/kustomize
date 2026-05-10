// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cpuguy83/dockercfg"
	"github.com/mdelapenya/tlscert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

func withOptionalRegistryCertificate(useTls bool, cert *tlscert.Certificate) testcontainers.CustomizeRequestOption {
	if useTls {
		return func(req *testcontainers.GenericContainerRequest) error {
			const container_cert_path = "/certs/registry.pem"
			const container_key_path = "/certs/registry.key"
			req.Files = append(req.Files,
				testcontainers.ContainerFile{
					Reader:            bytes.NewReader(cert.Bytes),
					ContainerFilePath: container_cert_path,
					FileMode:          0o644,
				},
				testcontainers.ContainerFile{
					Reader:            bytes.NewReader(cert.KeyBytes),
					ContainerFilePath: container_key_path,
					FileMode:          0o644,
				})
			req.Env["REGISTRY_HTTP_TLS_CERTIFICATE"] = container_cert_path
			req.Env["REGISTRY_HTTP_TLS_KEY"] = container_key_path
			return nil
		}
	} else {
		return func(req *testcontainers.GenericContainerRequest) error { return nil }
	}
}

func withOptionalHtpasswd(username string, password string) testcontainers.CustomizeRequestOption {
	if username != "" && password != "" {
		return func(req *testcontainers.GenericContainerRequest) error {
			if hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err != nil {
				return err
			} else {
				entry := fmt.Sprintf("%s:%s", username, string(hash))
				return registry.WithHtpasswd(entry)(req)
			}
		}
	} else {
		return func(req *testcontainers.GenericContainerRequest) error { return nil }
	}
}

func toClient(cert *tlscert.Certificate) *http.Client {
	if cert == nil {
		return http.DefaultClient
	}

	pool := x509.NewCertPool()
	pool.AddCert(cert.Cert)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}
}

func createRegistry(t *testing.T, username string, password string, useTLS bool) (address string, cert *tlscert.Certificate) {
	ctx := context.Background()

	caCert, err := tlscert.SelfSignedE(strings.Join([]string{
		"hub.docker.internal", // So that the certificate can be used when the test process is running inside a container
		"localhost",
		"127.0.0.1",
	}, ","))
	if err != nil {
		t.Fatal(err)
	}

	registryContainer, err := registry.Run(ctx, "registry:3.0.0",
		withOptionalRegistryCertificate(useTLS, caCert),
		withOptionalHtpasswd(username, password),
		// redefine necessary for HTTPS/TLS
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").
				WithTLS(useTLS, caCert.TLSConfig()).
				WithPort("5000/tcp").
				WithStartupTimeout(10*time.Second),
		),
	)
	testcontainers.CleanupContainer(t, registryContainer)
	if err != nil {
		t.Fatal(err)
	}

	address, err = registryContainer.HostAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if useTLS {
		return address, caCert
	} else {
		return address, nil
	}
}

func createFileWithDirs(path string) (*os.File, error) {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return os.Create(path)
}

func createDockerConfig(t *testing.T, address string, user string, password string) {
	auth_directory := t.TempDir()
	auth_file, err := createFileWithDirs(filepath.Join(auth_directory, "config.json"))
	if err != nil {
		t.Fatal(err)
	}

	auth_config, err := registry.DockerAuthConfig(address, user, password)
	if err != nil {
		t.Fatal(err)
	}

	encoder := json.NewEncoder(auth_file)
	if err := encoder.Encode(dockercfg.Config{AuthConfigs: auth_config}); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DOCKER_CONFIG", auth_directory)
}
