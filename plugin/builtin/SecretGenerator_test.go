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

func TestSecretGenerator(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "SecretGenerator")

	th := kusttest_test.NewKustTestHarnessFull(
		t, "/app", loader.RestrictionRootOnly, plugin.ActivePluginConfig())

	th.WriteF("/app/a.env", `
ROUTER_PASSWORD=admin
`)
	th.WriteF("/app/b.env", `
DB_PASSWORD=iloveyou
`)
	th.WriteF("/app/longsecret.txt", `
Lorem ipsum dolor sit amet,
consectetur adipiscing elit.
`)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: mySecret
behavior: merge
envFiles:
- a.env
- b.env
valueFiles:
- obscure=longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  DB_PASSWORD: aWxvdmV5b3U=
  FRUIT: YXBwbGU=
  ROUTER_PASSWORD: YWRtaW4=
  VEGETABLE: Y2Fycm90
  obscure: CkxvcmVtIGlwc3VtIGRvbG9yIHNpdCBhbWV0LApjb25zZWN0ZXR1ciBhZGlwaXNjaW5nIGVsaXQuCg==
kind: Secret
metadata:
  name: mySecret
type: Opaque
`)
}
