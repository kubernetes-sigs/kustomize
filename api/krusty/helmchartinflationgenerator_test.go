// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

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
