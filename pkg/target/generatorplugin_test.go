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

func writeGenerator(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: someteam.example.com/v1
kind: ServiceGenerator
metadata:
  name: myServiceGenerator
service: my-service
port: "12345"
`)
}

func TestGeneratorPlugin(t *testing.T) {
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
	writeGenerator(th, "/app/serviceGenerator.yaml")
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
