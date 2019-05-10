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

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
)

func writeServiceGenerator(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: ServiceGenerator
metadata:
  name: myServiceGenerator
service: my-service
port: "12345"
`)
}

func TestServiceGeneratorPlugin(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "ServiceGenerator")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.WriteK("/app", `
generators:
- serviceGenerator.yaml
`)
	writeServiceGenerator(th, "/app/serviceGenerator.yaml")
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
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

func writeSecretGeneratorConfig(th *kusttest_test.KustTestHarness, root string) {
	th.WriteF(filepath.Join(root, "secretGenerator.yaml"), `
apiVersion: builtin
kind: SecretGenerator
metadata:
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
	th.WriteF(filepath.Join(root, "a.env"), `
ROUTER_PASSWORD=admin
`)
	th.WriteF(filepath.Join(root, "b.env"), `
DB_PASSWORD=iloveyou
`)
	th.WriteF(filepath.Join(root, "longsecret.txt"), `
Lorem ipsum dolor sit amet,
consectetur adipiscing elit.
`)
}

func TestSecretGenerator(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "SecretGenerator")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.WriteK("/app", `
generators:
- secretGenerator.yaml
`)
	writeSecretGeneratorConfig(th, "/app")
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  DB_PASSWORD: aWxvdmV5b3U=
  FRUIT: YXBwbGU=
  ROUTER_PASSWORD: YWRtaW4=
  VEGETABLE: Y2Fycm90
  longsecret.txt: CkxvcmVtIGlwc3VtIGRvbG9yIHNpdCBhbWV0LApjb25zZWN0ZXR1ciBhZGlwaXNjaW5nIGVsaXQuCg==
kind: Secret
metadata:
  name: mySecret-g4g4kh4f7t
type: Opaque
`)
}

func TestConfigMapGenerator(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "ConfigMapGenerator")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.WriteK("/app", `
generators:
- configmapGenerator.yaml
`)
	th.WriteF("/app/configmapGenerator.yaml", `
apiVersion: someteam.example.com/v1
kind: ConfigMapGenerator
metadata:
  name: some-random-name
argsOneLiner: "admin secret"
`)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  password: secret
  username: admin
kind: ConfigMap
metadata:
  name: example-configmap-test
`)
}
