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
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
)

func writeServiceGenerator(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: someteam.example.com/v1
kind: ServiceGenerator
metadata:
  name: myServiceGenerator
service: my-service
port: "12345"
`)
}

func TestServiceGeneratorPlugin(t *testing.T) {
	tc := NewTestEnvController(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "ServiceGenerator")

	th := NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.writeK("/app", `
generators:
- serviceGenerator.yaml
`)
	writeServiceGenerator(th, "/app/serviceGenerator.yaml")
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dev
  name: my-service
spec:
  ports:
  - port: 12345
  selector:
    app: dev
`)
}

func writeSecretGeneratorConfig(th *KustTestHarness, root string) {
	th.writeF(filepath.Join(root, "secretGenerator.yaml"), `
apiVersion: kustomize.config.k8s.io/v1
kind: SecretGenerator
metadata:
  name: secretGenerator
name: mySecret
behavior: merge
envFiles:
- a.env
- b.env
valueFiles:
- longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
`)
	th.writeF(filepath.Join(root, "a.env"), `
ROUTER_PASSWORD=admin
`)
	th.writeF(filepath.Join(root, "b.env"), `
DB_PASSWORD=iloveyou
`)
	th.writeF(filepath.Join(root, "longsecret.txt"), `
Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
sed do eiusmod tempor incididunt
ut labore et dolore magna aliqua.
`)
}

// nolint:lll
func TestSecretGenerator(t *testing.T) {
	tc := NewTestEnvController(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"kustomize.config.k8s.io", "v1", "SecretGenerator")

	th := NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.writeK("/app", `
generators:
- secretGenerator.yaml
`)
	writeSecretGeneratorConfig(th, "/app")
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  FRUIT: YXBwbGU=
  ROUTER_PASSWORD: YWRtaW4=
  VEGETABLE: Y2Fycm90
  longsecret.txt: CkxvcmVtIGlwc3VtIGRvbG9yIHNpdCBhbWV0LApjb25zZWN0ZXR1ciBhZGlwaXNjaW5nIGVsaXQsCnNlZCBkbyBlaXVzbW9kIHRlbXBvciBpbmNpZGlkdW50CnV0IGxhYm9yZSBldCBkb2xvcmUgbWFnbmEgYWxpcXVhLgo=
kind: Secret
metadata:
  name: -2kt2h55789
type: Opaque
`)
}

func xTestConfigMapGenerator(t *testing.T) {
	tc := NewTestEnvController(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "ConfigMapGenerator")

	th := NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.writeK("/app", `
generators:
- configmapGenerator.yaml
`)
	th.writeF("/app/configmapGenerator.yaml", `
apiVersion: someteam.example.com/v1
kind: ConfigMapGenerator
metadata:
  name: some-random-name
`)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  password: secret
  username: admin
kind: ConfigMap
metadata:
  name: example-configmap-test
`)
}
