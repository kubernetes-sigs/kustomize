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
	. "sigs.k8s.io/kustomize/api/internal/localizer"
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
			assert.NoErrorf(t, findErr, "unexpected file %q", path)
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

func TestTargetIsScope(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
`,
	}
	addFiles(t, fSys, "/a", kustomization)
	err := Run("/a", "", "/a/b/dst", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustomization)
	addFiles(t, fSysExpected, "/a/b/dst", kustomization)
	checkFSys(t, fSysExpected, fSys)
}

func TestTargetNestedInScope(t *testing.T) {
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
	err := Run("/a/b", "/", "/a/b/dst", fSys)
	require.NoError(t, err)

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

	err := Run("/a", "/", "/dst", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustomization)
	addFiles(t, fSysExpected, "/dst/a", map[string]string{
		"kustomization.yaml": kustomization["Kustomization"],
	})
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeFileName(t *testing.T) {
	for name, path := range map[string]string{
		"nested_directories":                  "a/b/c/d/patch.yaml",
		"localize_dir_name_when_no_remote":    LocalizeDir,
		"in_localize_dir_name_when_no_remote": fmt.Sprintf("%s/patch.yaml", LocalizeDir),
		"no_file_extension":                   "patch",
		"kustomization_name":                  "a/kustomization.yaml",
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

			err := Run("/a", "/", "/a/dst", fSys)
			require.NoError(t, err)

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

	err := Run("/alpha/beta/gamma", "/", "", fSys)
	require.NoError(t, err)

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

	err := Run("/", "", "", fSys)
	require.NoError(t, err)

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

	err := Run("/a/b", "", "/dst", fSys)
	require.Error(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a/b", kustAndPatch)
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeGenerators(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
generators:
- plugin.yaml
- |
  apiVersion: builtin
  behavior: create
  kind: ConfigMapGenerator
  literals:
  - APPLE=orange
  metadata:
    name: another-map
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
`,
		"plugin.yaml": `apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: map
`,
	}
	addFiles(t, fSys, "/a", kustAndPlugins)

	err := Run("/a", "", "/alpha/dst", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustAndPlugins)
	addFiles(t, fSysExpected, "/alpha/dst", map[string]string{
		"kustomization.yaml": kustAndPlugins["kustomization.yaml"],
	})
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeTransformers(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- |
  apiVersion: builtin
  jsonOp: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
  kind: PatchJson6902Transformer
  metadata:
    name: patch6902
  target:
    name: deployment
  ---
  apiVersion: builtin
  kind: ReplacementTransformer
  metadata:
    name: replacement
  replacements:
  - source:
      fieldPath: spec.template.spec.containers.0.image
      kind: Deployment
    targets:
    - fieldPaths:
      - spec.template.spec.containers.1.image
      select:
        kind: Deployment
- plugin.yaml
`,
		"plugin.yaml": `apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: patchSM
paths:
- pod.yaml
`,
	}
	addFiles(t, fSys, "/a", kustAndPlugins)

	err := Run("/a", "", "/dst", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", kustAndPlugins)
	addFiles(t, fSysExpected, "/dst", map[string]string{
		"kustomization.yaml": kustAndPlugins["kustomization.yaml"],
	})
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeValidators(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndPlugin := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
validators:
- |-
  apiVersion: builtin
  kind: ReplacementTransformer
  metadata:
    name: replacement
  replacements:
  - source:
      kind: ConfigMap
      fieldPath: metadata.name
    targets:
    - select:
        kind: ConfigMap
      fieldPaths:
      - metadata.name
- replacement.yaml
`,
		"replacement.yaml": `apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: replacement-2
replacements:
- source:
    kind: Secret
    fieldPath: data.USER_NAME
  targets:
  - select:
      kind: Secret
    fieldPaths:
    - data.USER_NAME
`,
	}
	addFiles(t, fSys, "/", kustAndPlugin)
	err := Run("/", "", "/dst", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/", kustAndPlugin)
	addFiles(t, fSysExpected, "/dst", map[string]string{
		"kustomization.yaml": kustAndPlugin["kustomization.yaml"],
	})
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeBuiltinPluginsNotResource(t *testing.T) {
	type testCase struct {
		name         string
		files        map[string]string
		errPrefix    string
		inlineErrMsg string
		fileErrMsg   string
	}
	for _, test := range []testCase{
		{
			name: "bad_inline_resource",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
generators:
- |
  apiVersion: builtin
  kind: ConfigMapGenerator
kind: Kustomization
`,
			},
			errPrefix:    `unable to load generators entry: unable to load resource entry "apiVersion: builtin\nkind: ConfigMapGenerator\n"`,
			inlineErrMsg: `missing metadata.name in object {{builtin ConfigMapGenerator} {{ } map[] map[]}}`,
			fileErrMsg: `invalid file reference: '/apiVersion: builtin
kind: ConfigMapGenerator
' doesn't exist`,
		},
		{
			name: "bad_file_resource",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- plugin.yaml
`,
				"plugin.yaml": `apiVersion: builtin
metadata:
  name: PatchTransformer
`,
			},
			errPrefix:    `unable to load transformers entry: unable to load resource entry "plugin.yaml"`,
			inlineErrMsg: `missing Resource metadata`,
			fileErrMsg:   `missing kind in object {{builtin } {{PatchTransformer } map[] map[]}}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fSys := makeMemoryFs(t)
			addFiles(t, fSys, "/", test.files)
			err := Run("/", "", "/dst", fSys)

			var actualErr ResourceLoadError
			require.ErrorAs(t, err, &actualErr)
			require.EqualError(t, actualErr.InlineError, test.inlineErrMsg)
			require.EqualError(t, actualErr.FileError, test.fileErrMsg)

			require.EqualError(t, err, fmt.Sprintf(`unable to localize target "/": %s: when parsing as inline received error: %s
when parsing as filepath received error: %s`, test.errPrefix, test.inlineErrMsg, test.fileErrMsg))

			fSysExpected := makeMemoryFs(t)
			addFiles(t, fSysExpected, "/", test.files)
			checkFSys(t, fSysExpected, fSys)
		})
	}
}

func TestLocalizeDirInTarget(t *testing.T) {
	type testCase struct {
		name  string
		files map[string]string
	}
	for _, tc := range []testCase{
		{
			name: "multi_nested_child",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- delta/epsilon
kind: Kustomization
`,
				"delta/epsilon/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
namespace: kustomize-namespace
`,
			},
		},
		{
			name: "recursive",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- delta
kind: Kustomization
`,
				"delta/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
components:
- epsilon
kind: Component
`,
				"delta/epsilon/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
namespace: kustomize-namespace
`,
			},
		},
		{
			name: "file_in_dir",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- delta
kind: Kustomization
`,
				"delta/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
patches:
- path: patch.yaml
`,
				"delta/patch.yaml": podConfiguration,
			},
		},
		{
			name: "multiple_calls",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- delta
- delta/epsilon
kind: Kustomization
`,
				"delta/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
namespace: kustomize-namespace
`,
				"delta/epsilon/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
buildMetadata:
- managedByLabel
kind: Component
`,
			},
		},
		{
			name: "localize_directory_name_when_no_remote",
			files: map[string]string{
				"kustomization.yaml": fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
components:
- %s
kind: Kustomization
`, LocalizeDir),
				fmt.Sprintf("%s/kustomization.yaml", LocalizeDir): `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
namespace: kustomize-namespace
`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fSys := makeMemoryFs(t)
			addFiles(t, fSys, "/alpha/beta/gamma", tc.files)

			err := Run("/alpha/beta/gamma", "/alpha/beta", "/dst", fSys)
			require.NoError(t, err)

			fSysExpected := makeMemoryFs(t)
			addFiles(t, fSysExpected, "/alpha/beta/gamma", tc.files)
			addFiles(t, fSysExpected, "/dst/gamma", tc.files)
			checkFSys(t, fSysExpected, fSys)
		})
	}
}

func TestLocalizeDirCleanedSibling(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndComponents := map[string]string{
		// This test checks that winding paths that might traverse through directories
		// outside of scope, which will not be present at destination, are cleaned.
		"beta/gamma/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- delta/../../../../a/b/../../alpha/beta/sibling
kind: Kustomization`,
		"beta/sibling/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
namespace: kustomize-namespace
`,
	}
	addFiles(t, fSys, "/alpha", kustAndComponents)

	err := Run("/alpha/beta/gamma", "/alpha", "/alpha/beta/dst", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/alpha", kustAndComponents)
	cleanedFiles := map[string]string{
		"beta/gamma/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- ../sibling
kind: Kustomization
`,
		"beta/sibling/kustomization.yaml": kustAndComponents["beta/sibling/kustomization.yaml"],
	}
	addFiles(t, fSysExpected, "/alpha/beta/dst", cleanedFiles)
	checkFSys(t, fSysExpected, fSys)
}

func TestLocalizeComponents(t *testing.T) {
	fSys := makeMemoryFs(t)
	kustAndComponents := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- a
- alpha
kind: Kustomization
`,
		"a/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
namePrefix: my-
`,
		"alpha/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
nameSuffix: -test
`,
	}
	addFiles(t, fSys, "/", kustAndComponents)

	err := Run("/", "", "", fSys)
	require.NoError(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/", kustAndComponents)
	addFiles(t, fSysExpected, "/localized", kustAndComponents)
	checkFSys(t, fSysExpected, fSys)
}
