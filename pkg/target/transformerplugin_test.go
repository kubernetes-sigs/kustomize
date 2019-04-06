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

func writeDeployment(th *KustTestHarness, path string) {
	th.writeF(path, `
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

func writeStringPrefixer(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: someteam.example.com/v1
kind: StringPrefixer
metadata:
  name: myStringPrefixer
prefix: apple-
`)
}

func writeDatePrefixer(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: someteam.example.com/v1
kind: DatePrefixer
metadata:
  name: myDatePrefixer
`)
}

func TestOrderedTransformers(t *testing.T) {
	tc := NewTestEnvController(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "StringPrefixer")

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "DatePrefixer")

	th := NewKustTestHarnessWithPluginConfig(
		t, "/app", plugin.ActivePluginConfig())
	th.writeK("/app", `
resources:
- deployment.yaml
transformers:
- stringPrefixer.yaml
`)
	writeDeployment(th, "/app/deployment.yaml")
	writeStringPrefixer(th, "/app/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/datePrefixer.yaml")
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
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

func xTestTransformedTransformers(t *testing.T) {
	th := NewKustTestHarnessWithPluginConfig(
		t, "/app/overlay", plugin.ActivePluginConfig())

	th.writeK("/app/base", `
resources:
- stringPrefixer.yaml
transformers:
- datePrefixer.yaml
`)
	writeStringPrefixer(th, "/app/base/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/base/datePrefixer.yaml")

	th.writeK("/app/overlay", `
resources:
- deployment.yaml
transformers:
- ../base
`)
	writeDeployment(th, "/app/overlay/deployment.yaml")

	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
HEY
`)
}
