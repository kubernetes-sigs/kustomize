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

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
)

func TestSomeServiceGeneratorPlugin(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "SomeServiceGenerator")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: SomeServiceGenerator
metadata:
  name: myGenerator
name: my-service
port: "12345"
`)
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
