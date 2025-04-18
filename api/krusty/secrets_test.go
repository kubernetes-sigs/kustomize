// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Numbers and booleans are quoted
func TestSecretGeneratorIntVsStringNoMerge(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
secretGenerator:
- name: bob
  literals:
  - fruit=Mango
  - year=2025
  - crisis=true
`)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: demo
spec:
  clusterIP: None
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
kind: Service
metadata:
  name: demo
spec:
  clusterIP: None
---
apiVersion: v1
data:
  crisis: dHJ1ZQ==
  fruit: TWFuZ28=
  year: MjAyNQ==
kind: Secret
metadata:
  name: bob-9kb2bk8b2g
type: Opaque
`)
}

func TestSecretGeneratorIntVsStringWithMerge(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
secretGenerator:
- name: bob
  literals:
  - fruit=Mango
  - year=2025
  - crisis=true
`)
	th.WriteK("overlay", `
resources:
- ../base
secretGenerator:
- name: bob
  behavior: merge
  literals:
  - month=12
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  crisis: dHJ1ZQ==
  fruit: TWFuZ28=
  month: MTI=
  year: MjAyNQ==
kind: Secret
metadata:
  name: bob-mf6f2t4b62
type: Opaque
`)
}

func TestSecretGeneratorFromProperties(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
secretGenerator:
  - name: test-secret
    behavior: create
    envs:
    - properties
`)
	th.WriteF("base/properties", `
VAR1=100
`)
	th.WriteK("overlay", `
resources:
- ../base
secretGenerator:
- name: test-secret
  behavior: "merge"
  envs:
  - properties
`)
	th.WriteF("overlay/properties", `
VAR2=200
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  VAR1: MTAw
  VAR2: MjAw
kind: Secret
metadata:
  name: test-secret-c8c6d984gb
type: Opaque
`)
}

// TODO: This should be an error instead. However, we can't strict unmarshal until we have a yaml
// lib that support case-insensitive keys and anchors.
// See https://github.com/kubernetes-sigs/kustomize/issues/5061
func TestSecretGeneratorRepeatsInKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namePrefix: blah-
secretGenerator:
- name: bob
  behavior: create
  literals:
  - bean=pinto
  - star=wolf-rayet
  literals:
  - fruit=apple
  - vegetable=broccoli
  files:
  - forces.txt
  files:
  - nobles=nobility.txt
`)
	th.WriteF("forces.txt", `
gravitational
electromagnetic
strong nuclear
weak nuclear
`)
	th.WriteF("nobility.txt", `
helium
neon
argon
krypton
xenon
radon
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  fruit: YXBwbGU=
  nobles: CmhlbGl1bQpuZW9uCmFyZ29uCmtyeXB0b24KeGVub24KcmFkb24K
  vegetable: YnJvY2NvbGk=
kind: Secret
metadata:
  name: blah-bob-7dffd4g7cf
type: Opaque
`)
}

func TestSecretIssue3393(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- cm.yaml
secretGenerator:
  - name: project
    behavior: merge
    literals:
    - ANOTHER_ENV_VARIABLE="bar"
`)
	th.WriteF("cm.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: project
type: Opaque
data:
  A_FIRST_ENV_VARIABLE: "foo"
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  A_FIRST_ENV_VARIABLE: foo
  ANOTHER_ENV_VARIABLE: YmFy
kind: Secret
metadata:
  name: project
type: Opaque
`)
}

func TestSecretGeneratorSimpleOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: p-
secretGenerator:
- name: cm
  behavior: create
  literals:
  - fruit=apple
`)
	th.WriteK("overlay", `
resources:
- ../base
secretGenerator:
- name: cm
  behavior: merge
  literals:
  - veggie=broccoli
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  fruit: YXBwbGU=
  veggie: YnJvY2NvbGk=
kind: Secret
metadata:
  name: p-cm-5fg9k7g895
type: Opaque
`)
}

var binaryHunter2 = []byte{
	0xff, // non-utf8
	0x68, // h
	0x75, // u
	0x6e, // n
	0x74, // t
	0x65, // e
	0x72, // r
	0x32, // 2
}

func manyHunter2s(count int) (result []byte) {
	for i := 0; i < count; i++ {
		result = append(result, binaryHunter2...)
	}
	return
}

func TestSecretGeneratorOverlaysBinaryData(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/data.bin", string(manyHunter2s(30)))
	th.WriteK("base", `
namePrefix: p1-
secretGenerator:
- name: com1
  behavior: create
  files:
  - data.bin
`)
	th.WriteK("overlay", `
resources:
- ../base
secretGenerator:
- name: com1
  behavior: merge
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  data.bin: |
    /2h1bnRlcjL/aHVudGVyMv9odW50ZXIy/2h1bnRlcjL/aHVudGVyMv9odW50ZXIy/2h1bn
    RlcjL/aHVudGVyMv9odW50ZXIy/2h1bnRlcjL/aHVudGVyMv9odW50ZXIy/2h1bnRlcjL/
    aHVudGVyMv9odW50ZXIy/2h1bnRlcjL/aHVudGVyMv9odW50ZXIy/2h1bnRlcjL/aHVudG
    VyMv9odW50ZXIy/2h1bnRlcjL/aHVudGVyMv9odW50ZXIy/2h1bnRlcjL/aHVudGVyMv9o
    dW50ZXIy/2h1bnRlcjL/aHVudGVyMv9odW50ZXIy
kind: Secret
metadata:
  name: p1-com1-4k9fgt2gct
type: Opaque
`)
}

func TestSecretGeneratorOverlays(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base1", `
namePrefix: p1-
secretGenerator:
- name: com1
  behavior: create
  literals:
  - from=base
`)
	th.WriteK("base2", `
namePrefix: p2-
secretGenerator:
- name: com2
  behavior: create
  literals:
  - from=base
`)
	th.WriteK("overlay/o1", `
resources:
- ../../base1
secretGenerator:
- name: com1
  behavior: merge
  literals:
  - from=overlay
`)
	th.WriteK("overlay/o2", `
resources:
- ../../base2
secretGenerator:
- name: com2
  behavior: merge
  literals:
  - from=overlay
`)
	th.WriteK("overlay", `
resources:
- o1
- o2
secretGenerator:
- name: com1
  behavior: merge
  literals:
  - foo=bar
  - baz=qux
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  baz: cXV4
  foo: YmFy
  from: b3ZlcmxheQ==
kind: Secret
metadata:
  name: p1-com1-9tg2879fh2
type: Opaque
---
apiVersion: v1
data:
  from: b3ZlcmxheQ==
kind: Secret
metadata:
  name: p2-com2-b4g8g6529g
type: Opaque
`)
}

func secretGeneratorMergeNamePrefix(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
secretGenerator:
- name: cm
  behavior: create
  literals:
  - foo=bar
`)
	th.WriteK("o1", `
resources:
- ../base
namePrefix: o1-
`)
	th.WriteK("o2", `
resources:
- ../base
nameSuffix: -o2
`)
	th.WriteK(".", `
resources:
- o1
- o2
secretGenerator:
- name: o1-cm
  behavior: merge
  literals:
  - big=bang
- name: cm-o2
  behavior: merge
  literals:
  - big=crunch
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  big: bang
  foo: bar
kind: Secret
metadata:
  name: o1-cm-ft9mmdc8c6
type: Opaque
---
apiVersion: v1
data:
  big: crunch
  foo: bar
kind: Secret
metadata:
  name: cm-o2-5k95kd76ft
type: Opaque
`)
}

func secretGeneratorLiteralNewline(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
generators:
- secrets.yaml
`)
	th.WriteF("secrets.yaml", `
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: testing
type: Opaque
literals:
  - |
    initial.txt=greetings
    everyone
  - |
    final.txt=different
    behavior
---
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
data:
  final.txt: |
    different
    behavior
  initial.txt: |
    greetings
    everyone
kind: Secret
metadata:
  name: testing-tt4769fb52
type: Opaque
`)
}

// regression test for https://github.com/kubernetes-sigs/kustomize/issues/4233
func TestSecretDataEndsWithQuotes(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
secretGenerator:
  - name: test
    literals:
      - TEST=this is a 'test'
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
data:
  TEST: dGhpcyBpcyBhICd0ZXN0Jw==
kind: Secret
metadata:
  name: test-tg88t27545
type: Opaque
`)
}

func TestSecretDataIsSingleQuote(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
secretGenerator:
  - name: test
    literals:
      - TEST='
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
data:
  TEST: Jw==
kind: Secret
metadata:
  name: test-7dthck49k5
type: Opaque
`)
}

// Regression test for https://github.com/kubernetes-sigs/kustomize/issues/5047
func TestSecretPrefixSuffix(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("kustomization.yaml", `
resources:
- a
- b
`)

	th.WriteF("a/kustomization.yaml", `
resources:
- ../common

namePrefix: a
`)

	th.WriteF("b/kustomization.yaml", `
resources:
- ../common

namePrefix: b
`)

	th.WriteF("common/kustomization.yaml", `
resources:
- service

secretGenerator:
- name: "-example-secret"
`)

	th.WriteF("common/service/deployment.yaml", `
kind: Deployment
apiVersion: apps/v1

metadata:
  name: "-"

spec:
  template:
    spec:
      containers:
      - name: app
        envFrom:
        - secretRef:
            name: "-example-secret"
`)

	th.WriteF("common/service/kustomization.yaml", `
resources:
- deployment.yaml

nameSuffix: api
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a-api
spec:
  template:
    spec:
      containers:
      - envFrom:
        - secretRef:
            name: a-example-secret-46f8b28mk5
        name: app
---
apiVersion: v1
data: {}
kind: Secret
metadata:
  name: a-example-secret-46f8b28mk5
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-api
spec:
  template:
    spec:
      containers:
      - envFrom:
        - secretRef:
            name: b-example-secret-46f8b28mk5
        name: app
---
apiVersion: v1
data: {}
kind: Secret
metadata:
  name: b-example-secret-46f8b28mk5
type: Opaque
`)
}

// Regression test for https://github.com/kubernetes-sigs/kustomize/issues/5047
func TestSecretPrefixSuffix2(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("kustomization.yaml", `
resources:
- a
- b
`)

	th.WriteF("a/kustomization.yaml", `
resources:
- ../common

namePrefix: a
`)

	th.WriteF("b/kustomization.yaml", `
resources:
- ../common

namePrefix: b
`)

	th.WriteF("common/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "-example"
spec:
  template:
    spec:
      containers:
      - name: app
        envFrom:
        - secretRef:
            name: "-example-secret"
`)

	th.WriteF("common/kustomization.yaml", `
resources:
- deployment.yaml

secretGenerator:
- name: "-example-secret"
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a-example
spec:
  template:
    spec:
      containers:
      - envFrom:
        - secretRef:
            name: a-example-secret-46f8b28mk5
        name: app
---
apiVersion: v1
data: {}
kind: Secret
metadata:
  name: a-example-secret-46f8b28mk5
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-example
spec:
  template:
    spec:
      containers:
      - envFrom:
        - secretRef:
            name: b-example-secret-46f8b28mk5
        name: app
---
apiVersion: v1
data: {}
kind: Secret
metadata:
  name: b-example-secret-46f8b28mk5
type: Opaque
`)
}
