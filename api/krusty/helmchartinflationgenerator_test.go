// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

/*
import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

var expected string = `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft
type: Opaque
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    volume.alpha.kubernetes.io/storage-class: default
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft-datadir
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
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
  type: LoadBalancer
`

func TestHelmChartInflationGenerator(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
helmChartInflationGenerator:
- chartName: minecraft
  chartRepoUrl: https://kubernetes-charts.storage.googleapis.com
  chartVersion: v1.2.0
  releaseName: test
  releaseNamespace: testNamespace
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}


func TestHelmChartInflationGeneratorAsPlugin(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
generators:
- helm.yaml
`)

	th.WriteF("/app/helm.yaml", `
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myMap
chartName: minecraft
chartRepoUrl: https://kubernetes-charts.storage.googleapis.com
chartVersion: v1.2.0
releaseName: test
releaseNamespace: testNamespace
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}
*/
