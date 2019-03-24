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

const result = `
apiVersion: v1
data:
  FRUIT: YXBwbGU=
  MOUNTAIN: ZXZlcmVzdA==
  OCEAN: cGFjaWZpYw==
  VEGETABLE: Y2Fycm90
  foo.env: Ck1PVU5UQUlOPWV2ZXJlc3QKT0NFQU49cGFjaWZpYwo=
  passphrase: ZGF0IHBocmFzZQ==
kind: Secret
metadata:
  name: bob-kf5c9fccbt
type: Opaque
`

func writeDataFiles(th *KustTestHarness) {
	th.writeF("/app/foo.env", `
MOUNTAIN=everest
OCEAN=pacific
`)
	th.writeF("/app/phrase.dat", "dat phrase")
}

func TestBuiltinPlugins(t *testing.T) {
	th := NewKustTestHarness(t, "/app")
	th.writeK("/app", `
secretGenerator:
- name: bob
  kvSources:
  - pluginType: builtin
    name: literals
    args:
    - FRUIT=apple
    - VEGETABLE=carrot
  - pluginType: builtin
    name: files
    args:
    - foo.env
    - passphrase=phrase.dat
  - pluginType: builtin
    name: envfiles
    args:
    - foo.env
`)
	writeDataFiles(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, result)
}

func TestBuiltinIsTheDefault(t *testing.T) {
	th := NewKustTestHarness(t, "/app")
	th.writeK("/app", `
secretGenerator:
- name: bob
  kvSources:
  - name: literals
    args:
    - FRUIT=apple
    - VEGETABLE=carrot
  - name: files
    args:
    - foo.env
    - passphrase=phrase.dat
  - name: envfiles
    args:
    - foo.env
`)
	writeDataFiles(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, result)
}

func TestFilterAndTrim(t *testing.T) {
	th := NewKustTestHarness(t, "/app")
	th.writeK("/app", `
secretGenerator:
- name: filter
  kvSources:
  - name: literals
    args:
    - MYPREFIX_FRUIT=apple
    - VEGETABLE=carrot
    prefixFilter: MYPREFIX_
- name: trim
  kvSources:
  - name: literals
    args:
    - MYPREFIX_FRUIT=apple
    - VEGETABLE=carrot
    prefixFilter: MYPREFIX_
    trimPrefix: true
`)
	writeDataFiles(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  MYPREFIX_FRUIT: YXBwbGU=
kind: Secret
metadata:
  name: filter-48c9g4d876
type: Opaque
---
apiVersion: v1
data:
  FRUIT: YXBwbGU=
kind: Secret
metadata:
  name: trim-c5kkgkbk87
type: Opaque
`)
}
