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
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpuguy83/dockercfg"
	"github.com/distribution/reference"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
	"sigs.k8s.io/kustomize/api/types"
)

type PusherTestSuite struct {
	suite.Suite
	registry       *registry.RegistryContainer
	user           string
	password       string
	address        string
	root_directory string
	ca_directory   string
	valid_auth     string
	ctx            context.Context
}

func withOptionalRegistryCertificate(key_path string, certificate_path string) testcontainers.CustomizeRequestOption {
	if key_path != "" && certificate_path != "" {
		return func(req *testcontainers.GenericContainerRequest) error {
			const container_cert_path = "/certs/registry.pem"
			const container_key_path = "/certs/registry.key"
			req.Files = append(req.Files,
				testcontainers.ContainerFile{
					HostFilePath:      certificate_path,
					ContainerFilePath: container_cert_path,
					FileMode:          0o644,
				},
				testcontainers.ContainerFile{
					HostFilePath:      key_path,
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

func createRegistry(t *testing.T, key_path string, cert_path string, username string, password string) string {
	ctx := context.Background()
	registryContainer, err := registry.Run(ctx, "registry:3.0.0",
		withOptionalRegistryCertificate(key_path, cert_path),
		withOptionalHtpasswd(username, password),
		// Necessary to work on podman/containerized podman
		testcontainers.WithWaitStrategy(
			wait.ForMappedPort("5000/tcp"),
		),
	)
	testcontainers.CleanupContainer(t, registryContainer)
	if err != nil {
		t.Fatal(err)
	}

	address, err := registryContainer.HostAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}

	return address
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

// Create a public/private key pair
func generateSelfSignedCert(t *testing.T) (certificate string, key string) {
	directory := t.TempDir()
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}

	key_path := filepath.Join(directory, "key", "registry.key")
	key_file, err := createFileWithDirs(key_path)
	if err != nil {
		t.Fatal(err)
	}

	if err := pem.Encode(key_file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		t.Fatal(err)
	}

	// Certificate template
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames: []string{
			"localhost",
			// So that the certificate can be used when the test process is running inside a container
			"hub.docker.internal",
		},
		BasicConstraintsValid: true,
	}

	// Self-sign the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	certificate_path := filepath.Join(directory, "cert", "registry.crt")
	certificate_file, err := createFileWithDirs(certificate_path)
	if err != nil {
		t.Fatal(err)
	}

	if err := pem.Encode(certificate_file, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}); err != nil {
		t.Fatal(err)
	}

	return certificate_path, key_path
}

func AsNamedTagged(name string, tag string) reference.NamedTagged {
	named, _ := reference.WithName(name)
	tagged, _ := reference.WithTag(named, tag)
	return tagged
}

func TestPusherNeedsTargets(t *testing.T) {
	err := PushToOciRegistries(&PushOptions{})
	require.ErrorContains(t, err, "At least one target is required.")
}

func TestPusherNeedsNonNullKustomization(t *testing.T) {
	pushOptions := PushOptions{targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")}}

	err := PushToOciRegistries(&pushOptions)

	require.ErrorContains(t, err, "kustomization cannot be null")
}

func TestPusherNeedsNonEmptyKustomization(t *testing.T) {
	pushOptions := PushOptions{
		kustomization: &types.Kustomization{},
		targets:       []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "kustomization.yaml is empty")
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
			require.ErrorContains(t, err, "kustomization has field errors")
		})
	}
}

func TestLogsDeprecatedFields(t *testing.T) {
	dummy, _, _ := loctest.PrepareFs(t, []string{}, map[string]string{})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	pushOptions := PushOptions{
		fSys: dummy,
		kustomization: &types.Kustomization{
			Namespace:    "somethingnonempty",
			CommonLabels: map[string]string{"sdfsd": "sdfsf"},
			Vars:         []types.Var{{Name: "sdf"}},
		},
		targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
	}

	_ = PushToOciRegistries(&pushOptions)
	require.Contains(t, buf.String(), "Warning: 'commonLabels' is deprecated.")
	require.Contains(t, buf.String(), "Warning: 'vars' is deprecated.")
}

func TestKustomizationFilePathsMustBeLocalToDirectory(t *testing.T) {
	fields := map[string]struct {
		fieldName string
		factory   func(string) types.Kustomization
	}{
		"components": {
			"Components",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Components: []string{p},
				}
			},
		},
		"resources": {
			"Resources",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Resources: []string{p},
				}
			},
		},
		"crds": {
			"Crds",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Crds: []string{p},
				}
			},
		},
		"configurations": {
			"Configurations",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Configurations: []string{p},
				}
			},
		},
		"generators": {
			"Generators",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Generators: []string{p},
				}
			},
		},
		"transformers": {
			"Transformers",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Transformers: []string{p},
				}
			},
		},
		"validators": {
			"Validators",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Validators: []string{p},
				}
			},
		},
		"patches": {
			"Patches",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Patches: []types.Patch{{Path: p}},
				}
			},
		},
		"replacements": {
			"Replacements",
			func(p string) types.Kustomization {
				return types.Kustomization{
					Replacements: []types.ReplacementField{{Path: p}},
				}
			},
		},
		"configMapGenerator files": {
			"ConfigMapGenerator",
			func(p string) types.Kustomization {
				return types.Kustomization{
					ConfigMapGenerator: []types.ConfigMapArgs{{GeneratorArgs: types.GeneratorArgs{KvPairSources: types.KvPairSources{FileSources: []string{p}}}}},
				}
			},
		},
		"configMapGenerator envs": {
			"ConfigMapGenerator",
			func(p string) types.Kustomization {
				return types.Kustomization{
					ConfigMapGenerator: []types.ConfigMapArgs{{GeneratorArgs: types.GeneratorArgs{KvPairSources: types.KvPairSources{EnvSources: []string{p}}}}},
				}
			},
		},
		"secretGenerator files": {
			"SecretGenerator",
			func(p string) types.Kustomization {
				return types.Kustomization{
					SecretGenerator: []types.SecretArgs{{GeneratorArgs: types.GeneratorArgs{KvPairSources: types.KvPairSources{FileSources: []string{p}}}}},
				}
			},
		},
		"SecretGenerator envs": {
			"SecretGenerator",
			func(p string) types.Kustomization {
				return types.Kustomization{
					SecretGenerator: []types.SecretArgs{{GeneratorArgs: types.GeneratorArgs{KvPairSources: types.KvPairSources{EnvSources: []string{p}}}}},
				}
			},
		},
		"helmCharts valuesFile": {
			"HelmCharts",
			func(p string) types.Kustomization {
				return types.Kustomization{
					HelmCharts: []types.HelmChart{{ValuesFile: p}},
				}
			},
		},
		"helmCharts additionalValuesFile": {
			"HelmCharts",
			func(p string) types.Kustomization {
				return types.Kustomization{
					HelmCharts: []types.HelmChart{{AdditionalValuesFiles: []string{p}}},
				}
			},
		},
	}
	paths := map[string]string{
		// "invalid fileurl": "file://asdfsd/something.txt",
		"parent directory": "..",
	}

	for fieldName, generator := range fields {

		for pathName, path := range paths {
			t.Run(fieldName+"|"+pathName, func(t *testing.T) {
				dummy, _, _ := loctest.PrepareFs(t, []string{}, map[string]string{})
				kustomization := generator.factory(path)

				pushOptions := PushOptions{
					fSys:          dummy,
					kustomization: &kustomization,
					targets:       []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
				}

				err := PushToOciRegistries(&pushOptions)
				require.ErrorContains(t, err, "kustomization includes non-local file paths")
				require.ErrorContains(t, err, fmt.Sprintf("Path '%s' in element %s is not local", path, generator.fieldName))
			})
		}
	}
}

func TestUntrustedCertificate(t *testing.T) {
	username := "username"
	password := "password"

	certificate, key := generateSelfSignedCert(t)
	// Explicitly _NOT_ setting the certificate directory
	address := createRegistry(t, key, certificate, username, password)

	kustomization := map[string]string{
		"src/kustomization.yaml": `namePrefix: test-
`,
	}

	_, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	pushOptions := PushOptions{
		fSys: actual,
		kustomization: &types.Kustomization{
			Namespace: "somethingnonempty",
		},
		targets: []reference.NamedTagged{AsNamedTagged(address+"/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")
}

func TestNoCredentialFile(t *testing.T) {
	username := "username"
	password := "password"

	certificate, key := generateSelfSignedCert(t)
	t.Setenv("SSL_CERT_DIR", filepath.Dir(certificate))

	address := createRegistry(t, key, certificate, username, password)

	kustomization := map[string]string{
		"src/kustomization.yaml": `namePrefix: test-
`,
	}

	_, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	pushOptions := PushOptions{
		fSys: actual,
		kustomization: &types.Kustomization{
			Namespace: "somethingnonempty",
		},
		targets: []reference.NamedTagged{AsNamedTagged(address+"/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "basic credential not found")
}

func TestInvalidCredentials(t *testing.T) {
	certificate, key := generateSelfSignedCert(t)
	t.Setenv("SSL_CERT_DIR", filepath.Dir(certificate))

	address := createRegistry(t, key, certificate, "expectedusername", "expectedpassword")
	createDockerConfig(t, address, "actualusername", "actualpassword")

	kustomization := map[string]string{
		"src/kustomization.yaml": `namePrefix: test-
`,
	}

	_, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	pushOptions := PushOptions{
		fSys: actual,
		kustomization: &types.Kustomization{
			Namespace: "somethingnonempty",
		},
		targets: []reference.NamedTagged{AsNamedTagged(address+"/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "response status code 401: Unauthorized")
}

func TestPush(t *testing.T) {
	username := "username"
	password := "password"

	certificate, key := generateSelfSignedCert(t)
	t.Setenv("SSL_CERT_DIR", filepath.Dir(certificate))

	address := createRegistry(t, key, certificate, username, password)
	createDockerConfig(t, address, username, password)

	kustomization := map[string]string{
		"src/kustomization.yaml": `namePrefix: test-
`,
	}

	_, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	pushOptions := PushOptions{
		fSys: actual,
		kustomization: &types.Kustomization{
			Namespace: "somethingnonempty",
		},
		targets: []reference.NamedTagged{AsNamedTagged(address+"/something", "sometag")},
	}

	err := PushToOciRegistries(&pushOptions)
	require.NoError(t, err)
}
