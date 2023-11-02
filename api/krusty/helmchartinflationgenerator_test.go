// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
)

const (
	expectedHelm = `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: test-minecraft
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: test-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
`

	expectedHelmExternalDNS = `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: test
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: external-dns
    helm.sh/chart: external-dns-6.19.2
  name: test-external-dns
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: test
      app.kubernetes.io/name: external-dns
  template:
    metadata:
      annotations: null
      labels:
        app.kubernetes.io/instance: test
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/name: external-dns
        helm.sh/chart: external-dns-6.19.2
    spec:
      affinity:
        nodeAffinity: null
        podAffinity: null
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: test
                  app.kubernetes.io/name: external-dns
              topologyKey: kubernetes.io/hostname
            weight: 1
      containers:
      - args:
        - --metrics-address=:7979
        - --log-level=info
        - --log-format=text
        - --policy=upsert-only
        - --provider=aws
        - --registry=txt
        - --interval=1m
        - --source=service
        - --source=ingress
        - --aws-api-retries=3
        - --aws-zone-type=
        - --aws-batch-change-size=1000
        env:
        - name: AWS_DEFAULT_REGION
          value: us-east-1
        envFrom: null
        image: docker.io/bitnami/external-dns:0.13.4-debian-11-r14
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 2
          httpGet:
            path: /healthz
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
        name: external-dns
        ports:
        - containerPort: 7979
          name: http
        readinessProbe:
          failureThreshold: 6
          httpGet:
            path: /healthz
            port: http
          initialDelaySeconds: 5
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
        resources:
          limits: {}
          requests: {}
        volumeMounts: null
      securityContext:
        fsGroup: 1001
        runAsUser: 1001
      serviceAccountName: default
      volumes: null
`
)

func TestHelmChartInflationGeneratorOld(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmChartInflationGenerator:
- chartName: minecraft
  chartRepoUrl: https://itzg.github.io/minecraft-server-charts
  chartVersion: 3.1.3
  releaseName: test
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelm)
}

func TestHelmChartInflationGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmCharts:
- name: minecraft
  repo: https://itzg.github.io/minecraft-server-charts
  version: 3.1.3
  releaseName: test
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelm)
}

func TestHelmChartInflationGeneratorWithOciRepository(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmCharts:
- name: external-dns
  repo: oci://registry-1.docker.io/bitnamicharts
  version: 6.19.2
  releaseName: test
  valuesInline:
    crd:
      create: false
    rbac:
      create: false
    serviceAccount:
      create: false
    service:
      enabled: false

`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelmExternalDNS)
}

func TestHelmChartInflationGeneratorWithOciRepositoryWithAppendSlash(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmCharts:
- name: external-dns
  repo: oci://registry-1.docker.io/bitnamicharts/
  version: 6.19.2
  releaseName: test
  valuesInline:
    crd:
      create: false
    rbac:
      create: false
    serviceAccount:
      create: false
    service:
      enabled: false

`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelmExternalDNS)
}

// Last mile helm - show how kustomize puts helm charts into different
// namespaces with different customizations.
func TestHelmChartProdVsDev(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	dirBase := th.MkDir("base")
	dirProd := th.MkDir("prod")
	dirDev := th.MkDir("dev")
	dirBoth := th.MkDir("both")

	th.WriteK(dirBase, `
helmCharts:
- name: minecraft
  repo: https://itzg.github.io/minecraft-server-charts
  version: 3.1.3
  releaseName: test
`)
	th.WriteK(dirProd, `
namespace: prod
namePrefix: myProd-
resources:
- ../base
`)
	th.WriteK(dirDev, `
namespace: dev
namePrefix: myDev-
resources:
- ../base
`)
	th.WriteK(dirBoth, `
resources:
- ../dev
- ../prod
`)

	// Base unchanged
	m := th.Run(dirBase, th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelm)

	// Prod has a "prod" namespace and a prefix.
	m = th.Run(dirProd, th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
`)

	// Both has two namespaces.
	m = th.Run(dirBoth, th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myDev-test-minecraft
  namespace: dev
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myDev-test-minecraft
  namespace: dev
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
---
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
`)
}

func TestHelmChartInflationGeneratorMultipleValuesFiles(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyValuesFilesTestChartsIntoHarness(t, th)

	th.WriteK(th.GetRoot(), `
helmCharts:
  - name: test-chart
    releaseName: test-chart
    additionalValuesFiles:
    - charts/valuesFiles/file1.yaml
    - charts/valuesFiles/file2.yaml
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	asYaml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, string(asYaml), `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    chart: test-1.0.0
  name: my-deploy
  namespace: file-2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    spec:
      containers:
      - image: test-image-file1:file1
        imagePullPolicy: Never
---
apiVersion: apps/v1
kind: Pod
metadata:
  annotations:
    helm.sh/hook: test
  name: test-chart
`)
}

func TestHelmChartInflationGeneratorApiVersions(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyValuesFilesTestChartsIntoHarness(t, th)

	th.WriteK(th.GetRoot(), `
helmCharts:
  - name: test-chart
    releaseName: test-chart
    apiVersions:
    - foo/v1
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	asYaml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, string(asYaml), `apiVersion: foo/v1
kind: Deployment
metadata:
  labels:
    chart: test-1.0.0
  name: my-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    spec:
      containers:
      - image: test-image:v1.0.0
        imagePullPolicy: Always
---
apiVersion: foo/v1
kind: Pod
metadata:
  annotations:
    helm.sh/hook: test
  name: test-chart
`)
}

func TestHelmChartInflationGeneratorSkipTests(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyValuesFilesTestChartsIntoHarness(t, th)

	th.WriteK(th.GetRoot(), `
helmCharts:
  - name: test-chart
    releaseName: test-chart
    skipTests: true
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	asYaml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, string(asYaml), `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    chart: test-1.0.0
  name: my-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    spec:
      containers:
      - image: test-image:v1.0.0
        imagePullPolicy: Always
`)
}

func TestHelmChartInflationGeneratorNameTemplate(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyValuesFilesTestChartsIntoHarness(t, th)

	th.WriteK(th.GetRoot(), `
helmCharts:
  - name: test-chart
    nameTemplate: name-template
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	asYaml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, string(asYaml), `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    chart: test-1.0.0
  name: my-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    spec:
      containers:
      - image: test-image:v1.0.0
        imagePullPolicy: Always
---
apiVersion: apps/v1
kind: Pod
metadata:
  annotations:
    helm.sh/hook: test
  name: name-template
`)
}

// Reference: https://github.com/kubernetes-sigs/kustomize/issues/5163
func TestHelmChartInflationGeneratorForMultipleChartsDifferentVersion(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyValuesFilesTestChartsIntoHarness(t, th)

	th.WriteK(th.GetRoot(), `
namespace: default
helmCharts:
  - name: test-chart
    releaseName: test
    version: 1.0.0
    skipTests: true
  - name: minecraft
    repo: https://itzg.github.io/minecraft-server-charts
    version: 3.1.3
    releaseName: test-1
  - name: minecraft
    repo: https://itzg.github.io/minecraft-server-charts
    version: 3.1.4
    releaseName: test-2
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    chart: test-1.0.0
  name: my-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    spec:
      containers:
      - image: test-image:v1.0.0
        imagePullPolicy: Always
---
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-1-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test-1
  name: test-1-minecraft
  namespace: default
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-1-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test-1
  name: test-1-minecraft
  namespace: default
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-1-minecraft
  type: ClusterIP
---
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-2-minecraft
    chart: minecraft-3.1.4
    heritage: Helm
    release: test-2
  name: test-2-minecraft
  namespace: default
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-2-minecraft
    chart: minecraft-3.1.4
    heritage: Helm
    release: test-2
  name: test-2-minecraft
  namespace: default
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-2-minecraft
  type: ClusterIP
`)
}

func TestHelmChartInflationGeneratorForMultipleKubeVersions(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyValuesFilesTestChartsIntoHarness(t, th)

	th.WriteK(th.GetRoot(), `
namespace: default
helmCharts:
  - name: minecraft
    repo: https://itzg.github.io/minecraft-server-charts
    version: 4.11.0
    releaseName: test
    kubeVersion: "1.16"
    valuesInline:
      minecraftServer:
        extraPorts:
          - name: map
            containerPort: 8123
            protocol: TCP
            service:
              enabled: false
            ingress:
              enabled: true
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft-rcon
  namespace: default
type: Opaque
---
apiVersion: v1
data:
  cf-api-key: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft-curseforge
  namespace: default
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft
  namespace: default
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  labels:
    app: test-minecraft-map
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft-map
  namespace: default
spec:
  rules: null
`)

	th.WriteK(th.GetRoot(), `
namespace: default
helmCharts:
  - name: minecraft
    repo: https://itzg.github.io/minecraft-server-charts
    version: 4.11.0
    releaseName: test
    kubeVersion: "1.27"
    valuesInline:
      minecraftServer:
        extraPorts:
          - name: map
            containerPort: 8123
            protocol: TCP
            service:
              enabled: false
            ingress:
              enabled: true
`)

	m = th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft-rcon
  namespace: default
type: Opaque
---
apiVersion: v1
data:
  cf-api-key: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft-curseforge
  namespace: default
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft
  namespace: default
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    app: test-minecraft-map
    app.kubernetes.io/instance: test-minecraft
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/version: 4.11.0
    chart: minecraft-4.11.0
    heritage: Helm
    release: test
  name: test-minecraft-map
  namespace: default
spec:
  rules: null
`)
}

func copyValuesFilesTestChartsIntoHarness(t *testing.T, th *kusttest_test.HarnessEnhanced) {
	t.Helper()

	thDir := filepath.Join(th.GetRoot(), "charts")
	chartDir := "testdata/helmcharts"

	fs := th.GetFSys()
	require.NoError(t, fs.MkdirAll(filepath.Join(thDir, "templates")))
	require.NoError(t, copyutil.CopyDir(th.GetFSys(), chartDir, thDir))
}
