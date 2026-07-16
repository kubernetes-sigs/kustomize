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

func TestSecretGeneratorEmitStringDataFromProperties(t *testing.T) {
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
  emitStringData: true
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
kind: Secret
metadata:
  name: test-secret-cbkdg78c5m
stringData:
  VAR2: "200"
type: Opaque
`)
}

// Generate Secrets similar to TestGeneratorBasics with emitStringData enabled and
func TestSecretGeneratorDoesNotTouchStringData(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
secretGenerator:
- name: test-secret
  emitStringData: true
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
  emitStringData: false
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
  VAR2: MjAw
kind: Secret
metadata:
  name: test-secret-962dt6476k
stringData:
  VAR1: "100"
type: Opaque
`)
}

func TestSecretGeneratorOverrideDataWithStringDataWorksAtKubeAPI(t *testing.T) {
	// The resulting Secret will have a duplicate key in both data and stringData
	// The stringData override will work, because the kube API considers stringData authoritative and write-only
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
secretGenerator:
- name: test-secret
  behavior: create
  envs:
  - properties
`)
	th.WriteF("base/properties", `
CHANGING=data-before
BASE=red
`)
	th.WriteK("overlay", `
resources:
- ../base
secretGenerator:
- name: test-secret
  emitStringData: true
  behavior: "merge"
  envs:
  - properties
`)
	th.WriteF("overlay/properties", `
CHANGING=stringData-after
OVERLAY=blue
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  BASE: cmVk
  CHANGING: ZGF0YS1iZWZvcmU=
kind: Secret
metadata:
  name: test-secret-c4g4kdc558
stringData:
  CHANGING: stringData-after
  OVERLAY: blue
type: Opaque
`)
}

func TestSecretGeneratorOverrideStringDataWithDataSilentlyFailsAtKubeAPI(t *testing.T) {
	// The resulting Secret will have a duplicate key in both data and stringData
	// The data override will fail, because the kube API considers the older value in stringData authoritative and write-only
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
secretGenerator:
- name: test-secret
  emitStringData: true
  behavior: create
  envs:
  - properties
`)
	th.WriteF("base/properties", `
CHANGING=stringData-before
BASE=red
`)
	th.WriteK("overlay", `
resources:
- ../base
secretGenerator:
- name: test-secret
  emitStringData: false
  behavior: "merge"
  envs:
  - properties
`)
	th.WriteF("overlay/properties", `
CHANGING=data-after
OVERLAY=blue
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  CHANGING: ZGF0YS1hZnRlcg==
  OVERLAY: Ymx1ZQ==
kind: Secret
metadata:
  name: test-secret-b65ktgckfh
stringData:
  BASE: red
  CHANGING: stringData-before
type: Opaque
`)
}

// Generate Secrets similar to TestGeneratorBasics with stringData enabled and
// disabled.
func TestSecretGeneratorEmitStringData(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namePrefix: blah-
secretGenerator:
- name: bob
  literals:
  - fruit=apple
  - vegetable=broccoli
  envs:
  - foo.env
  env: bar.env
  files:
  - passphrase=phrase.dat
  - forces.txt
- name: json
  literals:
  - 'v2=[{"path": "var/druid/segment-cache"}]'
  - >-
    druid_segmentCache_locations=[{"path":
    "var/druid/segment-cache",
    "maxSize": 32000000000,
    "freeSpacePercent": 1.0}]
- name: bob-string-data
  emitStringData: true
  literals:
  - fruit=apple
  - vegetable=broccoli
  envs:
  - foo.env
  env: bar.env
  files:
  - passphrase=phrase.dat
  - forces.txt
- name: json-string-data
  emitStringData: true
  literals:
  - 'v2=[{"path": "var/druid/segment-cache"}]'
  - >-
    druid_segmentCache_locations=[{"path":
    "var/druid/segment-cache",
    "maxSize": 32000000000,
    "freeSpacePercent": 1.0}]
`)
	th.WriteF("foo.env", `
MOUNTAIN=everest
OCEAN=pacific
`)
	th.WriteF("bar.env", `
BIRD=falcon
`)
	th.WriteF("phrase.dat", `
Life is short.
But the years are long.
Not while the evil days come not.
`)
	th.WriteF("forces.txt", `
gravitational
electromagnetic
strong nuclear
weak nuclear
`)
	opts := th.MakeDefaultOptions()
	m := th.Run(".", opts)
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
data:
  BIRD: ZmFsY29u
  MOUNTAIN: ZXZlcmVzdA==
  OCEAN: cGFjaWZpYw==
  forces.txt: |
    CmdyYXZpdGF0aW9uYWwKZWxlY3Ryb21hZ25ldGljCnN0cm9uZyBudWNsZWFyCndlYWsgbn
    VjbGVhcgo=
  fruit: YXBwbGU=
  passphrase: |
    CkxpZmUgaXMgc2hvcnQuCkJ1dCB0aGUgeWVhcnMgYXJlIGxvbmcuCk5vdCB3aGlsZSB0aG
    UgZXZpbCBkYXlzIGNvbWUgbm90Lgo=
  vegetable: YnJvY2NvbGk=
kind: Secret
metadata:
  name: blah-bob-58g62h555c
type: Opaque
---
apiVersion: v1
data:
  druid_segmentCache_locations: |
    W3sicGF0aCI6ICJ2YXIvZHJ1aWQvc2VnbWVudC1jYWNoZSIsICJtYXhTaXplIjogMzIwMD
    AwMDAwMDAsICJmcmVlU3BhY2VQZXJjZW50IjogMS4wfV0=
  v2: W3sicGF0aCI6ICJ2YXIvZHJ1aWQvc2VnbWVudC1jYWNoZSJ9XQ==
kind: Secret
metadata:
  name: blah-json-5cdg9f2644
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: blah-bob-string-data-5g7kgmf529
stringData:
  BIRD: falcon
  MOUNTAIN: everest
  OCEAN: pacific
  forces.txt: |2

    gravitational
    electromagnetic
    strong nuclear
    weak nuclear
  fruit: apple
  passphrase: |2

    Life is short.
    But the years are long.
    Not while the evil days come not.
  vegetable: broccoli
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: blah-json-string-data-g7ktb2m7bm
stringData:
  druid_segmentCache_locations: '[{"path": "var/druid/segment-cache", "maxSize": 32000000000,
    "freeSpacePercent": 1.0}]'
  v2: '[{"path": "var/druid/segment-cache"}]'
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

func manyHunter2s(count int) (result []byte) {
	binaryHunter2 := []byte{
		0xff, // non-utf8
		0x68, // h
		0x75, // u
		0x6e, // n
		0x74, // t
		0x65, // e
		0x72, // r
		0x32, // 2
	}

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

func TestSecretGeneratorOverlaysBinaryDataEmitStringDataFallback(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/data.bin", string(manyHunter2s(30)))
	th.WriteK("base", `
namePrefix: p1-
secretGenerator:
- name: com1
  emitStringData: true
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

func TestSecretGeneratorMixedEmitStringDataFallback(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/data.bin", string(manyHunter2s(30)))
	th.WriteF("base/forces.txt", `
gravitational
electromagnetic
strong nuclear
weak nuclear
`)
	th.WriteK("base", `
namePrefix: p1-
secretGenerator:
- name: com1
  emitStringData: true
  behavior: create
  files:
  - data.bin
  - forces.txt
  literals:
  - fruit=Mango
  - year=2025
  - crisis=true
`)
	m := th.Run("base", th.MakeDefaultOptions())
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
  name: p1-com1-8k8d8b8g5b
stringData:
  crisis: "true"
  forces.txt: |2

    gravitational
    electromagnetic
    strong nuclear
    weak nuclear
  fruit: Mango
  year: "2025"
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
