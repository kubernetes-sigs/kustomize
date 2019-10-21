// +build notravis

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Disabled on travis, because don't want to install helm on travis.

package main_test

import (
	"regexp"
	"testing"

	"sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This test requires having the helm binary on the PATH.
//
// TODO: Download and inflate the chart, and check that
// in for the test.
func TestChartInflator(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "ChartInflator")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app")

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: ChartInflator
metadata:
  name: notImportantHere
chartName: minecraft`)

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
  name: release-name-minecraft
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
  name: release-name-minecraft-datadir
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
  name: release-name-minecraft
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
