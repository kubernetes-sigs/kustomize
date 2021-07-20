// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Numbers and booleans are quoted
func TestGeneratorIntVsStringNoMerge(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
configMapGenerator:
- name: bob
  literals:
  - fruit=Indian Gooseberry
  - year=2020
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
  crisis: "true"
  fruit: Indian Gooseberry
  year: "2020"
kind: ConfigMap
metadata:
  name: bob-79t79mt227
`)
}

func TestGeneratorIntVsStringWithMerge(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
configMapGenerator:
- name: bob
  literals:
  - fruit=Indian Gooseberry
  - year=2020
  - crisis=true
`)
	th.WriteK("overlay", `
resources:
- ../base
configMapGenerator:
- name: bob
  behavior: merge
  literals:
  - month=12
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  crisis: "true"
  fruit: Indian Gooseberry
  month: "12"
  year: "2020"
kind: ConfigMap
metadata:
  name: bob-bk46gm59c6
`)
}

func TestGeneratorFromProperties(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
configMapGenerator:
  - name: test-configmap
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
configMapGenerator:
- name: test-configmap
  behavior: "merge"
  envs:
  - properties
`)
	th.WriteF("overlay/properties", `
VAR2=200
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  VAR1: "100"
  VAR2: "200"
kind: ConfigMap
metadata:
  name: test-configmap-hdghb5ddkg
`)
}

// Generate a Secret and a ConfigMap from the same data
// to compare the result.
func TestGeneratorBasics(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namePrefix: blah-
configMapGenerator:
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
secretGenerator:
- name: bob
  literals:
  - fruit=apple
  - vegetable=broccoli
  envs:
  - foo.env
  files:
  - passphrase=phrase.dat
  - forces.txt
  env: bar.env
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
kind: ConfigMap
metadata:
  name: blah-bob-g9df72cd5b
---
apiVersion: v1
data:
  druid_segmentCache_locations: '[{"path":  "var/druid/segment-cache",  "maxSize":
    32000000000,  "freeSpacePercent": 1.0}]'
  v2: '[{"path": "var/druid/segment-cache"}]'
kind: ConfigMap
metadata:
  name: blah-json-5298bc8g99
---
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
`)
}

// TODO: These should be errors instead.
func TestGeneratorRepeatsInKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namePrefix: blah-
configMapGenerator:
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
  fruit: apple
  nobles: |2

    helium
    neon
    argon
    krypton
    xenon
    radon
  vegetable: broccoli
kind: ConfigMap
metadata:
  name: blah-bob-db529cg5bk
`)
}

func TestIssue3393(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- cm.yaml
configMapGenerator:
  - name: project
    behavior: merge
    literals:
    - ANOTHER_ENV_VARIABLE="bar"
    options:
      disableNameSuffixHash: true
`)
	th.WriteF("cm.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: project
data:
  A_FIRST_ENV_VARIABLE: "foo"
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  A_FIRST_ENV_VARIABLE: foo
  ANOTHER_ENV_VARIABLE: bar
kind: ConfigMap
metadata:
  name: project
`)
}

func TestGeneratorSimpleOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: p-
configMapGenerator:
- name: cm
  behavior: create
  literals:
  - fruit=apple
`)
	th.WriteK("overlay", `
resources:
- ../base
configMapGenerator:
- name: cm
  behavior: merge
  literals:
  - veggie=broccoli
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  fruit: apple
  veggie: broccoli
kind: ConfigMap
metadata:
  name: p-cm-877mt5hc89
`)
}

var binaryHello = []byte{
	0xff, // non-utf8
	0x68, // h
	0x65, // e
	0x6c, // l
	0x6c, // l
	0x6f, // o
}

func manyHellos(count int) (result []byte) {
	for i := 0; i < count; i++ {
		result = append(result, binaryHello...)
	}
	return
}

func TestGeneratorOverlaysBinaryData(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/data.bin", string(manyHellos(30)))
	th.WriteK("base", `
namePrefix: p1-
configMapGenerator:
- name: com1
  behavior: create
  files:
  - data.bin
`)
	th.WriteK("overlay", `
resources:
- ../base
configMapGenerator:
- name: com1
  behavior: merge
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
binaryData:
  data.bin: |
    /2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbG
    xv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hl
    bGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2
    hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv
kind: ConfigMap
metadata:
  name: p1-com1-96gmmt6gt5
`)
}

func TestGeneratorOverlays(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base1", `
namePrefix: p1-
configMapGenerator:
- name: com1
  behavior: create
  literals:
  - from=base
`)
	th.WriteK("base2", `
namePrefix: p2-
configMapGenerator:
- name: com2
  behavior: create
  literals:
  - from=base
`)
	th.WriteK("overlay/o1", `
resources:
- ../../base1
configMapGenerator:
- name: com1
  behavior: merge
  literals:
  - from=overlay
`)
	th.WriteK("overlay/o2", `
resources:
- ../../base2
configMapGenerator:
- name: com2
  behavior: merge
  literals:
  - from=overlay
`)
	th.WriteK("overlay", `
resources:
- o1
- o2
configMapGenerator:
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
  baz: qux
  foo: bar
  from: overlay
kind: ConfigMap
metadata:
  name: p1-com1-8tc62428t2
---
apiVersion: v1
data:
  from: overlay
kind: ConfigMap
metadata:
  name: p2-com2-87mcggf7d7
`)
}

func TestConfigMapGeneratorMergeNamePrefix(t *testing.T) {

	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
configMapGenerator:
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
configMapGenerator:
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
kind: ConfigMap
metadata:
  name: o1-cm-ft9mmdc8c6
---
apiVersion: v1
data:
  big: crunch
  foo: bar
kind: ConfigMap
metadata:
  name: cm-o2-5k95kd76ft
`)
}

func TestConfigMapGeneratorLiteralNewline(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
generators:
- configmaps.yaml
`)
	th.WriteF("configmaps.yaml", `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: testing
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
kind: ConfigMap
metadata:
  name: testing-tt4769fb52
`)
}
