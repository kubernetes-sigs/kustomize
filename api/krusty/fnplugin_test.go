package krusty_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const generateDeploymentDotSh = `#!/bin/sh

cat <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
  annotations:
    tshirt-size: small # this injects the resource reservations
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
EOF
`

func TestFnExecGenerator(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources:
- short_secret.yaml
generators:
- gener.yaml
`)

	// Create some additional resource just to make sure everything is added
	th.WriteF(filepath.Join(tmpDir.String(), "short_secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	th.WriteF(filepath.Join(tmpDir.String(), "generateDeployment.sh"), generateDeploymentDotSh)

	assert.NoError(t, os.Chmod(filepath.Join(tmpDir.String(), "generateDeployment.sh"), 0777))
	th.WriteF(filepath.Join(tmpDir.String(), "gener.yaml"), `
kind: executable
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./generateDeployment.sh
spec:
`)

	m := th.Run(tmpDir.String(), o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    tshirt-size: small
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecGeneratorWithOverlay(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	assert.NoError(t, fSys.Mkdir(base))
	assert.NoError(t, fSys.Mkdir(prod))
	th.WriteK(base, `
resources:
- short_secret.yaml
generators:
- gener.yaml
`)
	th.WriteK(prod, `
resources:
- ../base
`)
	th.WriteF(filepath.Join(base, "short_secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	th.WriteF(filepath.Join(base, "generateDeployment.sh"), generateDeploymentDotSh)

	assert.NoError(t, os.Chmod(filepath.Join(base, "generateDeployment.sh"), 0777))
	th.WriteF(filepath.Join(base, "gener.yaml"), `
kind: executable
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./generateDeployment.sh
spec:
`)

	m := th.Run(prod, o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    tshirt-size: small
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func skipIfNoDocker(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping because docker binary wasn't found in PATH")
	}
}

func TestFnContainerGenerator(t *testing.T) {
	skipIfNoDocker(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
generators:
- gener.yaml
`)
	// Create generator config
	th.WriteF("gener.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: duplicateMeWithNewName
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-fn/starlark:v0.3.0
data:
  newName: "new"
  source: |
    def gen(resources, newName):
      newR = dict()
      newR["apiVersion"] = "v1"
      newR["kind"] = "ConfigMap"
      newR["metadata"] = dict()
      newR["metadata"]["name"] = newName
      newR["data"] = dict()
      newR["data"]["field1"] = "value1"
      resources.append(newR)
    newName = ctx.resource_list["functionConfig"]["data"]["newName"]
    gen(ctx.resource_list["items"], newName)
`)
	m := th.Run(".", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  field1: value1
kind: ConfigMap
metadata:
  name: new
`)
}

func TestFnContainerTransformer(t *testing.T) {
	skipIfNoDocker(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
resources:
- secret.yaml
transformers:
- transf.yaml
`)

	th.WriteF("secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
stringData:
  field: "value"
`)

	th.WriteF("transf.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: updateSecret
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-fn/starlark:v0.3.0
data:
  name: "secret1"
  source: |
    def upd(resources, name):
      newRes = []
      for r in resources:
        if r["metadata"]["name"] == name:
           r["stringData"]["newField"] = "newValue"
    name = ctx.resource_list["functionConfig"]["data"]["name"]
    upd(ctx.resource_list["items"], name)
`)
	m := th.Run(".", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
stringData:
  field: value
  newField: newValue
`)
}

func TestFnContainerTransformerWithConfig(t *testing.T) {
	skipIfNoDocker(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
resources:
- data1.yaml
- data2.yaml
transformers:
- label_namespace.yaml
`)

	th.WriteF("data1.yaml", `apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
`)
	th.WriteF("data2.yaml", `apiVersion: v1
kind: Namespace
metadata:
  name: another-namespace
`)

	th.WriteF("label_namespace.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  name: label_namespace
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: gcr.io/kpt-functions/label-namespace@sha256:4f030738d6d25a207641ca517916431517578bd0eb8d98a8bde04e3bb9315dcd
data:
  label_name: my-ns-name
  label_value: function-test
`)

	m := th.Run(".", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    my-ns-name: function-test
  name: my-namespace
---
apiVersion: v1
kind: Namespace
metadata:
  labels:
    my-ns-name: function-test
  name: another-namespace
`)
}

func TestFnContainerEnvVars(t *testing.T) {
	skipIfNoDocker(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
generators:
- gener.yaml
`)

	// TODO: cheange image to gcr.io/kpt-functions/templater:stable
	// when https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/103
	// is merged
	th.WriteF("gener.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/kpt-templater:0.0.1
        envs:
        - TESTTEMPLATE=value
data:
  template: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: env
    data:
      value: '{{ env "TESTTEMPLATE" }}'
`)
	m := th.Run(".", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  value: value
kind: ConfigMap
metadata:
  name: env
`)
}

func skipIfNoKubectl(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("skipping because kubectl binary wasn't found in PATH")
	}
}

func TestFnContainerKubectlBackend(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoKubectl(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.UseKubectl = true
	o.PluginConfig.FnpLoadingOptions.KubectlGlobalArgs = "--context kind-kustomize-api-test"

	th.WriteK(".", `
resources:
- secret.yaml
transformers:
- transf.yaml
`)

	th.WriteF("secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
stringData:
  field: "value"
`)

	th.WriteF("transf.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: updateSecret
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-fn/starlark:v0.3.0
data:
  name: "secret1"
  source: |
    def upd(resources, name):
      newRes = []
      for r in resources:
        if r["metadata"]["name"] == name:
           r["stringData"]["newField"] = "newValue"
    name = ctx.resource_list["functionConfig"]["data"]["name"]
    upd(ctx.resource_list["items"], name)
`)
	m := th.Run(".", o)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
stringData:
  field: value
  newField: newValue
`)
}
