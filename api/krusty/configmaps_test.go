// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"testing"

	"sigs.k8s.io/kustomize/api/konfig"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

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

	m := th.Run("/app", th.MakeDefaultOptionsWithProperEnableKyaml())
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
  name: blah-bob-9t25t44gg4
type: Opaque
`
	th.AssertActualEqualsExpected(
		m, konfig.IfApiMachineryElseKyaml(
			fmt.Sprintf(expFmt,
				`CmdyYXZpdGF0aW9uYWwKZWxlY3Ryb21hZ25ldGljCnN0cm9uZyBudWNsZWFyCndlYWsgbnVjbGVhcgo`,
				`CkxpZmUgaXMgc2hvcnQuCkJ1dCB0aGUgeWVhcnMgYXJlIGxvbmcuCk5vdCB3aGlsZSB0aGUgZXZpbCBkYXlzIGNvbWUgbm90Lgo`),
			fmt.Sprintf(expFmt, `|
    CmdyYXZpdGF0aW9uYWwKZWxlY3Ryb21hZ25ldGljCnN0cm9uZyBudWNsZWFyCndlYWsgbn
    VjbGVhcgo=`, `|
    CkxpZmUgaXMgc2hvcnQuCkJ1dCB0aGUgeWVhcnMgYXJlIGxvbmcuCk5vdCB3aGlsZSB0aG
    UgZXZpbCBkYXlzIGNvbWUgbm90Lgo=`)))
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
  annotations: {}
  labels: {}
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
  annotations: {}
  labels: {}
  name: p1-com1-8tc62428t2
---
apiVersion: v1
data:
  from: overlay
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
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
  annotations: {}
  labels: {}
  name: o1-cm-ft9mmdc8c6
---
apiVersion: v1
data:
  big: crunch
  foo: bar
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: cm-o2-5k95kd76ft
`)
}
