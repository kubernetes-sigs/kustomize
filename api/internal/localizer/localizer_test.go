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
    - containerPort: 80`

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

func makeFileSystems(t *testing.T, target string, files map[string]string) (expected filesys.FileSystem, actual filesys.FileSystem) {
	t.Helper()

	copies := make([]filesys.FileSystem, 2)
	for i := range copies {
		copies[i] = makeMemoryFs(t)
		addFiles(t, copies[i], target, files)
	}
	return copies[0], copies[1]
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
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
`,
	}
	fSysExpected, fSysActual := makeFileSystems(t, "/a", kustomization)

	err := Run("/a", "", "/a/b/dst", fSysActual)
	require.NoError(t, err)

	addFiles(t, fSysExpected, "/a/b/dst", kustomization)
	checkFSys(t, fSysExpected, fSysActual)
}

func TestTargetNestedInScope(t *testing.T) {
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
	fSysExpected, fSysActual := makeFileSystems(t, "/a/b", kustomization)

	err := Run("/a/b", "/", "/a/b/dst", fSysActual)
	require.NoError(t, err)

	addFiles(t, fSysExpected, "/a/b/dst/a/b", kustomization)
	checkFSys(t, fSysExpected, fSysActual)
}

func TestLoadKustomizationName(t *testing.T) {
	kustomization := map[string]string{
		"Kustomization": `apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  label-one: value-one
  label-two: value-two
kind: Kustomization
`,
	}
	fSysExpected, fSysActual := makeFileSystems(t, "/a", kustomization)

	err := Run("/a", "/", "/dst", fSysActual)
	require.NoError(t, err)

	addFiles(t, fSysExpected, "/dst/a", kustomization)
	checkFSys(t, fSysExpected, fSysActual)
}

func TestLoadGVKNN(t *testing.T) {
	for name, kustomization := range map[string]string{
		"missing": `namePrefix: my-
`,
		"wrong": `kind: NotChecked
`,
	} {
		t.Run(name, func(t *testing.T) {
			files := map[string]string{
				"kustomization.yaml": kustomization,
			}
			fSysExpected, fSysActual := makeFileSystems(t, "/a", files)

			err := Run("/a", "/a", "/dst", fSysActual)
			require.NoError(t, err)

			addFiles(t, fSysExpected, "/dst", files)
			checkFSys(t, fSysExpected, fSysActual)
		})
	}
}

func TestLoadLegacyFields(t *testing.T) {
	// TODO(annasong): add referenced files when implement legacy field localization
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
bases:
- beta
configMapGenerator:
- env: env.properties
imageTags:
- name: postgres
  newName: my-registry/my-postgres
  newTag: v1
kind: Kustomization
`,
	}
	fSysExpected, fSysActual := makeFileSystems(t, "/alpha", kustomization)

	err := Run("/alpha", "/alpha", "/beta", fSysActual)
	require.NoError(t, err)

	addFiles(t, fSysExpected, "/beta", map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
bases:
- beta
configMapGenerator:
- env: env.properties
imageTags:
- name: postgres
  newName: my-registry/my-postgres
  newTag: v1
kind: Kustomization
`,
	})
	checkFSys(t, fSysExpected, fSysActual)
}

func TestLoadUnknownKustFields(t *testing.T) {
	fSysExpected, fSysTest := makeFileSystems(t, "/a", map[string]string{
		"kustomization.yaml": `namePrefix: valid
suffix: invalid`,
	})

	err := Run("/a", "", "", fSysTest)
	require.EqualError(t, err,
		`unable to localize target "/a": invalid Kustomization: error unmarshaling JSON: while decoding JSON: json: unknown field "suffix"`)

	checkFSys(t, fSysExpected, fSysTest)
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
			kustAndPatch := map[string]string{
				"kustomization.yaml": fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: %s
`, path),
				path: podConfiguration,
			}
			expected, actual := makeFileSystems(t, "/a", kustAndPatch)

			err := Run("/a", "/", "/a/dst", actual)
			require.NoError(t, err)

			addFiles(t, expected, "/a/dst/a", kustAndPatch)
			checkFSys(t, expected, actual)
		})
	}
}

func TestLocalizeFileCleaned(t *testing.T) {
	kustAndPatch := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: ../gamma/../../../alpha/beta/./gamma/patch.yaml
`,
		"patch.yaml": podConfiguration,
	}
	expected, actual := makeFileSystems(t, "/alpha/beta/gamma", kustAndPatch)

	err := Run("/alpha/beta/gamma", "/", "", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/localized-gamma/alpha/beta/gamma", map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: patch.yaml
`,
		"patch.yaml": podConfiguration,
	})
	checkFSys(t, expected, actual)
}

func TestLocalizeUnreferencedIgnored(t *testing.T) {
	targetAndUnreferenced := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
configMapGenerator:
- envs:
  - env
  name: referenced-file
kind: Kustomization
`,
		"env":            "APPLE=orange",
		"env.properties": "USERNAME=password",
		"resource.yaml":  podConfiguration,
	}
	expected, actual := makeFileSystems(t, "/alpha/beta", targetAndUnreferenced)

	err := Run("/alpha/beta", "/alpha", "/beta", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/beta/beta", map[string]string{
		"kustomization.yaml": targetAndUnreferenced["kustomization.yaml"],
		"env":                targetAndUnreferenced["env"],
	})
	checkFSys(t, expected, actual)
}

func TestLocalizePatches(t *testing.T) {
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
	expected, actual := makeFileSystems(t, "/", kustAndPatch)

	err := Run("/", "", "", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/localized", kustAndPatch)
	checkFSys(t, expected, actual)
}

func TestLocalizePatchesJson(t *testing.T) {
	kustAndPatches := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patchesJson6902:
- path: patch.yaml
  target:
    annotationSelector: zone=west
    name: pod
    version: v1
- patch: '[{"op": "add", "path": "/new/path", "value": "value"}]'
  target:
    group: apps
    kind: Pod
- path: patch.json
  target:
    namespace: my
`,
		"patch.yaml": `- op: add
  path: /some/new/path
  value: value
- op: replace
  path: /some/existing/path
  value: new value`,
		"patch.json": ` [
   {"op": "copy", "from": "/here", "path": "/there"},
   {"op": "remove", "path": "/some/existing/path"},
 ]`,
	}
	expected, actual := makeFileSystems(t, "/alpha/beta", kustAndPatches)

	err := Run("/alpha/beta", "/", "/beta", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/beta/alpha/beta", kustAndPatches)
	checkFSys(t, expected, actual)
}

func TestLocalizePatchesSM(t *testing.T) {
	kustAndPatches := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patchesStrategicMerge:
- |-
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: map
  data:
  - APPLE: orange
- patch.yaml
`,
		"patch.yaml": podConfiguration,
	}
	expected, actual := makeFileSystems(t, "/a", kustAndPatches)

	err := Run("/a", "", "/dst", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/dst", kustAndPatches)
	checkFSys(t, expected, actual)
}

func TestLocalizeReplacements(t *testing.T) {
	kustAndReplacement := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
replacements:
- path: replacement.yaml
- source:
    fieldPath: path
    name: map
  targets:
  - fieldPaths:
    - path
    select:
      name: my-map
`,
		"replacement.yaml": `source:
  fieldPath: path.*.to.[some=field]
  kind: Pod
  options:
    delimiter: /
targets:
- fieldPaths:
  - config\.kubernetes\.io.annotations
  - second.path
  reject:
  - group: apps
    version: v2
  select:
    namespace: my`,
	}
	expected, actual := makeFileSystems(t, "/a", kustAndReplacement)

	err := Run("/a", "/", "/dst", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/dst/a", kustAndReplacement)
	checkFSys(t, expected, actual)
}

func TestLocalizeConfigMapGenerator(t *testing.T) {
	kustAndData := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
configMapGenerator:
- envs:
  - standard.env
  namespace: my
  options:
    immutable: true
- behavior: merge
  files:
  - key.properties
  literals:
  - PEAR=pineapple
kind: Kustomization
metadata:
  name: test
`,
		"standard.env": `SIZE=0.1
IS_GLOBAL=true`,
		"key.properties": "value",
	}
	expected, actual := makeFileSystems(t, "/a/b", kustAndData)

	err := Run("/a/b", "", "", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/localized-b", kustAndData)
	checkFSys(t, expected, actual)
}

func TestLocalizeSecretGenerator(t *testing.T) {
	kustAndData := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- behavior: create
  files:
  - key=b/value.properties
  - b/value
  name: secret
- envs:
  - crt
  - key
  type: kubernetes.io/tls
- literals:
  - APPLE=b3Jhbmdl
  - PLUM=cGx1b3Q=
  name: no-files
`,
		"crt":                "tls.crt=LS0tLS1CRUd...0tLQo=",
		"key":                "tls.key=LS0tLS1CRUd...0tLQo=",
		"b/value.properties": "dmFsdWU=",
		"b/value":            "dmFsdWU=",
	}
	expected, actual := makeFileSystems(t, "/a", kustAndData)

	err := Run("/a", "/", "/localized-a", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/localized-a/a", kustAndData)
	checkFSys(t, expected, actual)
}

func TestLocalizeFileNoFile(t *testing.T) {
	kustAndPatch := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: name-DNE.yaml
`,
	}
	expected, actual := makeFileSystems(t, "/a/b", kustAndPatch)

	err := Run("/a/b", "", "/dst", actual)
	require.Error(t, err)

	checkFSys(t, expected, actual)
}

func TestLocalizeGenerators(t *testing.T) {
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
	expected, actual := makeFileSystems(t, "/a", kustAndPlugins)

	err := Run("/a", "", "/alpha/dst", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/alpha/dst", map[string]string{
		"kustomization.yaml": kustAndPlugins["kustomization.yaml"],
	})
	checkFSys(t, expected, actual)
}

func TestLocalizeTransformers(t *testing.T) {
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
	expected, actual := makeFileSystems(t, "/a", kustAndPlugins)

	err := Run("/a", "", "/dst", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/dst", map[string]string{
		"kustomization.yaml": kustAndPlugins["kustomization.yaml"],
	})
	checkFSys(t, expected, actual)
}

func TestLocalizeValidators(t *testing.T) {
	kustAndPlugin := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
validators:
- |
  apiVersion: builtin
  kind: ReplacementTransformer
  metadata:
    name: replacement
  replacements:
  - source:
      fieldPath: data.[field=value]
      group: apps
    targets:
    - fieldPaths:
      - spec.*
      select:
        kind: Pod
- replacement.yaml
`,
		"replacement.yaml": `apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: replacement-2
replacements:
- source:
    fieldPath: spec.containers.1.image
    kind: Custom
    namespace: test
  targets:
  - fieldPaths:
    - path
    select:
      namespace: test
`,
	}
	expected, actual := makeFileSystems(t, "/", kustAndPlugin)

	err := Run("/", "", "/dst", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/dst", map[string]string{
		"kustomization.yaml": kustAndPlugin["kustomization.yaml"],
	})
	checkFSys(t, expected, actual)
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
			expected, actual := makeFileSystems(t, "/", test.files)

			err := Run("/", "", "/dst", actual)

			var actualErr ResourceLoadError
			require.ErrorAs(t, err, &actualErr)
			require.EqualError(t, actualErr.InlineError, test.inlineErrMsg)
			require.EqualError(t, actualErr.FileError, test.fileErrMsg)

			require.EqualError(t, err, fmt.Sprintf(`unable to localize target "/": %s: when parsing as inline received error: %s
when parsing as filepath received error: %s`, test.errPrefix, test.inlineErrMsg, test.fileErrMsg))

			checkFSys(t, expected, actual)
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
			expected, actual := makeFileSystems(t, "/alpha/beta/gamma", tc.files)

			err := Run("/alpha/beta/gamma", "/alpha/beta", "/dst", actual)
			require.NoError(t, err)

			addFiles(t, expected, "/dst/gamma", tc.files)
			checkFSys(t, expected, actual)
		})
	}
}

func TestLocalizeDirCleanedSibling(t *testing.T) {
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
	expected, actual := makeFileSystems(t, "/alpha", kustAndComponents)

	err := Run("/alpha/beta/gamma", "/alpha", "/alpha/beta/dst", actual)
	require.NoError(t, err)

	cleanedFiles := map[string]string{
		"beta/gamma/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- ../sibling
kind: Kustomization
`,
		"beta/sibling/kustomization.yaml": kustAndComponents["beta/sibling/kustomization.yaml"],
	}
	addFiles(t, expected, "/alpha/beta/dst", cleanedFiles)
	checkFSys(t, expected, actual)
}

func TestLocalizeComponents(t *testing.T) {
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
	expected, actual := makeFileSystems(t, "/", kustAndComponents)

	err := Run("/", "", "", actual)
	require.NoError(t, err)

	addFiles(t, expected, "/localized", kustAndComponents)
	checkFSys(t, expected, actual)
}
