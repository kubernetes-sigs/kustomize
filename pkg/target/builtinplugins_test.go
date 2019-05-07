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

	"sigs.k8s.io/kustomize/pkg/kusttest"
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

func writeDataFiles(th *kusttest_test.KustTestHarness) {
	th.WriteF("/app/foo.env", `
MOUNTAIN=everest
OCEAN=pacific
`)
	th.WriteF("/app/phrase.dat", "dat phrase")
}

func TestBuiltinPlugins(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")
	th.WriteK("/app", `
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
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, result)
}

func TestBuiltinIsTheDefault(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")
	th.WriteK("/app", `
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
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, result)
}
