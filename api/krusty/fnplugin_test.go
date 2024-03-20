// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	repoRootDir = "../../"
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

const krmTransformerDotSh = `#!/bin/bash
cat << EOF
apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
EOF
`

func TestFnExecGeneratorInBase(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
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

	require.NoError(t, os.Chmod(filepath.Join(tmpDir.String(), "generateDeployment.sh"), 0777))
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
	yml, err := m.AsYaml()
	require.NoError(t, err)
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
	require.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecGeneratorInBaseWithOverlay(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	require.NoError(t, fSys.Mkdir(base))
	require.NoError(t, fSys.Mkdir(prod))
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

	require.NoError(t, os.Chmod(filepath.Join(base, "generateDeployment.sh"), 0777))
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
	require.NoError(t, err)
	yml, err := m.AsYaml()
	require.NoError(t, err)
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
	require.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecGeneratorInOverlay(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	require.NoError(t, fSys.Mkdir(base))
	require.NoError(t, fSys.Mkdir(prod))
	th.WriteK(base, `
resources:
- short_secret.yaml
`)
	th.WriteK(prod, `
resources:
- ../base
generators:
- gener.yaml
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
	th.WriteF(filepath.Join(prod, "generateDeployment.sh"), generateDeploymentDotSh)

	require.NoError(t, os.Chmod(filepath.Join(prod, "generateDeployment.sh"), 0777))
	th.WriteF(filepath.Join(prod, "gener.yaml"), `
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
	require.NoError(t, err)
	yml, err := m.AsYaml()
	require.NoError(t, err)
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
	require.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecTransformerInBase(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	require.NoError(t, fSys.Mkdir(base))
	th.WriteK(base, `
resources:
- secret.yaml
transformers:
- krm-transformer.yaml
`)
	th.WriteF(filepath.Join(base, "secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  name: dummy
type: Opaque
stringData:
  foo: bar
`)
	th.WriteF(filepath.Join(base, "krmTransformer.sh"), krmTransformerDotSh)

	require.NoError(t, os.Chmod(filepath.Join(base, "krmTransformer.sh"), 0777))
	th.WriteF(filepath.Join(base, "krm-transformer.yaml"), `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: MyPlugin
metadata:
  name: notImportantHere
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./krmTransformer.sh
`)

	m := th.Run(base, o)
	yml, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
`, string(yml))
	require.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecTransformerInBaseWithOverlay(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	require.NoError(t, fSys.Mkdir(base))
	require.NoError(t, fSys.Mkdir(prod))
	th.WriteK(base, `
resources:
- secret.yaml
transformers:
- krm-transformer.yaml
`)
	th.WriteK(prod, `
resources:
- ../base
`)
	th.WriteF(filepath.Join(base, "secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  name: dummy
type: Opaque
stringData:
  foo: bar
`)
	th.WriteF(filepath.Join(base, "krmTransformer.sh"), krmTransformerDotSh)

	require.NoError(t, os.Chmod(filepath.Join(base, "krmTransformer.sh"), 0777))
	th.WriteF(filepath.Join(base, "krm-transformer.yaml"), `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: MyPlugin
metadata:
  name: notImportantHere
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./krmTransformer.sh
`)

	m := th.Run(prod, o)
	yml, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
`, string(yml))
	require.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecTransformerInOverlay(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	require.NoError(t, fSys.Mkdir(base))
	require.NoError(t, fSys.Mkdir(prod))
	th.WriteK(base, `
resources:
- secret.yaml
`)
	th.WriteK(prod, `
resources:
- ../base
transformers:
- krm-transformer.yaml
`)
	th.WriteF(filepath.Join(base, "secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  name: dummy
type: Opaque
stringData:
  foo: bar
`)
	th.WriteF(filepath.Join(prod, "krmTransformer.sh"), krmTransformerDotSh)

	require.NoError(t, os.Chmod(filepath.Join(prod, "krmTransformer.sh"), 0777))
	th.WriteF(filepath.Join(prod, "krm-transformer.yaml"), `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: MyPlugin
metadata:
  name: notImportantHere
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./krmTransformer.sh
`)

	m := th.Run(prod, o)
	yml, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
`, string(yml))
	require.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping because docker binary wasn't found in PATH")
	}
}

func TestFnContainerGenerator(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources:
- deployment.yaml
generators:
- service-set.yaml
`)
	// Create generator config
	th.WriteF(filepath.Join(tmpDir.String(), "service-set.yaml"), `
apiVersion: kustomize.sigs.k8s.io/v1alpha1
kind: ServiceGenerator
metadata:
  name: simplegenerator
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/e2econtainersimplegenerator
spec:
  port: 8081
`)
	// Create another resource just to make sure everything is added
	th.WriteF(filepath.Join(tmpDir.String(), "deployment.yaml"), `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simplegenerator
`)

	build := exec.Command("docker", "build", ".",
		"-f", "./cmd/config/internal/commands/e2e/e2econtainersimplegenerator/Dockerfile",
		"-t", "gcr.io/kustomize-functions/e2econtainersimplegenerator",
	)
	build.Dir = repoRootDir
	require.NoError(t, run(build))

	m := th.Run(tmpDir.String(), o)
	actual, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: simplegenerator
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: simplegenerator
  name: simplegenerator-svc
spec:
  ports:
  - name: http
    port: 8081
    protocol: TCP
    targetPort: 8081
  selector:
    app: simplegenerator
`, string(actual))
}

func TestFnContainerTransformer(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources:
- deployment.yaml
transformers:
- e2econtainerconfig.yaml
`)
	th.WriteF(filepath.Join(tmpDir.String(), "deployment.yaml"), `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	th.WriteF(filepath.Join(tmpDir.String(), "e2econtainerconfig.yaml"), `
apiVersion: example.com/v1alpha1
kind: Input
metadata:
  name: foo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: "gcr.io/kustomize-functions/e2econtainerconfig"
`)
	build := exec.Command("docker", "build", ".",
		"-f", "./cmd/config/internal/commands/e2e/e2econtainerconfig/Dockerfile",
		"-t", "gcr.io/kustomize-functions/e2econtainerconfig",
	)
	build.Dir = repoRootDir
	require.NoError(t, run(build))
	m := th.Run(tmpDir.String(), o)
	actual, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    a-bool-value: "false"
    a-int-value: "0"
    a-string-value: ""
  name: foo
`, string(actual))
}

func TestFnContainerTransformerWithConfig(t *testing.T) {
	skipIfNoDocker(t)
	//https://docs.docker.com/engine/reference/commandline/build/#git-repositories
	build := exec.Command("docker", "build", "https://github.com/GoogleContainerTools/kpt-functions-sdk.git#go-sdk-v0.0.1:ts/hello-world",
		"-f", "build/label_namespace.Dockerfile",
		"-t", "gcr.io/kpt-functions/label-namespace:go-sdk-v0.0.1",
	)
	require.NoError(t, run(build))
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
resources:
- data1.yaml
- data2.yaml
transformers:
- label_namespace.yaml
`)))
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "data1.yaml"), []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
`)))
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "data2.yaml"), []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: another-namespace
`)))
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "label_namespace.yaml"), []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: label_namespace
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: gcr.io/kpt-functions/label-namespace:go-sdk-v0.0.1
data:
  label_name: my-ns-name
  label_value: function-test
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	require.NoError(t, err)
	actual, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
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
`, string(actual))
}

func TestFnContainerEnvVars(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
generators:
- gener.yaml
`)))
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "gener.yaml"), []byte(`
apiVersion: kustomize.sigs.k8s.io/v1alpha1
kind: EnvTemplateGenerator
metadata:
  name: e2econtainerenvgenerator
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/e2econtainerenvgenerator
        envs:
        - TESTTEMPLATE=value
template: |
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: env
  data:
    value: %q
`)))
	build := exec.Command("docker", "build", ".",
		"-f", "./cmd/config/internal/commands/e2e/e2econtainerenvgenerator/Dockerfile",
		"-t", "gcr.io/kustomize-functions/e2econtainerenvgenerator",
	)
	build.Dir = repoRootDir
	require.NoError(t, run(build))

	m, err := b.Run(
		fSys,
		tmpDir.String())
	require.NoError(t, err)
	actual, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
data:
  value: value
kind: ConfigMap
metadata:
  name: env
`, string(actual))
}

func TestFnContainerFnMounts(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
generators:
- gener.yaml
`)))
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "gener.yaml"), []byte(`
apiVersion: kustomize.sigs.k8s.io/v1alpha1
kind: RenderHelmChart
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/e2econtainermountbind
        mounts:
        - type: "bind"
          src: "./yaml"
          dst: "/tmp/yaml"
path: /tmp/yaml/resources.yaml
`)))
	require.NoError(t, fSys.MkdirAll(filepath.Join(tmpDir.String(), "yaml", "tmp")))
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "yaml", "resources.yaml"), []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: name
spec:
  replicas: 3
`)))
	build := exec.Command("docker", "build", ".",
		"-f", "./cmd/config/internal/commands/e2e/e2econtainermountbind/Dockerfile",
		"-t", "gcr.io/kustomize-functions/e2econtainermountbind",
	)
	build.Dir = repoRootDir
	require.NoError(t, run(build))

	m, err := b.Run(
		fSys,
		tmpDir.String())
	require.NoError(t, err)
	actual, err := m.AsYaml()
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: name
spec:
  replicas: 3
`, string(actual))
}

func TestFnContainerMountsLoadRestrictions_absolute(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
generators:
  - |-
    apiVersion: v1alpha1
    kind: RenderHelmChart
    metadata:
      name: demo
      annotations:
        config.kubernetes.io/function: |
          container:
            image: gcr.io/kpt-fn/render-helm-chart:v0.1.0
            mounts:
            - type: "bind"
              src: "/tmp/dir"
              dst: "/tmp/charts"
`)))
	_, err = b.Run(
		fSys,
		tmpDir.String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading generator plugins: failed to load generator: plugin RenderHelmChart."+
		"v1alpha1.[noGrp]/demo.[noNs] with mount path '/tmp/dir' is not permitted; mount paths must"+
		" be relative to the current kustomization directory")
}

func TestFnContainerMountsLoadRestrictions_outsideCurrentDir(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
generators:
  - |-
    apiVersion: v1alpha1
    kind: RenderHelmChart
    metadata:
      name: demo
      annotations:
        config.kubernetes.io/function: |
          container:
            image: gcr.io/kpt-fn/render-helm-chart:v0.1.0
            mounts:
            - type: "bind"
              src: "./tmp/../../dir"
              dst: "/tmp/charts"
`)))
	_, err = b.Run(
		fSys,
		tmpDir.String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading generator plugins: failed to load generator: plugin RenderHelmChart."+
		"v1alpha1.[noGrp]/demo.[noNs] with mount path './tmp/../../dir' is not permitted; mount paths must "+
		"be under the current kustomization directory")
}

func TestFnContainerMountsLoadRestrictions_root(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
generators:
- gener.yaml
`)
	// Create generator config
	th.WriteF("gener.yaml", `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: CockroachDB
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
spec:
  replicas: 3
`)
	err := th.RunWithErr(".", th.MakeOptionsPluginsEnabled())
	require.Error(t, err)
	assert.EqualError(t, err, "couldn't execute function: root working directory '/' not allowed")
}

// run calls Cmd.Run and wraps the error to include the output to make debugging
// easier. Not safe for real code, but fine for tests.
func run(cmd *exec.Cmd) error {
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n--- COMMAND OUTPUT ---\n%s", err, string(out))
	}
	return nil
}
