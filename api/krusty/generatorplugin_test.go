package krusty_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestGeneratorHashSuffixWithMergeBehavior(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	generatorFilename := "generateWithHashRequest.sh"

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources: 
- configmap.yaml
generators:
- |-
  kind: Executable
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        exec:
          path: ./`+generatorFilename+`
`)

	th.WriteF(filepath.Join(tmpDir.String(), "configmap.yaml"), `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmap
data:
  a: b
`)
	th.WriteF(filepath.Join(tmpDir.String(), generatorFilename), `#!/bin/sh

cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmap
  annotations:
    kustomize.config.k8s.io/needs-hash: "true"
    kustomize.config.k8s.io/behavior: "merge"
data:
  c: d
EOF
`)
	assert.NoError(t, os.Chmod(filepath.Join(tmpDir.String(), generatorFilename), 0777))
	m := th.Run(tmpDir.String(), o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
data:
  a: b
  c: d
kind: ConfigMap
metadata:
  name: cmap-fh478f99mk
`, string(yml))
}

func TestGeneratorHashSuffixWithReplaceBehavior(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	generatorFilename := "generateWithHashRequest.sh"

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources: 
- configmap.yaml
generators:
- |-
  kind: Executable
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        exec:
          path: ./`+generatorFilename+`
`)

	th.WriteF(filepath.Join(tmpDir.String(), "configmap.yaml"), `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmap
data:
  a: b
`)
	th.WriteF(filepath.Join(tmpDir.String(), generatorFilename), `#!/bin/sh

cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmap
  annotations:
    kustomize.config.k8s.io/needs-hash: "true"
    kustomize.config.k8s.io/behavior: "replace"
data:
  c: d
EOF
`)
	assert.NoError(t, os.Chmod(filepath.Join(tmpDir.String(), generatorFilename), 0777))
	m := th.Run(tmpDir.String(), o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
data:
  c: d
kind: ConfigMap
metadata:
  name: cmap-gbdtcf54mt
`, string(yml))
}
