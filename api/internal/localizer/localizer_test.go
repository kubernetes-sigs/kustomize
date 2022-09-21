// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"bytes"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func makeMemoryFs(t *testing.T) filesys.FileSystem {
	t.Helper()
	req := require.New(t)

	fSys := filesys.MakeFsInMemory()
	req.NoError(fSys.MkdirAll("/a/b"))
	req.NoError(fSys.WriteFile("/a/pod.yaml", []byte("pod configuration")))

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

func TestPatchStrategicMergeOnFile(t *testing.T) {
	req := require.New(t)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	fSys := makeMemoryFs(t)

	// tests both inline and file patches
	// tests localize handles nested directory and winding path
	files := map[string]string{
		"kustomization.yaml": `
patchesStrategicMerge:
- ../beta/say/patch.yaml
- |-
  apiVersion: apps/v1
  metadata:
    name: myDeployment
  kind: Deployment
  spec:
    replica: 2

resources:
- localized-files`,
		// in the absence of remote references, localize directory name can be used by other files
		"localized-files": "localized-files configuration",
		"say/patch.yaml":  "patch configuration",
	}
	addFiles(t, fSys, "/alpha/beta", files)
	err := localizer.Run(fSys, "/alpha/beta", "", "/alpha/newDir")
	req.NoError(err)
	req.Empty(buf.String())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/alpha/beta", files)
	files["kustomization.yaml"] = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patchesStrategicMerge:
- say/patch.yaml
- |-
  apiVersion: apps/v1
  metadata:
    name: myDeployment
  kind: Deployment
  spec:
    replica: 2
resources:
- localized-files
`
	// directories in scope, but not referenced should not be copied to destination
	addFiles(t, fSysExpected, "/alpha/newDir", files)
	req.Equal(fSysExpected, fSys)
}

func TestComponents(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	fSys := makeMemoryFs(t)

	// components test directory references
	files := map[string]string{
		// winding directory path
		"a/kustomization.yaml": `
components:
- b/../../alpha/beta/..
- localized-files
resources:
- pod.yaml
- job.yaml
`,

		"a/job.yaml": "job configuration",

		// should recognize different kustomization names
		// inline and file replacements
		"alpha/kustomization.yml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
replacements:
- source:
    fieldPath: metadata.name
    kind: Job
  targets:
  - fieldPaths:
    - metadata.name
    select:
      kind: Pod
- path: my-replacement.yaml
`,

		"alpha/my-replacement.yaml": "replacement configuration",

		// in the absence of remote references, directories can share localize directory name
		"a/localized-files/Kustomization": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
patches:
- patch: |-
    - op: replace
      path: /spec/containers/0/name
      value: my-nginx
  target:
    kind: Pod
- path: patch.yaml
  target:
    kind: Pod
`,

		"a/localized-files/patch.yaml": "patch configuration",
	}
	addFiles(t, fSys, "/", files)

	err := localizer.Run(fSys, "/a", "/", "")
	require.NoError(t, err)
	require.Empty(t, buf.String())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/", files)

	filesExpected := map[string]string{
		"a/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
components:
- ../alpha
- localized-files
kind: Kustomization
resources:
- pod.yaml
- job.yaml
`,
		"a/pod.yaml":                           "pod configuration",
		"a/job.yaml":                           files["a/job.yaml"],
		"alpha/kustomization.yaml":             files["alpha/kustomization.yml"],
		"alpha/my-replacement.yaml":            files["alpha/my-replacement.yaml"],
		"a/localized-files/kustomization.yaml": files["a/localized-files/Kustomization"],
		"a/localized-files/patch.yaml":         files["a/localized-files/patch.yaml"],
	}
	addFiles(t, fSysExpected, "/localized-a", filesExpected)
	require.Equal(t, fSysExpected, fSys)
}

func TestNestedRoots(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	fSys := makeMemoryFs(t)
	files := map[string]string{
		// both file and directory resources
		// kustomization fields without paths should also be copied
		"beta/gamma/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: nested-roots-
resources:
- delta/deployment.yaml
- ../say
`,

		// configMapGenerator with envs and files, with both the default filename and keys
		"beta/say/kustomization.yaml": `
resources:
- ../gamma/./delta/
- ../../beta/gamma/delta/epsilon
configMapGenerator:
- name: my-config-map
  behavior: create
  files:
  - application.properties
  - environment.properties=../gamma/../say/weird-name
  literals:
  - THIS_KEY=/really/does/not/matter
  envs:
  - ./more.properties`,

		"beta/say/application.properties": "application properties",

		"beta/say/weird-name": "weird-name properties",

		"beta/say/more.properties": "more properties",

		"beta/gamma/delta/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  label: value
crds:
- epsilon/type-new-kind.yaml
kind: Kustomization
resources:
- new-kind.yaml
`,

		"beta/gamma/delta/new-kind.yaml": "new-kind configuration",

		"beta/gamma/delta/epsilon/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  label: anotherValue
crds:
- type-new-kind.yaml
kind: Kustomization
resources:
- new-kind.yaml
`,

		"beta/gamma/delta/epsilon/new-kind.yaml": "another new-kind configuration",

		// referenced more than once
		"beta/gamma/delta/epsilon/type-new-kind.yaml": "new-kind definition",
	}
	addFiles(t, fSys, "/alpha", files)
	err := localizer.Run(fSys, "/alpha/beta/gamma", "/alpha",
		"/alpha/beta/gamma/delta/newDir")
	require.NoError(t, err)
	require.Empty(t, buf.String())

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/alpha", files)
	files["beta/say/kustomization.yaml"] = `apiVersion: kustomize.config.k8s.io/v1beta1
configMapGenerator:
- behavior: create
  envs:
  - more.properties
  files:
  - application.properties
  - environment.properties=weird-name
  literals:
  - THIS_KEY=/really/does/not/matter
  name: my-config-map
kind: Kustomization
resources:
- ../gamma/delta
- ../gamma/delta/epsilon
`
	files["beta/gamma/delta/deployment.yaml"] = "deployment configuration"
	addFiles(t, fSysExpected, "/alpha/beta/gamma/delta/newDir", files)
	require.Equal(t, fSysExpected, fSys)
}

func TestExtensions(t *testing.T) {
	req := require.New(t)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	fSys := makeMemoryFs(t)
	files := map[string]string{
		"a/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
generators:
- generator.yaml
kind: Kustomization
resources:
- pod.yaml
transformers:
- transformer.yaml
`,

		"a/generator.yaml": "generator configuration",

		"a/transformer.yaml": "transformer configuration",
	}
	addFiles(t, fSys, "/", files)

	err := localizer.Run(fSys, "a", ".", "newDir")
	req.NoError(err)
	req.Contains(buf.String(), localizer.GeneratorWarning)
	req.Contains(buf.String(), localizer.TransformerWarning)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/", files)
	filesExpected := map[string]string{
		"a/kustomization.yaml": files["a/kustomization.yaml"],
		"a/pod.yaml":           "pod configuration",
	}
	addFiles(t, fSysExpected, "/newDir", filesExpected)
	req.Equal(fSysExpected, fSys)
}

func TestCleanDstOnErr(t *testing.T) {
	fSys := makeMemoryFs(t)
	files := map[string]string{
		"a/kustomization.yaml": `
resources:
- pod.yaml
- b
namePrefix: my-`,
		"b/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
invalidField: invalidValue`,
	}
	addFiles(t, fSys, "/", files)
	err := localizer.Run(fSys, "/a", "/", "/newDir")
	require.Error(t, err)

	fSysExpected := makeMemoryFs(t)
	addFiles(t, fSysExpected, "/", files)
	require.Equal(t, fSysExpected, fSys)
}
