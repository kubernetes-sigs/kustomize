/*
Copyright 2018 The Kubernetes Authors.

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

package builtin

import (
	"testing"

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/pkg/loader"
)

func TestConfigMapGenerator(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "ConfigMapGenerator")

	th := kusttest_test.NewKustTestHarnessFull(
		t, "/app", loader.RestrictionRootOnly, plugin.ActivePluginConfig())

	th.WriteF("/app/devops.env", `
SERVICE_PORT=32
`)
	th.WriteF("/app/uxteam.env", `
COLOR=red
`)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: myMap
envFiles:
- devops.env
- uxteam.env
literals:
- FRUIT=apple
- VEGETABLE=carrot
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  COLOR: red
  FRUIT: apple
  SERVICE_PORT: "32"
  VEGETABLE: carrot
kind: ConfigMap
metadata:
  name: myMap
`)
}
