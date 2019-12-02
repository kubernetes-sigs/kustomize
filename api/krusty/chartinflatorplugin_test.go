// +build notravis

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Disabled on travis, because don't want to install helm on travis.

package krusty_test

import (
	"regexp"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This is an example of using a helm chart as a base,
// inflating it and then customizing it with a nameprefix
// applied to all its resources.
//
// The helm chart used is downloaded from
//   https://github.com/helm/charts/tree/master/stable/minecraft
// with each test run, so it's a bit brittle as that
// chart could change obviously and break the test.
//
// This test requires having the helm binary on the PATH.
//
// TODO: Download and inflate the chart, and check that
// in for the test.
func TestChartInflatorPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "ChartInflator")
	defer th.Reset()

	th.WriteK("/app", `
generators:
- chartInflator.yaml
namePrefix: LOOOOOOOONG-
`)

	th.WriteF("/app/chartInflator.yaml", `
apiVersion: someteam.example.com/v1
kind: ChartInflator
metadata:
  name: notImportantHere
chartName: minecraft
`)

	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
	chartName := regexp.MustCompile("chart: minecraft-[0-9.]+")
	th.AssertActualEqualsExpectedWithTweak(m,
		func(x []byte) []byte {
			return chartName.ReplaceAll(x, []byte("chart: minecraft-SOMEVERSION"))
		}, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: release-name-minecraft
    chart: minecraft-SOMEVERSION
    heritage: Tiller
    release: release-name
  name: LOOOOOOOONG-release-name-minecraft
type: Opaque
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    volume.alpha.kubernetes.io/storage-class: default
  labels:
    app: release-name-minecraft
    chart: minecraft-SOMEVERSION
    heritage: Tiller
    release: release-name
  name: LOOOOOOOONG-release-name-minecraft-datadir
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
    app: release-name-minecraft
    chart: minecraft-SOMEVERSION
    heritage: Tiller
    release: release-name
  name: LOOOOOOOONG-release-name-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: release-name-minecraft
  type: LoadBalancer
`)
}
