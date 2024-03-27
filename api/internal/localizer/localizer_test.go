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

const (
	podConfiguration = `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    - containerPort: 80`

	replacementTransformerWithPath = `apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: replacement
replacements:
- path: replacement.yaml
- source:
    fieldPath: metadata.[name=my-pod]
    group: apps
    namespace: test
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.name
    select:
      name: another-pod
`

	replacements = `
- source:
    name: src
    fieldPath: path
    options:
      delimiter: '='
      index: 1
  targets:
  - select:
      kind: Pod
    reject:
      version: v1
    fieldPaths:
    - metadata.annotations.config\.kubernetes\.io/local-config
    - sequence.*
- source:
    kind: Deployment
    fieldPath: sequence.-
  targets:
  - select:
      namespace: my
    fieldPaths:
    - path
    options:
      delimiter: '='
      index: 0
`

	valuesFile = `minecraftServer:
  difficulty: peaceful
`
)

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

func checkRun(t *testing.T, fSys filesys.FileSystem, target, scope, dst string) {
	t.Helper()

	actualDst, err := Run(target, scope, dst, fSys)
	require.NoError(t, err)
	require.Equal(t, dst, actualDst)
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
			require.NoErrorf(t, findErr, "unexpected file %q", path)
			if findErr == nil {
				assert.Equal(t, string(expectedContent), string(actualContent))
			}
		}
		return nil
	})
	require.NoError(t, err)

	err = fSysExpected.Walk("/", func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		_, exists := visited[path]
		assert.Truef(t, exists, "expected path %q not found", path)
		return nil
	})
	require.NoError(t, err)
}

func checkLocalizeInTargetSuccess(t *testing.T, files map[string]string) {
	t.Helper()

	fSys := makeMemoryFs(t)
	addFiles(t, fSys, "/a", files)

	checkRun(t, fSys, "/a", "/", "/dst")
	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/a", files)
	addFiles(t, fSysExpected, "/dst/a", files)
	checkFSys(t, fSysExpected, fSys)
}

func TestTargetIsScope(t *testing.T) {
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
`,
	}
	fSysExpected, fSysActual := makeFileSystems(t, "/a", kustomization)

	checkRun(t, fSysActual, "/a", "/a", "/a/b/dst")
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

	checkRun(t, fSysActual, "/a/b", "/", "/a/b/dst")
	addFiles(t, fSysExpected, "/a/b/dst/a/b", kustomization)
	checkFSys(t, fSysExpected, fSysActual)
}

func TestLoadKustomizationName(t *testing.T) {
	kustomization := map[string]string{
		"Kustomization": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
labels:
- pairs:
    label-one: value-one
    label-two: value-two
`,
	}
	checkLocalizeInTargetSuccess(t, kustomization)
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
			checkLocalizeInTargetSuccess(t, files)
		})
	}
}

func TestLoadLegacyFields(t *testing.T) {
	kustomization := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  app: bingo
imageTags:
- name: postgres
  newName: my-registry/my-postgres
  newTag: v1
kind: Kustomization
`,
	}
	checkLocalizeInTargetSuccess(t, kustomization)
}

func TestLoadUnknownKustFields(t *testing.T) {
	fSysExpected, fSysTest := makeFileSystems(t, "/a", map[string]string{
		"kustomization.yaml": `namePrefix: valid
suffix: invalid`,
	})

	_, err := Run("/a", "", "", fSysTest)
	require.EqualError(t, err,
		`unable to localize target "/a": invalid Kustomization: json: unknown field "suffix"`)

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
			checkLocalizeInTargetSuccess(t, kustAndPatch)
		})
	}
}

func TestLocalizeFileCleaned(t *testing.T) {
	kustAndPatch := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: ../gamma/../../../alpha/beta/./gamma/patch.yaml
- path: /alpha/beta/../beta/./gamma/patch2.yaml
`,
		"patch.yaml":  podConfiguration,
		"patch2.yaml": podConfiguration,
	}
	expected, actual := makeFileSystems(t, "/alpha/beta/gamma", kustAndPatch)

	checkRun(t, actual, "/alpha/beta/gamma", "/", "/localized-gamma")
	addFiles(t, expected, "/localized-gamma/alpha/beta/gamma", map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: patch.yaml
- path: patch2.yaml
`,
		"patch.yaml":  podConfiguration,
		"patch2.yaml": podConfiguration,
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
		"env":               "APPLE=orange",
		"env.properties":    "USERNAME=password",
		"dir/resource.yaml": podConfiguration,
	}
	expected, actual := makeFileSystems(t, "/alpha/beta", targetAndUnreferenced)

	checkRun(t, actual, "/alpha/beta", "/alpha", "/beta")
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
		"patch.yaml": podConfiguration,
	}
	checkLocalizeInTargetSuccess(t, kustAndPatch)
}

func TestLocalizeOpenAPI(t *testing.T) {
	type testCase struct {
		name  string
		files map[string]string
	}
	for _, test := range []testCase{
		{
			name: "no_path",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
openapi:
  version: v1.20.4
`,
			},
		},
		{
			name: "path",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
openapi:
  path: openapi.json
`,
				"openapi.json": `{
  "definitions": {
    "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta": {
      "properties": {
        "name": {
          "type": "string"
        }
      },
      "type": "object"
    }
  }
}`,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkLocalizeInTargetSuccess(t, test.files)
		})
	}
}

func TestLocalizeConfigurations(t *testing.T) {
	kustAndConfigs := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
configurations:
- commonLabels.yaml
- namePrefix.yaml
kind: Kustomization
`,
		"commonLabels.yaml": `commonLabels:
- path: new/path
  create: true`,
		"namePrefix.yaml": `namePrefix:
- version: v1
  path: metadata/name
- group: custom
  path: metadata/name`,
	}
	checkLocalizeInTargetSuccess(t, kustAndConfigs)
}

func TestLocalizeCrds(t *testing.T) {
	kustAndCrds := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
crds:
- crd1.yaml
- crd2.yaml
kind: Kustomization
`,
		"crd1.yaml": `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: controller.stable.example.com`,
		"crd2.yaml": `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
scope: Cluster`,
	}
	checkLocalizeInTargetSuccess(t, kustAndCrds)
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
	checkLocalizeInTargetSuccess(t, kustAndPatches)
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
	checkLocalizeInTargetSuccess(t, kustAndPatches)
}

func TestLocalizeReplacements(t *testing.T) {
	kustAndReplacement := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
replacements:
- path: replacement.yaml
- source:
    fieldPath: path.0
    name: map
  targets:
  - fieldPaths:
    - path
    select:
      name: my-map
`,
		"replacement.yaml": replacements,
	}
	checkLocalizeInTargetSuccess(t, kustAndReplacement)
}

func TestLocalizeConfigMapGenerator(t *testing.T) {
	kustAndData := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
configMapGenerator:
- env: single.env
  envs:
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
		"single.env": `MAY=contain
MORE=than
ONE=pair`,
		"standard.env": `SIZE=0.1
IS_GLOBAL=true`,
		"key.properties": "value",
	}
	checkLocalizeInTargetSuccess(t, kustAndData)
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
  - APPLE=orange
  - PLUM=pluot
  name: no-files
- env: more-fruit
`,
		"crt":                "tls.crt=LS0tLS1CRUd...0tLQo=",
		"key":                "tls.key=LS0tLS1CRUd...0tLQo=",
		"more-fruit":         "GRAPE=lime",
		"b/value.properties": "value",
		"b/value":            "value",
	}
	checkLocalizeInTargetSuccess(t, kustAndData)
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

	_, err := Run("/a/b", "", "/dst", actual)
	require.EqualError(t, err, `unable to localize target "/a/b": unable to localize patches: invalid file reference: '/a/b/name-DNE.yaml' doesn't exist`)

	checkFSys(t, expected, actual)
}

func TestLocalizePluginsInlineAndFile(t *testing.T) {
	for _, test := range []struct {
		name  string
		files map[string]string
	}{
		{
			name: "generators",
			files: map[string]string{
				"kustomization.yaml": `generators:
- generator.yaml
- |
  apiVersion: builtin
  env: second.properties
  kind: ConfigMapGenerator
  metadata:
    name: inline
`,
				"generator.yaml": `apiVersion: builtin
env: first.properties
kind: ConfigMapGenerator
metadata:
  name: file
`,
				"first.properties":  "APPLE=orange",
				"second.properties": "BANANA=pear",
			},
		},
		{
			name: "transformers",
			files: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- |
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: inline
  path: patchSM-one.yaml
- patch.yaml
`,
				"patch.yaml": `apiVersion: builtin
kind: PatchTransformer
metadata:
  name: file
path: patchSM-two.yaml
`,
				"patchSM-one.yaml": podConfiguration,
				"patchSM-two.yaml": podConfiguration,
			},
		},
		{
			name: "validators",
			files: map[string]string{
				"kustomization.yaml": `validators:
- |
  apiVersion: builtin
  kind: ReplacementTransformer
  metadata:
    name: inline
  replacements:
  - path: first.yaml
- second.yaml
`,
				"first.yaml":       replacementTransformerWithPath,
				"second.yaml":      replacementTransformerWithPath,
				"replacement.yaml": replacements,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkLocalizeInTargetSuccess(t, test.files)
		})
	}
}

func TestLocalizeMultiplePluginsInEntry(t *testing.T) {
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- |
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: one
  path: patchSM-one.yaml
  ---
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: two
  path: patchSM-two.yaml
`,
		"patchSM-one.yaml": podConfiguration,
		"patchSM-two.yaml": podConfiguration,
	}
	checkLocalizeInTargetSuccess(t, kustAndPlugins)
}

func TestLocalizeCleanedPathInPath(t *testing.T) {
	const patchf = `apiVersion: builtin
kind: PatchTransformer
metadata:
  name: cleaned-path
path: %s
`
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- patch.yaml
`,
		"patch.yaml":   fmt.Sprintf(patchf, "../a/patchSM.yaml"),
		"patchSM.yaml": podConfiguration,
	}
	expected, actual := makeFileSystems(t, "/a", kustAndPlugins)

	checkRun(t, actual, "/a", "/a", "/dst")
	addFiles(t, expected, "/dst", map[string]string{
		"kustomization.yaml": kustAndPlugins["kustomization.yaml"],
		"patch.yaml":         fmt.Sprintf(patchf, "patchSM.yaml"),
		"patchSM.yaml":       kustAndPlugins["patchSM.yaml"],
	})
	checkFSys(t, expected, actual)
}

func TestLocalizeGeneratorsConfigMap(t *testing.T) {
	files := map[string]string{
		"kustomization.yaml": `generators:
- configMapGenerator
`,
		"configMapGenerator": `apiVersion: builtin
behavior: create
env: one.env
envs:
- two.env
- three.env
files:
- four.properties
- key=five.properties
kind: ConfigMapGenerator
metadata:
  name: custom-generator
options:
  disableNameSuffix: true
`,
		"one.env":         "key1=value1",
		"two.env":         "key2=value2",
		"three.env":       "key3=value3",
		"four.properties": "key4=value4",
		"five.properties": "key5=value5",
	}
	checkLocalizeInTargetSuccess(t, files)
}

func TestLocalizeGeneratorsSecret(t *testing.T) {
	files := map[string]string{
		"kustomization.yaml": `generators:
- secretGenerator
`,
		"secretGenerator": `apiVersion: builtin
env: one.env
envs:
- two.env
- three.env
files:
- four.properties
- key=five.properties
kind: SecretGenerator
literals:
- key6=value6
metadata:
  name: custom-generator
`,
		"one.env":         "key1=value1",
		"two.env":         "key2=value2",
		"three.env":       "key3=value3",
		"four.properties": "key4=value4",
		"five.properties": "key5=value5",
	}
	checkLocalizeInTargetSuccess(t, files)
}

func TestLocalizeTransformersPatch(t *testing.T) {
	kustAndPatches := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- |
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: no-path
  patch: '[{"op": "add", "path": "/path", "value": "value"}]'
  target:
    name: pod
- patch.yaml
`,
		"patch.yaml": `apiVersion: builtin
kind: PatchTransformer
metadata:
  name: path
path: patchSM.yaml
`,
		"patchSM.yaml": podConfiguration,
	}
	checkLocalizeInTargetSuccess(t, kustAndPatches)
}

func TestLocalizeTransformersPatchJson(t *testing.T) {
	kustAndPatches := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- patch.yaml
`,
		"patch.yaml": `apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: path
path: nested-patch.yaml
target:
  name: pod
  namespace: test
---
apiVersion: builtin
jsonOp: |-
  op: replace
  path: /path
  value: new value
kind: PatchJson6902Transformer
metadata:
  name: patch6902
target:
  name: deployment
`,
		"nested-patch.yaml": ` [
   {"op": "copy", "from": "/existing/path", "path": "/another/path"},
 ]
`,
	}
	checkLocalizeInTargetSuccess(t, kustAndPatches)
}

func TestLocalizeTransformersPatchSM(t *testing.T) {
	kustAndPatches := map[string]string{
		"kustomization.yaml": `transformers:
- patch.yaml
`,
		"patch.yaml": `apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: path
paths:
- nested-patch.yaml
- |-
  apiVersion: v1
  kind: Pod
  metadata:
    name: my-pod
`,
		"nested-patch.yaml": podConfiguration,
	}
	checkLocalizeInTargetSuccess(t, kustAndPatches)
}

func TestLocalizeTransformersReplacement(t *testing.T) {
	kustAndReplacements := map[string]string{
		"kustomization.yaml": `transformers:
- replacement-transformer.yaml
`,
		"replacement-transformer.yaml": replacementTransformerWithPath,
		"replacement.yaml":             replacements,
	}
	checkLocalizeInTargetSuccess(t, kustAndReplacements)
}

func TestLocalizePluginsNoPaths(t *testing.T) {
	kustAndPlugins := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
- |
  apiVersion: different
  kind: MyTransformer
  metadata:
    name: still-copied
  path: /nothing/special
- prefix.yaml
`,
		"prefix.yaml": `apiVersion: builtin
kind: PrefixTransformer
metadata:
  name: other-built-ins-still-copied
prefix: copy
`,
	}
	checkLocalizeInTargetSuccess(t, kustAndPlugins)
}

func TestLocalizeValidators(t *testing.T) {
	kustAndPlugin := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
validators:
- replacement-no-change.yaml
`,
		"replacement-no-change.yaml": replacementTransformerWithPath,
		"replacement.yaml":           replacements,
	}
	checkLocalizeInTargetSuccess(t, kustAndPlugin)
}

func TestLocalizeBuiltinPlugins_SequenceScalarEquivalence(t *testing.T) {
	kustomization := map[string]string{
		"kustomization.yaml": `transformers:
- |
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: path-should-be-scalar-but-accept-sequence
  path:
  - patchSM.yaml
`,
		"patchSM.yaml": podConfiguration,
	}
	checkLocalizeInTargetSuccess(t, kustomization)
}

func TestLocalizeBuiltinPlugins_NotResource(t *testing.T) {
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

			_, err := Run("/", "", "/dst", actual)

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

func TestLocalizeBuiltinPlugins_Errors(t *testing.T) {
	for name, test := range map[string]struct {
		files        map[string]string
		fieldSpecErr string
		locErr       string
	}{
		"file_dne": {
			files: map[string]string{
				"kustomization.yaml": `transformers:
- |
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: file-does-not-exist
  path: patchSM.yaml
`,
			},
			fieldSpecErr: "considering field 'path' of object PatchTransformer.builtin.[noGrp]/file-does-not-exist.[noNs]",
			locErr:       "invalid file reference: '/a/patchSM.yaml' doesn't exist",
		},
		"not_sequence_or_scalar": {
			files: map[string]string{
				"kustomization.yaml": `transformers:
- |
  apiVersion: builtin
  kind: PatchTransformer
  metadata:
    name: path-node-has-wrong-kind
  path:
    mappingNode: patchSM.yaml
`,
				"patchSM.yaml": podConfiguration,
			},
			fieldSpecErr: "considering field 'path' of object PatchTransformer.builtin.[noGrp]/path-node-has-wrong-kind.[noNs]",
			locErr:       "expected sequence or scalar node",
		},
	} {
		t.Run(name, func(t *testing.T) {
			expected, actual := makeFileSystems(t, "/a", test.files)
			_, err := Run("/a", "", "/dst", actual)
			const errPrefix = `unable to localize target "/a"`
			require.EqualError(t, err, fmt.Sprintf(
				"%s: %s: %s", errPrefix, test.fieldSpecErr, test.locErr))
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
			checkLocalizeInTargetSuccess(t, tc.files)
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

	checkRun(t, actual, "/alpha/beta/gamma", "/alpha", "/alpha/beta/dst")
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

func TestLocalizeBases(t *testing.T) {
	kustAndBases := map[string]string{
		"kustomization.yaml": `bases:
- b
- c/d
`,
		"b/kustomization.yaml": `kind: Kustomization
`,
		"c/d/kustomization.yaml": `kind: Kustomization
`,
	}
	checkLocalizeInTargetSuccess(t, kustAndBases)
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
	checkLocalizeInTargetSuccess(t, kustAndComponents)
}

func TestLocalizeResources(t *testing.T) {
	kustAndResources := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pod.yaml
- /a/b/pod2.yaml
- ../../alpha
`,
		"pod.yaml":  podConfiguration,
		"pod2.yaml": podConfiguration,
		"../../alpha/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
`,
	}

	// Absolute path of `/a/b/pod2.yaml` is expected to be converted to a path
	// relative to the kustomization root.
	expectedKustAndResources := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pod.yaml
- pod2.yaml
- ../../alpha
`,
		"pod.yaml":  podConfiguration,
		"pod2.yaml": podConfiguration,
		"../../alpha/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
`,
	}

	expectedFs, actualFs := makeFileSystems(t, "/a/b", kustAndResources)

	checkRun(t, actualFs, "/a/b", "/", "/localized-b")
	addFiles(t, expectedFs, "/localized-b/a/b", expectedKustAndResources)
	checkFSys(t, expectedFs, actualFs)
}

func TestLocalizePathError(t *testing.T) {
	kustAndResources := map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- b
`,
	}
	expected, actual := makeFileSystems(t, "/a", kustAndResources)

	_, err := Run("/a", "/", "", actual)

	const expectedFileErr = `invalid file reference: '/a/b' must resolve to a file`
	const expectedRootErr = `unable to localize root "b": unable to find one of 'kustomization.yaml', 'kustomization.yml' or 'Kustomization' in directory '/a/b'`
	var actualErr PathLocalizeError
	require.ErrorAs(t, err, &actualErr)
	require.Equal(t, "b", actualErr.Path)
	require.EqualError(t, actualErr.FileError, expectedFileErr)
	require.EqualError(t, actualErr.RootError, expectedRootErr)

	const expectedErrPrefix = `unable to localize target "/a": unable to localize resources entry`
	require.EqualError(t, err, fmt.Sprintf(`%s: could not localize path "b" as file: %s; could not localize path "b" as directory: %s`,
		expectedErrPrefix, expectedFileErr, expectedRootErr))

	checkFSys(t, expected, actual)
}

func TestLocalizeHelmChartInflationGenerator(t *testing.T) {
	helmKust := map[string]string{
		"kustomization.yaml": `helmChartInflationGenerator:
- chartName: nothing-to-localize
  chartRepoUrl: https://itzg.github.io/warcraft-server-charts
  releaseName: moria
- chartName: localize-values
  values: minecraftValues.yaml
  valuesLocal:
    minecraftServer:
      eula: true
  valuesMerge: replace
- chartHome: home
  chartName: copy-chartHome
`,
		"minecraftValues.yaml":               valuesFile,
		"charts/localize-values/values.yaml": valuesFile,
		"home/copy-chartHome/values.yaml":    valuesFile,
	}
	checkLocalizeInTargetSuccess(t, helmKust)
}

func TestLocalizeHelmCharts(t *testing.T) {
	for _, test := range []struct {
		name  string
		files map[string]string
	}{
		{
			name: "charts_only",
			files: map[string]string{
				"kustomization.yaml": `helmCharts:
- name: nothing-to-localize
  repo: https://helm.releases.hashicorp.com
  version: 1.0.0
- includeCRDs: true
  name: localize-valuesFile
  valuesFile: file
- additionalValuesFiles:
  - another
  - third
`,
				"file":                                   valuesFile,
				"another":                                valuesFile,
				"third":                                  valuesFile,
				"charts/nothing-to-localize/values.yaml": valuesFile,
				"charts/localize-valuesFile/values.yaml": valuesFile,
			},
		},
		{
			name: "charts_globals_no_home",
			files: map[string]string{
				"kustomization.yaml": `helmCharts:
- name: default
helmGlobals:
  configHome: .
`,
				"charts/default/values.yaml": valuesFile,
			},
		},
		{
			name: "home_only",
			files: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: home
`,
				"home/name/values.yaml": valuesFile,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkLocalizeInTargetSuccess(t, test.files)
		})
	}
}

func TestLocalizeHelmChartsNoDefault(t *testing.T) {
	files := map[string]string{
		"kustomization.yaml": `helmGlobals:
  chartHome: home
`,
		"home/name/values.yaml":   valuesFile,
		"charts/name/values.yaml": valuesFile,
	}
	expected, actual := makeFileSystems(t, "/a", files)

	checkRun(t, actual, "/a", "/a", "/dst")
	addFiles(t, expected, "/dst", map[string]string{
		"kustomization.yaml":    files["kustomization.yaml"],
		"home/name/values.yaml": valuesFile,
	})
	checkFSys(t, expected, actual)
}

func TestCopyChartHomeSimple(t *testing.T) {
	for _, test := range []struct {
		name  string
		files map[string]string
	}{
		{
			name: "does_not_exist",
			files: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: untar-dir
`,
			},
		},
		{
			name: "chart_home_structure",
			files: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: home
`,
				"home/minecraft-3.1.3.tgz": "blah",
				"home/terraform-1.0.0.tgz": "la",
				"home/minecraft/Chart.yaml": `annotations:
  artifacthub.io/links: |
  - name: source
    url: https://minecraft.net/
`,
				"home/terraform/Chart.yaml": `description: Minecraft server
`,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkLocalizeInTargetSuccess(t, test.files)
		})
	}
}

func TestCopyChartHomeChanges(t *testing.T) {
	for name, test := range map[string]struct {
		files       map[string]string
		copiedFiles map[string]string
	}{
		"clean_does_not_exist": {
			files: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: ../b/home
`,
			},
			copiedFiles: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: home
`,
			},
		},
		"clean_default": {
			files: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: ../b/charts
`,
				"charts/name/values.yaml": valuesFile,
			},
			copiedFiles: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: charts
`,
				"charts/name/values.yaml": valuesFile,
			},
		},
		"not_copied": {
			files: map[string]string{
				"kustomization.yaml": `helmCharts:
- name: name
  valuesFile: home/name/values.yaml
helmGlobals:
  chartHome: home
`,
				"home/name/values.yaml":      valuesFile,
				"home/name/many-other-files": "other contents",
			},
			copiedFiles: map[string]string{
				"kustomization.yaml": `helmCharts:
- name: name
  valuesFile: home/name/values.yaml
helmGlobals:
  chartHome: home
`,
				"home/name/values.yaml": valuesFile,
			},
		},
		"does_not_exist_exits_scope": {
			files: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: ../home
`,
				"../../home/will-exist-at-dst/values.yaml": valuesFile,
			},
			copiedFiles: map[string]string{
				"kustomization.yaml": `helmGlobals:
  chartHome: ../home
`,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expected, actual := makeFileSystems(t, "/a/b", test.files)

			checkRun(t, actual, "/a/b", "/a/b", "/dst")
			addFiles(t, expected, "/dst", test.copiedFiles)
			checkFSys(t, expected, actual)
		})
	}
}

func TestCopyChartHomeEmpty(t *testing.T) {
	kustomization := map[string]string{
		"kustomization.yaml": `helmGlobals:
  chartHome: home
`,
	}
	expected, actual := makeFileSystems(t, "/a", kustomization)
	require.NoError(t, actual.Mkdir("/a/home"))
	require.NoError(t, expected.Mkdir("/a/home"))

	checkRun(t, actual, "/a", "/a", "/dst")
	addFiles(t, expected, "/dst", kustomization)
	require.NoError(t, expected.Mkdir("/dst/home"))
	checkFSys(t, expected, actual)
}

func TestCopyChartHomeError(t *testing.T) {
	for name, test := range map[string]struct {
		err   string
		files map[string]string
	}{
		"absolute": {
			err: `unable to copy helmGlobals: absolute path "/a/b/home" not handled in alpha`,
			files: map[string]string{
				"a/b/kustomization.yaml": `helmGlobals:
  chartHome: /a/b/home
`,
				"a/b/home/name/values.yaml": valuesFile,
			},
		},
		"file": {
			err: `unable to copy helmGlobals: unable to copy home "home": invalid chart home: invalid root reference: must build at directory: '/a/b/home': file is not directory`,
			files: map[string]string{
				"a/b/kustomization.yaml": `helmGlobals:
  chartHome: home
`,
				"a/b/home": valuesFile,
			},
		},
		"scope": {
			err: `unable to copy helmGlobals: unable to copy home "../../alpha/home": invalid chart home: root "/alpha/home" outside localize scope "/a"`,
			files: map[string]string{
				"a/b/kustomization.yaml": `helmGlobals:
  chartHome: ../../alpha/home
`,
				"alpha/home/values.yaml": valuesFile,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expected, actual := makeFileSystems(t, "/", test.files)

			_, err := Run("/a/b", "/a", "/dst", actual)
			const prefix = `unable to localize target "/a/b"`
			require.EqualError(t, err, fmt.Sprintf("%s: %s", prefix, test.err))

			checkFSys(t, expected, actual)
		})
	}
}

func TestLocalizeGeneratorsHelm(t *testing.T) {
	files := map[string]string{
		"kustomization.yaml": `generators:
- default.yaml
- explicit.yaml
`,
		"default.yaml": `apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: no-explicit-references
name: minecraft
releaseName: moria
repo: https://itzg.github.io/minecraft-server-charts
version: 3.1.3
`,
		"explicit.yaml": `additionalValuesFiles:
- time.yaml
- life.yaml
- light.yaml
apiVersion: builtin
chartHome: home
kind: HelmChartInflationGenerator
metadata:
  name: explicit-references
name: mapleStory
valuesFile: mapleValues.yaml
`,
		"time.yaml":                    valuesFile,
		"life.yaml":                    valuesFile,
		"light.yaml":                   valuesFile,
		"mapleValues.yaml":             valuesFile,
		"home/mapleStory/values.yaml":  valuesFile,
		"charts/minecraft/values.yaml": valuesFile,
	}
	checkLocalizeInTargetSuccess(t, files)
}

func TestLocalizeGeneratorsNoHelm(t *testing.T) {
	files := map[string]string{
		"kustomization.yaml": `generators:
- configMap.yaml
`,
		"configMap.yaml": `apiVersion: builtin
kind: ConfigMapGenerator
literals:
- APPLE=orange
metadata:
  name: not-helm-shouldn't-copy-default-helm-chart-home
`,
		"charts/minecraft/values.yaml": valuesFile,
	}
	expected, actual := makeFileSystems(t, "/a", files)

	checkRun(t, actual, "/a", "/a", "/dst")
	addFiles(t, expected, "/dst", map[string]string{
		"kustomization.yaml": files["kustomization.yaml"],
		"configMap.yaml":     files["configMap.yaml"],
	})
	checkFSys(t, expected, actual)
}

func TestLocalizeEmpty(t *testing.T) {
	for name, kustomization := range map[string]string{
		"file": `configurations:
- ""
`,
		"root": `bases:
- ""
`,
		"resource": `resources:
- ""
`,
		"generator_file_src": `configMapGenerator:
- files:
  - ""
`,
		"patchesStrategicMerge": `patchesStrategicMerge:
- ""
`,
		"custom_transformers": `transformers:
- ""
`,
		"custom_transformer_field": `transformers:
- |
  apiVersion: builtin
  kind: PatchStrategicMergeTransformer
  metadata:
    name: empty
  paths:
  - ""
`,
	} {
		t.Run(name, func(t *testing.T) {
			checkLocalizeInTargetSuccess(t, map[string]string{
				"kustomization.yaml": kustomization,
			})
		})
	}
}
