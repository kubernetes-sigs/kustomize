// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/oci/ocitest"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	ociScheme       = "oci://"
	ociRepo         = "/some-org/some-repo:v1.0.0"
	ociOverlay      = "//overlays/dev"
	resourcesPrefix = "resources:\n- "
)

// An artifact with a base and an overlay, as `flux push artifact` would build it.
func ociArtifactFiles() map[string]string {
	return map[string]string{
		"base/kustomization.yaml": `
resources:
- pod.yaml
`,
		"base/pod.yaml": `
apiVersion: v1
kind: Pod
metadata:
  name: myPod
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
`,
		"overlays/dev/kustomization.yaml": `
resources:
- ../../base
namePrefix: dev-
images:
- name: nginx
  newTag: 1.8.0
`,
	}
}

const ociBaseBuild = `apiVersion: v1
kind: Pod
metadata:
  name: myPod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`

const ociDevBuild = `apiVersion: v1
kind: Pod
metadata:
  name: dev-myPod
spec:
  containers:
  - image: nginx:1.8.0
    name: nginx
`

func TestOciArtifactBase(t *testing.T) {
	host := ocitest.StartRegistry(t)
	ref := host + ociRepo
	digest := ocitest.Push(t, ref, ocitest.FluxArtifact(ociArtifactFiles()))

	testcases := map[string]struct {
		resources string
		expected  string
	}{
		"base by tag": {
			resources: ociScheme + ref + "//base",
			expected:  ociBaseBuild,
		},
		"overlay by tag": {
			resources: ociScheme + ref + ociOverlay,
			expected:  ociDevBuild,
		},
		"overlay by digest": {
			resources: ociScheme + host + "/some-org/some-repo@" + digest + ociOverlay,
			expected:  ociDevBuild,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsOnDisk()
			th := kusttest_test.MakeHarnessWithFs(t, fSys)
			dir := t.TempDir()
			th.WriteK(dir, resourcesPrefix+tc.resources+"\n")

			m := th.Run(dir, th.MakeDefaultOptions())
			yml, err := m.AsYaml()
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(yml))
		})
	}
}

func TestOciArtifactAsTarget(t *testing.T) {
	host := ocitest.StartRegistry(t)
	ref := host + ociRepo
	ocitest.Push(t, ref, ocitest.Files(ociArtifactFiles()))

	th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
	m := th.Run(ociScheme+ref+ociOverlay, th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, ociDevBuild, string(yml))
}

func TestOciArtifactErrors(t *testing.T) {
	host := ocitest.StartRegistry(t)
	ref := host + ociRepo
	ocitest.Push(t, ref, ocitest.Files(ociArtifactFiles()))

	testcases := map[string]struct {
		resources string
		err       string
	}{
		"missing tag": {
			resources: ociScheme + host + "/some-org/some-repo:v2.0.0//base",
			err:       "pulling oci artifact",
		},
		"missing directory": {
			resources: ociScheme + ref + "//missing",
			err:       "missing: no such file or directory",
		},
		"file instead of directory": {
			resources: ociScheme + ref + "//base/pod.yaml",
			err:       "expecting directory",
		},
		"path exits artifact": {
			resources: ociScheme + ref + "//../escape",
			err:       "url path exits artifact",
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
			dir := t.TempDir()
			th.WriteK(dir, resourcesPrefix+tc.resources+"\n")

			err := th.RunWithErr(dir, th.MakeDefaultOptions())
			require.Contains(t, err.Error(), tc.err)
		})
	}
}

func TestOciArtifactCannotEscape(t *testing.T) {
	// An artifact whose overlay refers to a base outside the artifact.
	host := ocitest.StartRegistry(t)
	ref := host + "/some-org/escaping:v1.0.0"
	ocitest.Push(t, ref, ocitest.Files(map[string]string{
		// The artifact is pulled into a directory directly below the
		// temporary directory, so this is that temporary directory.
		"overlay/kustomization.yaml": "resources:\n- ../..\n",
	}))

	th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
	dir := t.TempDir()
	th.WriteK(dir, resourcesPrefix+ociScheme+ref+"//overlay\n")

	err := th.RunWithErr(dir, th.MakeDefaultOptions())
	require.Contains(t, err.Error(), "must be within the artifact")
}
