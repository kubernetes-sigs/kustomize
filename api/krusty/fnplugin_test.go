// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/krusty"
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

func TestFnExecGeneratorInBaseWithOverlay(t *testing.T) {
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

func TestFnExecGeneratorInOverlay(t *testing.T) {
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

	assert.NoError(t, os.Chmod(filepath.Join(prod, "generateDeployment.sh"), 0777))
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

func TestFnExecTransformerInBase(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	assert.NoError(t, fSys.Mkdir(base))
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

	assert.NoError(t, os.Chmod(filepath.Join(base, "krmTransformer.sh"), 0777))
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
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecTransformerInBaseWithOverlay(t *testing.T) {
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

	assert.NoError(t, os.Chmod(filepath.Join(base, "krmTransformer.sh"), 0777))
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
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestFnExecTransformerInOverlay(t *testing.T) {
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

	assert.NoError(t, os.Chmod(filepath.Join(prod, "krmTransformer.sh"), 0777))
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
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: dummyTransformed
stringData:
  foo: bar
type: Opaque
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
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
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources:
- deployment.yaml
generators:
- project-service-set.yaml
`)
	// Create generator config
	th.WriteF(filepath.Join(tmpDir.String(), "project-service-set.yaml"), `
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-fn/enable-gcp-services:v0.1.0
spec:
  services:
    - compute.googleapis.com
  projectID: foo
`)
	// Create another resource just to make sure everything is added
	th.WriteF(filepath.Join(tmpDir.String(), "deployment.yaml"), `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	m := th.Run(tmpDir.String(), o)
	actual, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  annotations:
    blueprints.cloud.google.com/ownerReference: blueprints.cloud.google.com/ProjectServiceSet/demo
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-fn/enable-gcp-services:v0.1.0
  name: demo-compute
spec:
  projectRef:
    external: foo
  resourceID: compute.googleapis.com
`, string(actual))
}

func TestFnContainerTransformer(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
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
	build := exec.Command("docker", "build", ".", "-t", "gcr.io/kustomize-functions/e2econtainerconfig")
	build.Dir = "../../cmd/config/internal/commands/e2e/e2econtainerconfig"
	assert.NoError(t, build.Run())
	m := th.Run(tmpDir.String(), o)
	actual, err := m.AsYaml()
	assert.NoError(t, err)
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
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
resources:
- data1.yaml
- data2.yaml
transformers:
- label_namespace.yaml
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "data1.yaml"), []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "data2.yaml"), []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: another-namespace
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "label_namespace.yaml"), []byte(`
apiVersion: v1
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
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	assert.NoError(t, err)
	actual, err := m.AsYaml()
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
generators:
- gener.yaml
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "gener.yaml"), []byte(`
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
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	assert.NoError(t, err)
	actual, err := m.AsYaml()
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
generators:
- gener.yaml
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "gener.yaml"), []byte(`
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
          src: "./charts"
          dst: "/tmp/charts"
helmCharts:
- name: helloworld-chart
  releaseName: test
  valuesFile: /tmp/charts/helloworld-values/values.yaml
`)))
	assert.NoError(t, fSys.MkdirAll(filepath.Join(tmpDir.String(), "charts", "helloworld-chart", "templates")))
	assert.NoError(t, fSys.MkdirAll(filepath.Join(tmpDir.String(), "charts", "helloworld-values")))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "charts", "helloworld-chart", "Chart.yaml"), []byte(`
apiVersion: v2
name: helloworld-chart
description: A Helm chart for Kubernetes
type: application
version: 0.1.0
appVersion: 1.16.0
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "charts", "helloworld-chart", "templates", "deployment.yaml"), []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: name
spec:
  replicas: {{ .Values.replicaCount }}
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "charts", "helloworld-values", "values.yaml"), []byte(`
replicaCount: 5
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	assert.NoError(t, err)
	actual, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: name
spec:
  replicas: 5
`, string(actual))
}

func TestFnContainerMountsLoadRestrictions_absolute(t *testing.T) {
	skipIfNoDocker(t)
	th := kusttest_test.MakeHarness(t)
	o := th.MakeOptionsPluginsEnabled()
	fSys := filesys.MakeFsOnDisk()
	b := MakeKustomizer(&o)
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
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
	assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
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
	assert.Error(t, err)
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
	assert.Error(t, err)
	assert.EqualError(t, err, "couldn't execute function: root working directory '/' not allowed")
}
