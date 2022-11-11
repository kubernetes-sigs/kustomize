// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/hasher"
	. "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/validate"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const podConfiguration = `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    -containerPort: 80
`

func makeMemoryFs(t *testing.T) filesys.FileSystem {
	t.Helper()
	req := require.New(t)

	fSys := filesys.MakeFsInMemory()
	req.NoError(fSys.MkdirAll("/a/b"))
	req.NoError(fSys.WriteFile("/a/pod.yaml", []byte(podConfiguration)))

	dirChain := "/alpha/beta/gamma/delta"
	req.NoError(fSys.MkdirAll(dirChain))
	req.NoError(fSys.WriteFile(filepath.Join(dirChain, "deployment.yaml"), []byte("deployment configuration")))
	req.NoError(fSys.Mkdir("/alpha/beta/say"))
	return fSys
}

func addFiles(t *testing.T, fSys filesys.FileSystem, parentDir string, files map[string]string) {
	t.Helper()

	// in-memory file system makes all necessary dirs when writing files
	for file, content := range files {
		require.NoError(t, fSys.WriteFile(filepath.Join(parentDir, file), []byte(content)))
	}
}

func createLocalizer(t *testing.T, fSys filesys.FileSystem, target string, scope string, newDir string) *Localizer {
	t.Helper()

	// no need to re-test Loader
	ldr, _, err := NewLoader(target, scope, newDir, fSys)
	require.NoError(t, err)
	rmFactory := resmap.NewFactory(resource.NewFactory(&hasher.Hasher{}))
	lc, err := NewLocalizer(
		ldr,
		validate.NewFieldValidator(),
		rmFactory,
		// file system can be in memory, as plugin configuration will prevent the use of file system anyway
		loader.NewLoader(types.DisabledPluginConfig(), rmFactory, fSys))
	require.NoError(t, err)
	return lc
}

func TestNewLocalizerTargetIsScope(t *testing.T) {
	fSys := makeMemoryFs(t)
	_ = createLocalizer(t, fSys, "/a", "", "/a/b/dst")

	fSysExpected := makeMemoryFs(t)
	require.NoError(t, fSysExpected.MkdirAll("/a/b/dst"))
	require.Equal(t, fSysExpected, fSys)
}

func TestNewLocalizerTargetNestedInScope(t *testing.T) {
	fSys := makeMemoryFs(t)
	_ = createLocalizer(t, fSys, "/a/b", "/", "/a/b/dst")

	fSysExpected := makeMemoryFs(t)
	require.NoError(t, fSysExpected.MkdirAll("/a/b/dst/a/b"))
	require.Equal(t, fSysExpected, fSys)
}

func TestLocalizeKustomizationName(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustomization := map[string]string{
		"Kustomization": `apiVersion: kustomize.config.k8s.io/v1beta1
configMapGenerator:
- behavior: create
  literals:
  - APPLE=orange
  name: map
kind: Kustomization
resources:
- pod.yaml
`,
	}
	addFiles(t, fSys, "/a", kustomization)

	lclzr := createLocalizer(t, fSys, "/a", "/", "/dst")
	require.NoError(t, lclzr.Localize())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustomization)
	addFiles(t, fSysExpected, "/dst/a", map[string]string{
		"kustomization.yaml": kustomization["Kustomization"],
	})
	require.Equal(t, fSysExpected, fSys)
}
