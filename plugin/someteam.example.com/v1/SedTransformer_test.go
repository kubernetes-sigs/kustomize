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
	"sigs.k8s.io/kustomize/pkg/loader"
)

func TestSedTransformer(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin("someteam.example.com", "v1", "SedTransformer")
	th := kusttest_test.NewKustTestHarnessFull(
		t, "/app", loader.RestrictionRootOnly, plugin.ActivePluginConfig())

	th.WriteF("/app/sed-input.txt", `
s/$FRUIT/orange/g
s/$VEGGIE/tomato/g
`)

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: SedTransformer
metadata:
  name: notImportantHere
argsOneLiner: s/one/two/g
argsFromFile: sed-input.txt
`,
		`apiVersion: apps/v1
kind: MeatBall
metadata:
  name: notImportantHere
beans: one one one one
fruit: $FRUIT
vegetable: $VEGGIE
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
beans: two two two two
fruit: orange
kind: MeatBall
metadata:
  annotations: {}
  name: notImportantHere
vegetable: tomato
`)
}
