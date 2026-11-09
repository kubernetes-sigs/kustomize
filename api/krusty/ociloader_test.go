// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/oci/ocitest"
	"sigs.k8s.io/kustomize/api/krusty/localizer"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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

// Builds pull through the default HTTP client, so these tests use a registry
// on the loopback interface, which go-containerregistry reaches over plain HTTP.

func TestOciArtifactBase(t *testing.T) {
	host := ocitest.StartLoopbackRegistry(t)
	ref := host + ociRepo
	digest := ocitest.Push(t, nil, ref, ocitest.FluxArtifact(t, ociArtifactFiles()))

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

func TestOciArtifactPulledOncePerBuild(t *testing.T) {
	// The build refers to the tag twice, under different root paths, and
	// the tag moves to another artifact once pulled: the build pulls the
	// tag once, and both references see the same artifact.
	moved := ociArtifactFiles()
	moved["overlays/dev/kustomization.yaml"] = strings.Replace(
		moved["overlays/dev/kustomization.yaml"], "dev-", "moved-", 1)
	h := ocitest.Handler()
	var pulls atomic.Int32
	host := ocitest.StartLoopbackServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/manifests/v1.0.0") &&
			pulls.Add(1) > 1 {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "v1.0.0") + "moved"
		}
		h.ServeHTTP(w, r)
	}))
	ocitest.Push(t, nil, host+ociRepo, ocitest.Files(t, ociArtifactFiles()))
	ocitest.Push(t, nil, host+"/some-org/some-repo:moved", ocitest.Files(t, moved))
	pulls.Store(0)

	th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
	dir := t.TempDir()
	th.WriteK(dir, resourcesPrefix+ociScheme+host+ociRepo+"//base\n- "+ociScheme+host+ociRepo+ociOverlay+"\n")

	m := th.Run(dir, th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, ociBaseBuild+"---\n"+ociDevBuild, string(yml))
	require.Equal(t, int32(1), pulls.Load())
}

func TestOciArtifactOriginAnnotations(t *testing.T) {
	// A resource, a generated ConfigMap and a transformer from the
	// artifact: their annotations name its repository, tag and paths.
	host := ocitest.StartLoopbackRegistry(t)
	ocitest.Push(t, nil, host+ociRepo, ocitest.Files(t, map[string]string{
		"base/kustomization.yaml": `
resources:
- pod.yaml
configMapGenerator:
- name: settings
  literals:
  - mode=base
generatorOptions:
  disableNameSuffixHash: true
`,
		"base/pod.yaml": ociArtifactFiles()["base/pod.yaml"],
		"overlays/dev/kustomization.yaml": `
resources:
- ../../base
namePrefix: dev-
`,
	}))

	th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
	dir := t.TempDir()
	th.WriteK(dir, "buildMetadata: [originAnnotations, transformerAnnotations]\n"+
		resourcesPrefix+ociScheme+host+ociRepo+ociOverlay+"\n")

	m := th.Run(dir, th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	repo := "oci://" + host + "/some-org/some-repo"
	require.Equal(t, `apiVersion: v1
kind: Pod
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - repo: `+repo+`
        ref: v1.0.0
        configuredIn: overlays/dev/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
    config.kubernetes.io/origin: |
      path: base/pod.yaml
      repo: `+repo+`
      ref: v1.0.0
  name: dev-myPod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
---
apiVersion: v1
data:
  mode: base
kind: ConfigMap
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - repo: `+repo+`
        ref: v1.0.0
        configuredIn: overlays/dev/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
    config.kubernetes.io/origin: |
      repo: `+repo+`
      ref: v1.0.0
      configuredIn: base/kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: dev-settings
`, string(yml))
}

func TestOciArtifactAsTarget(t *testing.T) {
	host := ocitest.StartLoopbackRegistry(t)
	ref := host + ociRepo
	ocitest.Push(t, nil, ref, ocitest.Files(t, ociArtifactFiles()))

	th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
	m := th.Run(ociScheme+ref+ociOverlay, th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, ociDevBuild, string(yml))
}

func TestOciArtifactLocalize(t *testing.T) {
	// Localize copies the artifact, by tag and by digest, below
	// localized-files, and the copy builds as the original does.
	host := ocitest.StartLoopbackRegistry(t)
	digest := ocitest.Push(t, nil, host+ociRepo, ocitest.Files(t, ociArtifactFiles()))
	fSys := filesys.MakeFsOnDisk()
	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	target := t.TempDir()
	th.WriteK(target, resourcesPrefix+ociScheme+host+ociRepo+ociOverlay+"\n- "+
		ociScheme+host+"/some-org/some-repo@"+digest+"//base\n")

	dst, err := localizer.Run(fSys, target, "", filepath.Join(t.TempDir(), "dst"))
	require.NoError(t, err)
	registry, _, err := net.SplitHostPort(host)
	require.NoError(t, err)
	artifact := filepath.Join("localized-files", "oci-schemed", registry, "some-org", "some-repo")
	content, err := fSys.ReadFile(filepath.Join(dst, "kustomization.yaml"))
	require.NoError(t, err)
	var kust types.Kustomization
	require.NoError(t, yaml.Unmarshal(content, &kust))
	require.Equal(t, []string{
		filepath.Join(artifact, "v1.0.0", "overlays", "dev"),
		filepath.Join(artifact, strings.Replace(digest, ":", "-", 1), "base"),
	}, kust.Resources)

	m := th.Run(dst, th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, ociDevBuild+"---\n"+ociBaseBuild, string(yml))
}

func TestOciArtifactLocalizeTarget(t *testing.T) {
	// An artifact as the localize target: the default destination is
	// named after its root and tag, and the copy builds as it does.
	host := ocitest.StartLoopbackRegistry(t)
	ocitest.Push(t, nil, host+ociRepo, ocitest.Files(t, ociArtifactFiles()))
	fSys := filesys.MakeFsOnDisk()
	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	// The default destination is created in the working directory.
	t.Chdir(t.TempDir())

	dst, err := localizer.Run(fSys, ociScheme+host+ociRepo+ociOverlay, "", "")
	require.NoError(t, err)
	require.Equal(t, "localized-dev-v1.0.0", filepath.Base(dst))
	m := th.Run(filepath.Join(dst, "overlays", "dev"), th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, ociDevBuild, string(yml))
}

func TestOciArtifactLocalizeNeedsTagOrDigest(t *testing.T) {
	// Without a tag or a digest, the copy would be of the moving latest.
	host := ocitest.StartLoopbackRegistry(t)
	ocitest.Push(t, nil, host+"/some-org/some-repo:latest", ocitest.Files(t, ociArtifactFiles()))
	fSys := filesys.MakeFsOnDisk()
	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	dir := t.TempDir()
	th.WriteK(dir, resourcesPrefix+ociScheme+host+"/some-org/some-repo//base\n")

	_, err := localizer.Run(fSys, dir, "", filepath.Join(t.TempDir(), "dst"))
	require.ErrorContains(t, err, "missing tag or digest")
}

func TestOciArtifactErrors(t *testing.T) {
	host := ocitest.StartLoopbackRegistry(t)
	ref := host + ociRepo
	ocitest.Push(t, nil, ref, ocitest.Files(t, ociArtifactFiles()))

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
	host := ocitest.StartLoopbackRegistry(t)
	ref := host + "/some-org/escaping:v1.0.0"
	ocitest.Push(t, nil, ref, ocitest.Files(t, map[string]string{
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
