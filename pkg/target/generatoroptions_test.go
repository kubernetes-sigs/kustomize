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
)

func TestGeneratorOptionsWithBases(t *testing.T) {
	th := NewKustTestHarness(t, "/app/overlay")
	th.writeK("/app/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generatorOptions:
  disableNameSuffixHash: true
  labels:
    foo: bar
configMapGenerator:
- name: shouldNotHaveHash
  literals:
  - foo=bar
`)
	th.writeK("/app/overlay", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../base
generatorOptions:
  disableNameSuffixHash: false
  labels:
    fruit: apple
configMapGenerator:
- name: shouldHaveHash
  literals:
  - fruit=apple
`)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  fruit: apple
kind: ConfigMap
metadata:
  labels:
    fruit: apple
  name: shouldHaveHash-2k9hc848ff
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    foo: bar
  name: shouldNotHaveHash
`)
}
