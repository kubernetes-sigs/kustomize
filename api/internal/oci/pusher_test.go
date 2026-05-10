// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/distribution/reference"
	"github.com/stretchr/testify/require"

	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
	"sigs.k8s.io/kustomize/api/types"
)

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

	// Explicitly ignoring the certificates
	address, _ := createRegistry(t, username, password, true)
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
		http:    http.DefaultClient,
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")
}

func TestNoCredentialFile(t *testing.T) {
	username := "username"
	password := "password"

	address, caCert := createRegistry(t, username, password, true)

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
		http:    toClient(caCert),
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "basic credential not found")
}

func TestInvalidCredentials(t *testing.T) {
	address, caCert := createRegistry(t, "expectedusername", "expectedpassword", true)
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
		http:    toClient(caCert),
	}

	err := PushToOciRegistries(&pushOptions)
	require.ErrorContains(t, err, "response status code 401: Unauthorized")
}

func TestPush(t *testing.T) {
	username := "username"
	password := "password"

	address, caCert := createRegistry(t, username, password, true)
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
		http:    toClient(caCert),
	}

	err := PushToOciRegistries(&pushOptions)
	require.NoError(t, err)
}
