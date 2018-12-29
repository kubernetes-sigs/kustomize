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

func TestGenerator1(t *testing.T) {
	th := NewKustTestHarness(t, "/app/overlay")
	th.writeK("/app/base1", `
apiVersion: v1beta1
kind: Kustomization
namePrefix: p1-
configMapGenerator:
- name: com1
  behavior: create
  literals:
    - from=base
`)
	th.writeK("/app/base2", `
apiVersion: v1beta1
kind: Kustomization
namePrefix: p2-
configMapGenerator:
- name: com2
  behavior: create
  literals:
    - from=base
`)
	th.writeK("/app/overlay/o1", `
apiVersion: v1beta1
kind: Kustomization
bases:
- ../../base1
configMapGenerator:
- name: com1
  behavior: merge
  literals:
    - from=overlay
`)
	th.writeK("/app/overlay/o2", `
apiVersion: v1beta1
kind: Kustomization
bases:
- ../../base2
configMapGenerator:
- name: com2
  behavior: merge
  literals:
    - from=overlay
`)
	th.writeK("/app/overlay", `
apiVersion: v1beta1
kind: Kustomization
bases:
- o1
- o2
configMapGenerator:
- name: com1
  behavior: merge
  literals:
    - foo=bar
    - baz=qux
`)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  baz: qux
  foo: bar
  from: overlay
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: p1-com1-dhbbm922gd
---
apiVersion: v1
data:
  from: overlay
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: p2-com2-c4b8md75k9
`)
}
