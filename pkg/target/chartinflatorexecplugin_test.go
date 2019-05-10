// +build notravis

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Disabled on travis, because don't want to install helm on travis.

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
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
func TestChartInflatorExecPlugin(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "ChartInflatorExec")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.WriteK("/app", `
generators:
- chartInflatorExec.yaml
namePrefix: LOOOOOOOONG-
`)

	th.WriteF("/app/chartInflatorExec.yaml", `
apiVersion: someteam.example.com/v1
kind: ChartInflatorExec
metadata:
  name: notImportantHere
chartName: minecraft
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: release-name-minecraft
    chart: minecraft-0.3.2
    heritage: Tiller
    release: release-name
  name: LOOOOOOOONG-release-name-minecraft
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: release-name-minecraft
    chart: minecraft-0.3.2
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
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    volume.alpha.kubernetes.io/storage-class: default
  labels:
    app: release-name-minecraft
    chart: minecraft-0.3.2
    heritage: Tiller
    release: release-name
  name: LOOOOOOOONG-release-name-minecraft-datadir
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
`)
}
