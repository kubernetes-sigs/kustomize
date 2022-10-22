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
	t.Skip("wait for #3881")
	skipIfNoDocker(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
resources:
- short_secret.yaml
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
	// Create some additional resource just to make sure everything is added
	th.WriteF("short_secret.yaml", `
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
	m := th.Run(".", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
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
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: cockroachdb
    name: demo
  name: demo-budget
spec:
  minAvailable: 67%
  selector:
    matchLabels:
      app: cockroachdb
      name: demo
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: cockroachdb
    name: demo
  name: demo-public
spec:
  ports:
  - name: grpc
    port: 26257
    targetPort: 26257
  - name: http
    port: 8080
    targetPort: 8080
  selector:
    app: cockroachdb
    name: demo
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: _status/vars
    prometheus.io/port: "8080"
    prometheus.io/scrape: "true"
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
  labels:
    app: cockroachdb
    name: demo
  name: demo
spec:
  clusterIP: None
  ports:
  - name: grpc
    port: 26257
    targetPort: 26257
  - name: http
    port: 8080
    targetPort: 8080
  selector:
    app: cockroachdb
    name: demo
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: cockroachdb
    name: demo
  name: demo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cockroachdb
      name: demo
  serviceName: demo
  template:
    metadata:
      labels:
        app: cockroachdb
        name: demo
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - cockroachdb
              topologyKey: kubernetes.io/hostname
            weight: 100
      containers:
      - command:
        - /bin/bash
        - -ecx
        - |
          # The use of qualified `+"`hostname -f`"+` is crucial:
          # Other nodes aren't able to look up the unqualified hostname.
          CRARGS=("start" "--logtostderr" "--insecure" "--host" "$(hostname -f)" "--http-host" "0.0.0.0")
          # We only want to initialize a new cluster (by omitting the join flag)
          # if we're sure that we're the first node (i.e. index 0) and that
          # there aren't any other nodes running as part of the cluster that
          # this is supposed to be a part of (which indicates that a cluster
          # already exists and we should make sure not to create a new one).
          # It's fine to run without --join on a restart if there aren't any
          # other nodes.
          if [ ! "$(hostname)" == "cockroachdb-0" ] ||              [ -e "/cockroach/cockroach-data/cluster_exists_marker" ]
          then
            # We don't join cockroachdb in order to avoid a node attempting
            # to join itself, which currently doesn't work
            # (https://github.com/cockroachdb/cockroach/issues/9625).
            CRARGS+=("--join" "cockroachdb-public")
          fi
          exec /cockroach/cockroach ${CRARGS[*]}
        image: cockroachdb/cockroach:v1.1.0
        imagePullPolicy: IfNotPresent
        name: demo
        ports:
        - containerPort: 26257
          name: grpc
        - containerPort: 8080
          name: http
        volumeMounts:
        - mountPath: /cockroach/cockroach-data
          name: datadir
      initContainers:
      - args:
        - -on-start=/on-start.sh
        - -service=cockroachdb
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: cockroachdb/cockroach-k8s-init:0.1
        imagePullPolicy: IfNotPresent
        name: bootstrap
        volumeMounts:
        - mountPath: /cockroach/cockroach-data
          name: datadir
      terminationGracePeriodSeconds: 60
      volumes:
      - name: datadir
        persistentVolumeClaim:
          claimName: datadir
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
`)
}

func TestFnContainerTransformer(t *testing.T) {
	t.Skip("wait for #3881")
	skipIfNoDocker(t)

	// Function plugins should not need the env setup done by MakeEnhancedHarness
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
resources:
- data.yaml
transformers:
- transf1.yaml
- transf2.yaml
`)

	th.WriteF("data.yaml", `
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
`)
	// This transformer should add resource reservations based on annotation in data.yaml
	// See https://github.com/kubernetes-sigs/kustomize/tree/master/functions/examples/injection-tshirt-sizes
	th.WriteF("transf1.yaml", `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: Validator
metadata:
  name: valid
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: gcr.io/kustomize-functions/example-tshirt:v0.2.0
`)
	// This transformer will check resources without and won't do any changes
	// See https://github.com/kubernetes-sigs/kustomize/tree/master/functions/examples/validator-kubeval
	th.WriteF("transf2.yaml", `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: Kubeval
metadata:
  name: validate
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/example-validator-kubeval:v0.1.0
spec:
  strict: true
  ignoreMissingSchemas: true

  # TODO: Update this to use network/volumes features.
  # Relevant issues:
  #   - https://github.com/kubernetes-sigs/kustomize/issues/1901
  #   - https://github.com/kubernetes-sigs/kustomize/issues/1902
  kubernetesVersion: "1.16.0"
  schemaLocation: "file:///schemas"
`)
	m := th.Run(".", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
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
        resources:
          requests:
            cpu: 200m
            memory: 50M
`)
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
