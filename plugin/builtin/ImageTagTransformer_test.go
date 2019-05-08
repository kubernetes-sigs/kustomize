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

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/pkg/loader"
)

func TestImageTagTransformer(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "ImageTagTransformer")

	th := kusttest_test.NewKustTestHarnessFull(
		t, "/app", loader.RestrictionRootOnly, plugin.ActivePluginConfig())

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  newTag: v2
`,`
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      initContainers:
      - name: nginx2
        image: my-nginx:1.8.0
      - name: init-alpine
        image: alpine:1.8.0
      containers:
      - name: ngnix
        image: nginx:1.7.9
      - name: repliaced-with-digest
        image: foobar:1
      - name: postgresdb
        image: postgres:1.8.0
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:v2
        name: ngnix
      - image: foobar:1
        name: repliaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: my-nginx:1.8.0
        name: nginx2
      - image: alpine:1.8.0
        name: init-alpine
`)
}
