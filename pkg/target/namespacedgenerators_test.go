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

package target

import (
	"testing"
)

func TestNamespacedGenerator(t *testing.T) {
	th := NewKustTestHarness(t, "/app")
	th.writeK("/app", `
apiVersion: v1beta1
kind: Kustomization
configMapGenerator:
- name: the-non-default-namespace-map
  namespace: non-default
  literals:
  - altGreeting=Good Morning from non-default namespace!
  - enableRisky="false"
- name: the-map
  literals:
  - altGreeting=Good Morning from default namespace!
  - enableRisky="false"

secretGenerator:
- name: the-non-default-namespace-secret
  namespace: non-default
  commands:
    password.txt: "echo verySecret"
- name: the-secret
  commands:
    password.txt: "echo anotherSecret"
`)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  altGreeting: Good Morning from non-default namespace!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: the-non-default-namespace-map-b6h49k7mt8
  namespace: non-default
---
apiVersion: v1
data:
  altGreeting: Good Morning from default namespace!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: the-map-4959m5tm6c
---
apiVersion: v1
data:
  password.txt: dmVyeVNlY3JldAo=
kind: Secret
metadata:
  name: the-non-default-namespace-secret-9fgdmbbk5c
  namespace: non-default
type: Opaque
---
apiVersion: v1
data:
  password.txt: YW5vdGhlclNlY3JldAo=
kind: Secret
metadata:
  name: the-secret-7dd8hcgfhk
type: Opaque
`)
}
