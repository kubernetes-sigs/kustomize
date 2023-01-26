// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const expectedHelm = `
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
  - name: ocp-pipeline
    version: 0.1.16
    namespace: mynamespace
    releaseName: moria
    additionalValuesFiles:
    - file1.yaml
    - file2.yaml
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	asYaml, err := m.AsYaml()
	require.NoError(t, err)
	require.Contains(t, string(asYaml), `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: moria-ocp-pipeline
  namespace: mynamespace
rules:
- apiGroups:
  - ""
  resources:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: moria-ocp-pipeline
  namespace: mynamespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: moria-ocp-pipeline
subjects:
- kind: ServiceAccount
  name: jenkins
  namespace: mynamespace`)
	require.Contains(t, string(asYaml), `def releaseNamespace = ""`)
}

func copyFileIntoHarness(th *kusttest_test.HarnessEnhanced, src, dst string) {
	bytes, _ := os.ReadFile(src)
	th.WriteF(dst, string(bytes))
}

func copyValuesFilesTestChartsIntoHarness(t *testing.T, th *kusttest_test.HarnessEnhanced) {
	t.Helper()

	thDir := filepath.Join(th.GetRoot(), "charts", "ocp-pipeline")
	chartDir := "./testdata/helmcharts/multiple-values-files/charts/ocp-pipeline"

	fs := th.GetFSys()
	require.NoError(t, fs.MkdirAll(filepath.Join(thDir, "templates")))

	copyFileIntoHarness(th, filepath.Join(chartDir, ".helmignore"), filepath.Join(thDir, ".helmignore"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "Chart.yaml"), filepath.Join(thDir, "Chart.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "README.md"), filepath.Join(thDir, "README.md"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "values.yaml"), filepath.Join(thDir, "values.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "_helpers.tpl"), filepath.Join(thDir, "templates", "_helpers.tpl"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "deploymentValuesSecret.yaml"), filepath.Join(thDir, "templates", "deploymentValuesSecret.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "NOTES.txt"), filepath.Join(thDir, "templates", "NOTES.txt"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "oc-deploy.yaml"), filepath.Join(thDir, "templates", "oc-deploy.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "oc-gitWebhookSecret.yaml"), filepath.Join(thDir, "templates", "oc-gitWebhookSecret.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "oc-imageBuild.yaml"), filepath.Join(thDir, "templates", "oc-imageBuild.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "oc-imageStream.yaml"), filepath.Join(thDir, "templates", "oc-imageStream.yaml"))
	copyFileIntoHarness(th, filepath.Join(chartDir, "templates", "rbac.yaml"), filepath.Join(thDir, "templates", "rbac.yaml"))
	copyFileIntoHarness(th, "./testdata/helmcharts/multiple-values-files/valuesFiles/file1.yaml", filepath.Join(th.GetRoot(), "file1.yaml"))
	copyFileIntoHarness(th, "./testdata/helmcharts/multiple-values-files/valuesFiles/file2.yaml", filepath.Join(th.GetRoot(), "file2.yaml"))
}
