// +build notravis

// Disabled on travis, because don't want to install helm on travis.

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

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
)

// TODO: Make this test less brittle.
//
// To test ChartInflatorExec, it downloads the latest
// stable minecraft chart, inflates it with default values,
// and demands an exact match.
// Maybe just grep for particular strings instead.
//
// This test requires having the helm binary on the PATH.
//
func TestChartInflatorExecPlugin(t *testing.T) {
	tc := NewTestEnvController(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"kustomize.config.k8s.io", "v1", "ChartInflatorExec")

	th := NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.writeK("/app", `
generators:
- chartInflatorExec.yaml
namePrefix: LOOOOOOOONG-
`)

	th.writeF("/app/chartInflatorExec.yaml", `
apiVersion: kustomize.config.k8s.io/v1
kind: ChartInflatorExec
metadata:
  name: notImportantHere
chart: minecraft
`)

	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
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
