// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
    - containerPort: 80
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

func checkFSys(t *testing.T, fSysExpected filesys.FileSystem, fSysActual filesys.FileSystem) {
	t.Helper()

	assert.Equal(t, fSysExpected, fSysActual)
	if t.Failed() {
		reportFSysDiff(t, fSysExpected, fSysActual)
	}
}

func reportFSysDiff(t *testing.T, fSysExpected filesys.FileSystem, fSysActual filesys.FileSystem) {
	t.Helper()

	visited := make(map[string]struct{})
	err := fSysActual.Walk("/", func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)
		visited[path] = struct{}{}

		if info.IsDir() {
			assert.Truef(t, fSysExpected.IsDir(path), "unexpected directory %q", path)
		} else {
			actualContent, readErr := fSysActual.ReadFile(path)
			require.NoError(t, readErr)
			expectedContent, findErr := fSysExpected.ReadFile(path)
			assert.NoError(t, findErr)
			if findErr == nil {
				assert.Equal(t, string(expectedContent), string(actualContent))
			}
		}
		return nil
	})
	require.NoError(t, err)

	err = fSysExpected.Walk("/", func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)
		visited[path] = struct{}{}

		if _, exists := visited[path]; !exists {
			t.Errorf("expected path %q not found", path)
		}
		return nil
	})
	require.NoError(t, err)
}

func TestNewLocalizerTargetIsScope(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
`,
	}
	addFiles(t, fSys, "/a", kustomization)
	lclzr := createLocalizer(t, fSys, "/a", "", "/a/b/dst")
	require.NoError(t, lclzr.Localize())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustomization)
	addFiles(t, fSysExpected, "/a/b/dst", kustomization)
	checkFSys(t, fSysExpected, fSys)
}

func TestNewLocalizerTargetNestedInScope(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- patch: |-
    - op: replace
      path: /some/existing/path
      value: new value
  target:
    kind: Deployment
    labelSelector: env=dev
`,
	}
	addFiles(t, fSys, "/a/b", kustomization)
	lclzr := createLocalizer(t, fSys, "/a/b", "/", "/a/b/dst")
	require.NoError(t, lclzr.Localize())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a/b", kustomization)
	addFiles(t, fSysExpected, "/a/b/dst/a/b", kustomization)
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeKustomizationName(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustomization := map[string]string{
		"Kustomization": `apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  label-one: value-one
  label-two: value-two
kind: Kustomization
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
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeFileName(t *testing.T) {
	for name, path := range map[string]string{
		"nested_directories":               "a/b/c/d/patch.yaml",
		"localize_dir_name_when_absent":    LocalizeDir,
		"in_localize_dir_name_when_absent": fmt.Sprintf("%s/patch.yaml", LocalizeDir),
		"no_file_extension":                "patch",
		"kustomization_name":               "a/kustomization.yaml",
	} {
		t.Run(name, func(t *testing.T) {
			fSys := makeMemoryFs(t)
			kustAndPatch := map[string]string{
				"kustomization.yaml": fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: %s
`, path),
				path: podConfiguration,
			}
			addFiles(t, fSys, "/a", kustAndPatch)

			lclzr := createLocalizer(t, fSys, "/a", "/", "/a/dst")
			require.NoError(t, lclzr.Localize())

			fSysExpected := makeMemoryFs(t)
			addFiles(t, fSysExpected, "/a", kustAndPatch)
			addFiles(t, fSysExpected, "/a/dst/a", kustAndPatch)
			checkFSys(t, fSysExpected, fSys)
		})
	}
}

func TestLocalizeFileCleaned(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPatch := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: ../gamma/../../../alpha/beta/./gamma/patch.yaml
`,
		"patch.yaml": podConfiguration,
	}
	addFiles(t, fSys, "/alpha/beta/gamma", kustAndPatch)

	lclzr := createLocalizer(t, fSys, "/alpha/beta/gamma", "/", "")
	require.NoError(t, lclzr.Localize())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/alpha/beta/gamma", kustAndPatch)
	addFiles(t, fSysExpected, "/localized-gamma/alpha/beta/gamma", map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: patch.yaml
`,
		"patch.yaml": podConfiguration,
	})
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizePatches(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPatch := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- patch: |-
    apiVersion: v1
    kind: Deployment
    metadata:
      labels:
        app.kubernetes.io/version: 1.21.0
      name: dummy-app
  target:
    labelSelector: app.kubernetes.io/name=nginx
- options:
    allowNameChange: true
  path: patch.yaml
`,
		"patch.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Deployment
metadata:
  name: not-used
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.21.0
`,
	}
	addFiles(t, fSys, "/", kustAndPatch)

	lclzr := createLocalizer(t, fSys, "/", "", "")
	require.NoError(t, lclzr.Localize())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/", kustAndPatch)
	addFiles(t, fSysExpected, "/localized", kustAndPatch)
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeFileNoFile(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPatch := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: name-DNE.yaml
`,
	}
	addFiles(t, fSys, "/a/b", kustAndPatch)

	lclzr := createLocalizer(t, fSys, "/a/b", "", "/dst")
	require.Error(t, lclzr.Localize())
}

func TestLocalizePlugins(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
generators:
- |-
  apiVersion: builtin
  behavior: create
  kind: ConfigMapGenerator
  literals:
  - APPLE=orange
  metadata:
    name: map
  ---
  apiVersion: builtin
  kind: SecretGenerator
  literals:
  - APPLE=b3Jhbmdl
  metadata:
    name: secret
  options:
    disableNameSuffixHash: true
kind: Kustomization
transformers:
- plugin.yaml
`,
		"plugin.yaml": `apiVersion: builtin
kind: PatchTransformer
metadata:
  name: patch
patch: '[{"op": "replace", "path": "/spec/replicas", "value": "2"}]'
target:
  name: .*Deploy
`,
	}
	addFiles(t, fSys, "/a", kustAndPlugins)

	lclzr := createLocalizer(t, fSys, "/a", "", "/alpha/dst")
	require.NoError(t, lclzr.Localize())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustAndPlugins)
	addFiles(t, fSysExpected, "/alpha/dst", map[string]string{
		"kustomization.yaml": kustAndPlugins["kustomization.yaml"],
	})
	require.Equal(t, fSysExpected, fSys)
}

func TestLocalizePluginsNotBuiltin(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
generators:
- |-
  apiVersion: builtin
  kind: ConfigMapGenerator
  literals:
  - APPLE=orange
  metadata:
    name: map
kind: Kustomization
transformers:
- plugin.yaml
`,
		"plugin.yaml": `apiVersion: random
kind: PatchTransformer
metadata:
  name: patch
patch: '[{"op": "replace", "path": "/spec/replicas", "value": "2"}]'
target:
  name: .*Deploy
`,
	}
	addFiles(t, fSys, "/", kustAndPlugins)

	lclzr := createLocalizer(t, fSys, "/", "", "/dst")
	require.Error(t, lclzr.Localize())
}
