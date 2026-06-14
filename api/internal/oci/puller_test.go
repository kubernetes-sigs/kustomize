// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"net/http"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/require"

	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestPullerNeedsTargets(t *testing.T) {
	err := PullUsingOciManifest(&RepoSpec{}, filesys.MakeEmptyDirInMemory(), nil)
	require.ErrorContains(t, err, "At least one target is required.")
}

func TestPullerNeesSomething(t *testing.T) {
	repoSpec := RepoSpec{Reference: name.MustParseReference("localhost:3030/somerepo/someimage:sometag")}

	err := PullUsingOciManifest(&repoSpec, filesys.MakeEmptyDirInMemory(), nil)

	require.ErrorContains(t, err, "image must exist")
}

// func TestPusherNeedsNonEmptyKustomization(t *testing.T) {
// 	pushOptions := PushOptions{
// 		kustomization: &types.Kustomization{},
// 		targets:       []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
// 	}

// 	err := PushToOciRegistries(&pushOptions)
// 	require.ErrorContains(t, err, "kustomization.yaml is empty")
// }

// func TestPusherNeedsValidMetaIfSet(t *testing.T) {
// 	badData := map[string]types.TypeMeta{
// 		"nonempty_version": {
// 			APIVersion: "NonemptyVersion",
// 		},
// 		"invalid_kind": {
// 			Kind: "InvalidKind",
// 		},
// 		"invalid_version_for_kustomization_kind": {
// 			Kind:       types.KustomizationKind,
// 			APIVersion: "NonemptyVersion",
// 		},
// 		"invalid_version_for_compomenent_kind": {
// 			Kind:       types.ComponentKind,
// 			APIVersion: "NonemptyVersion",
// 		},
// 	}

// 	for name, testCase := range badData {
// 		t.Run(name, func(t *testing.T) {
// 			pushOptions := PushOptions{
// 				kustomization: &types.Kustomization{
// 					TypeMeta:  testCase,
// 					Namespace: "somethingnonempty",
// 				},
// 				targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
// 			}

// 			err := PushToOciRegistries(&pushOptions)
// 			require.ErrorContains(t, err, "kustomization has field errors")
// 		})
// 	}
// }

// func TestLogsDeprecatedFields(t *testing.T) {
// 	dummy, _, _ := loctest.PrepareFs(t, []string{}, map[string]string{})

// 	var buf bytes.Buffer
// 	log.SetOutput(&buf)
// 	defer func() {
// 		log.SetOutput(os.Stderr)
// 	}()

// 	pushOptions := PushOptions{
// 		fSys: dummy,
// 		kustomization: &types.Kustomization{
// 			Namespace:    "somethingnonempty",
// 			CommonLabels: map[string]string{"sdfsd": "sdfsf"},
// 			Vars:         []types.Var{{Name: "sdf"}},
// 		},
// 		targets: []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
// 	}

// 	_ = PushToOciRegistries(&pushOptions)
// 	require.Contains(t, buf.String(), "Warning: 'commonLabels' is deprecated.")
// 	require.Contains(t, buf.String(), "Warning: 'vars' is deprecated.")
// }

// func TestPullerKustomizationFilePathsMustBeLocalToDirectory(t *testing.T) {
// 	fields := map[string]struct {
// 		fieldName string
// 		factory   func(string) types.Kustomization
// 	}{
// 		"components": {
// 			"Components",
// 			func(p string) types.Kustomization {
// 				return types.Kustomization{
// 					Components: []string{p},
// 				}
// 			},
// 		},
// 		"resources": {
// 			"Resources",
// 			func(p string) types.Kustomization {
// 				return types.Kustomization{
// 					Resources: []string{p},
// 				}
// 			},
// 		},
// 	}
// 	paths := map[string]string{
// 		// "invalid fileurl": "file://asdfsd/something.txt",
// 		"parent directory": "..",
// 	}

// 	for fieldName, generator := range fields {

// 		for pathName, path := range paths {
// 			t.Run(fieldName+"|"+pathName, func(t *testing.T) {
// 				dummy, _, _ := loctest.PrepareFs(t, []string{}, map[string]string{})
// 				kustomization := generator.factory(path)

// 				pushOptions := PushOptions{
// 					fSys:          dummy,
// 					kustomization: &kustomization,
// 					targets:       []reference.NamedTagged{AsNamedTagged("registry.domain/something", "sometag")},
// 				}

// 				err := PushToOciRegistries(&pushOptions)
// 				require.ErrorContains(t, err, "kustomization includes non-local file paths")
// 				require.ErrorContains(t, err, fmt.Sprintf("Path '%s' in element %s is not local", path, generator.fieldName))
// 			})
// 		}
// 	}
// }

func TestPullerUntrustedCertificate(t *testing.T) {
	username := "username"
	password := "password"

	// Explicitly ignoring the certificates
	address, _ := createRegistry(t, username, password, true)

	repoSpec := RepoSpec{Reference: toReference(t, address+"/somerepo/someimage:sometag")}

	err := PullUsingOciManifest(&repoSpec, filesys.MakeEmptyDirInMemory(), http.DefaultClient)

	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")
}

func TestPullerMissingImageNoCredentialFileNoPassword(t *testing.T) {

	address, _ := createRegistry(t, "", "", false)

	repoSpec := RepoSpec{Reference: toReference(t, address+"/somerepo/someimage:sometag", name.Insecure)}

	err := PullUsingOciManifest(&repoSpec, filesys.MakeEmptyDirInMemory(), http.DefaultClient)
	require.ErrorContains(t, err, "basic credential not found")
}

func TestPullerNoCredentialFile(t *testing.T) {
	username := "username"
	password := "password"

	address, caCert := createRegistry(t, username, password, true)

	repoSpec := RepoSpec{Reference: toReference(t, address+"/somerepo/someimage:sometag")}

	err := PullUsingOciManifest(&repoSpec, filesys.MakeEmptyDirInMemory(), toClient(caCert))
	require.ErrorContains(t, err, "basic credential not found")
}

func TestPullerInvalidCredentials(t *testing.T) {
	address, caCert := createRegistry(t, "expectedusername", "expectedpassword", true)
	createDockerConfig(t, address, "actualusername", "actualpassword")

	repoSpec := RepoSpec{Reference: toReference(t, address+"/somerepo/someimage:sometag")}

	err := PullUsingOciManifest(&repoSpec, filesys.MakeEmptyDirInMemory(), toClient(caCert))
	require.ErrorContains(t, err, "response status code 401: Unauthorized")
}

func TestPullNoPassword(t *testing.T) {

	t.Setenv("TESTCONTAINERS_HOST_OVERRIDE", "host.docker.internal")??????

	address, _ := createRegistry(t, "", "", false)
	createDockerConfig(t, address, "", "")

	kustomization := map[string]string{
		"kustomization.yaml": `namePrefix: test-
`,
	}

	src, actual, target := loctest.PrepareFs(t, nil, kustomization)
	reference := toReference(t, address+"/somerepo:sometag")
	pushArtifact(t, src, target.String(), reference, "", "", nil)

	repoSpec := RepoSpec{
		Reference: reference,
		Dir:       target,
	}

	err := PullUsingOciManifest(&repoSpec, actual, http.DefaultClient)
	require.NoError(t, err)
	loctest.CheckFs(t, target.String(), src, actual)
}

func TestPullNoCertificateNoPassword(t *testing.T) {

	address, _ := createRegistry(t, "", "", false)
	createDockerConfig(t, address, "", "")

	kustomization := map[string]string{
		"kustomization.yaml": `namePrefix: test-
`,
	}

	src, actual, target := loctest.PrepareFs(t, nil, kustomization)
	reference := toReference(t, address+"/somerepo:sometag")
	pushArtifact(t, src, target.String(), reference, "", "", nil)

	repoSpec := RepoSpec{
		Reference: reference,
		Dir:       target,
	}

	err := PullUsingOciManifest(&repoSpec, actual, http.DefaultClient)
	require.NoError(t, err)
	loctest.CheckFs(t, target.String(), src, actual)
}

func TestPullNoCertificate(t *testing.T) {
	username := "username"
	password := "password"

	address, _ := createRegistry(t, username, password, false)
	createDockerConfig(t, address, username, password)

	kustomization := map[string]string{
		"kustomization.yaml": `namePrefix: test-
`,
	}

	src, actual, target := loctest.PrepareFs(t, nil, kustomization)
	reference := toReference(t, address+"/somerepo/someimage:sometag")
	pushArtifact(t, src, target.String(), reference, username, password, nil)

	repoSpec := RepoSpec{
		Reference: reference,
		Dir:       target,
	}

	err := PullUsingOciManifest(&repoSpec, actual, http.DefaultClient)
	require.NoError(t, err)
	loctest.CheckFs(t, target.String(), src, actual)
}

func TestPull(t *testing.T) {
	username := "username"
	password := "password"

	address, caCert := createRegistry(t, username, password, true)
	createDockerConfig(t, address, username, password)

	kustomization := map[string]string{
		"kustomization.yaml": `namePrefix: test-
`,
	}

	src, actual, target := loctest.PrepareFs(t, nil, kustomization)
	reference := toReference(t, address+"/somerepo/someimage:sometag")
	pushArtifact(t, src, target.String(), reference, username, password, caCert)

	repoSpec := RepoSpec{
		Reference: reference,
		Dir:       target,
	}

	err := PullUsingOciManifest(&repoSpec, actual, toClient(caCert))
	require.NoError(t, err)
	loctest.CheckFs(t, target.String(), src, actual)
}
