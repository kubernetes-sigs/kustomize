// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
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
  annotations:
    port: 8080
    happy: true
    color: green
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
  annotations:
    color: green
    happy: true
    port: 8080
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

// Observation: Numbers no longer quoted
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
	opts := th.MakeDefaultOptions()
	m := th.Run("overlay", opts)
	expFmt := `apiVersion: v1
data:
  crisis: %s
  fruit: Indian Gooseberry
  month: %s
  year: %s
kind: ConfigMap
metadata:
  name: bob-%s
`
	th.AssertActualEqualsExpected(
		m, opts.IfApiMachineryElseKyaml(
			fmt.Sprintf(expFmt, `"true"`, `"12"`, `"2020"`, `bk46gm59c6`),
			fmt.Sprintf(expFmt, `true`, `12`, `2020`, `bkmtk2t2fb`)))
}

// Generate a Secret and a ConfigMap from the same data
// to compare the result.
func TestGeneratorBasics(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
namePrefix: blah-
configMapGenerator:
- name: bob
  literals:
  - fruit=apple
  - vegetable=broccoli
  envs:
  - foo.env
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
`)
	th.WriteF("/app/foo.env", `
MOUNTAIN=everest
OCEAN=pacific
`)
	th.WriteF("/app/phrase.dat", `
Life is short.
But the years are long.
Not while the evil days come not.
`)
	th.WriteF("/app/forces.txt", `
gravitational
electromagnetic
strong nuclear
weak nuclear
`)
	opts := th.MakeDefaultOptions()
	m := th.Run("/app", opts)
	expFmt := `apiVersion: v1
data:
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
  name: blah-bob-d87t8m8tgm
---
apiVersion: v1
data:
  druid_segmentCache_locations: '[{"path":  "var/druid/segment-cache",  "maxSize": 32000000000,  "freeSpacePercent": 1.0}]'
  v2: '[{"path": "var/druid/segment-cache"}]'
kind: ConfigMap
metadata:
  name: blah-json-5298bc8g99
---
apiVersion: v1
data:
  MOUNTAIN: ZXZlcmVzdA==
  OCEAN: cGFjaWZpYw==
  forces.txt: %s
  fruit: YXBwbGU=
  passphrase: %s
  vegetable: YnJvY2NvbGk=
kind: Secret
metadata:
  name: blah-bob-%s
type: Opaque
`
	th.AssertActualEqualsExpected(
		m, opts.IfApiMachineryElseKyaml(
			fmt.Sprintf(expFmt,
				`CmdyYXZpdGF0aW9uYWwKZWxlY3Ryb21hZ25ldGljCnN0cm9uZyBudWNsZWFyCndlYWsgbnVjbGVhcgo=`,
				`CkxpZmUgaXMgc2hvcnQuCkJ1dCB0aGUgeWVhcnMgYXJlIGxvbmcuCk5vdCB3aGlsZSB0aGUgZXZpbCBkYXlzIGNvbWUgbm90Lgo=`,
				`ftht6hfgmb`),
			fmt.Sprintf(expFmt, `|
    CmdyYXZpdGF0aW9uYWwKZWxlY3Ryb21hZ25ldGljCnN0cm9uZyBudWNsZWFyCndlYWsgbn
    VjbGVhcgo=`, `|
    CkxpZmUgaXMgc2hvcnQuCkJ1dCB0aGUgeWVhcnMgYXJlIGxvbmcuCk5vdCB3aGlsZSB0aG
    UgZXZpbCBkYXlzIGNvbWUgbm90Lgo=`, `9t25t44gg4`)))
}

// TODO: These should be errors instead.
func TestGeneratorRepeatsInKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
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
	th.WriteF("/app/forces.txt", `
gravitational
electromagnetic
strong nuclear
weak nuclear
`)
	th.WriteF("/app/nobility.txt", `
helium
neon
argon
krypton
xenon
radon
`)
	m := th.Run("/app", th.MakeDefaultOptions())
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

func TestGeneratorOverlays(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base1", `
namePrefix: p1-
configMapGenerator:
- name: com1
  behavior: create
  literals:
  - from=base
`)
	th.WriteK("/app/base2", `
namePrefix: p2-
configMapGenerator:
- name: com2
  behavior: create
  literals:
  - from=base
`)
	th.WriteK("/app/overlay/o1", `
resources:
- ../../base1
configMapGenerator:
- name: com1
  behavior: merge
  literals:
  - from=overlay
`)
	th.WriteK("/app/overlay/o2", `
resources:
- ../../base2
configMapGenerator:
- name: com2
  behavior: merge
  literals:
  - from=overlay
`)
	th.WriteK("/app/overlay", `
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
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
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
	th.WriteK("/app/base", `
configMapGenerator:
- name: cm
  behavior: create
  literals:
  - foo=bar
`)
	th.WriteK("/app/o1", `
resources:
- ../base
namePrefix: o1-
`)
	th.WriteK("/app/o2", `
resources:
- ../base
nameSuffix: -o2
`)
	th.WriteK("/app", `
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
	m := th.Run("/app", th.MakeDefaultOptions())
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
