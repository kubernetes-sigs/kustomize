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

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
)

func writeDeployment(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
}

func writeStringPrefixer(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: StringPrefixer
metadata:
  name: apple
`)
}

func writeDatePrefixer(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: DatePrefixer
metadata:
  name: irrelevant
`)
}

func TestOrderedTransformers(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "StringPrefixer")

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "DatePrefixer")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.WriteK("/app", `
resources:
- deployment.yaml
transformers:
- stringPrefixer.yaml
# - datePrefixer.yaml
`)
	// TODO(monopole): assure ordering of loaded
	// transformers and this will work - the trouble
	// is we load into a map (ResMap), not a list.
	writeDeployment(th, "/app/deployment.yaml")
	writeStringPrefixer(th, "/app/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/datePrefixer.yaml")
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apple-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
`)
}

func TestSedTransformer(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "SedTransformer")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.WriteK("/app", `
transformers:
- sed-transformer.yaml

configMapGenerator:
- name: test
  literals:
  - FOO=$FOO
  - BAR=$BAR
`)
	th.WriteF("/app/sed-transformer.yaml", `
apiVersion: someteam.example.com/v1
kind: SedTransformer
metadata:
  name: some-random-name
argsFromFile: sed-input.txt
`)
	th.WriteF("/app/sed-input.txt", `
s/$FOO/foo/g
s/$BAR/bar/g
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  BAR: bar
  FOO: foo
kind: ConfigMap
metadata:
  name: test-k4bkhftttd
`)
}

func TestTransformedTransformers(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "StringPrefixer")

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "DatePrefixer")

	th := kusttest_test.NewKustTestHarnessWithPluginConfig(
		t, "/app/overlay", plugin.ActivePluginConfig())

	th.WriteK("/app/base", `
resources:
- stringPrefixer.yaml
transformers:
- datePrefixer.yaml
`)
	writeStringPrefixer(th, "/app/base/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/base/datePrefixer.yaml")

	th.WriteK("/app/overlay", `
resources:
- deployment.yaml
transformers:
- ../base
`)
	writeDeployment(th, "/app/overlay/deployment.yaml")

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 2018-05-11-apple-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
`)
}
