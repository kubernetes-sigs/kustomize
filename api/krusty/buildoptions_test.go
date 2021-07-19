// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	_ "sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestAnnoOriginLocalFiles(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
buildOptions:
  addAnnoOrigin: true
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: myService
spec:
  ports:
  - port: 7002
`)
}

func TestAnnoOriginLocalFilesWithOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: b-
resources:
- namespace.yaml
- role.yaml
- service.yaml
- deployment.yaml
`)
	th.WriteF("base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
`)
	th.WriteF("base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs
`)
	th.WriteF("base/role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: myDep
`)
	th.WriteK("prod", `
namePrefix: p-
resources:
- ../base
- service.yaml
- namespace.yaml
buildOptions:
  addAnnoOrigin: true
`)
	th.WriteF("prod/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService2
`)
	th.WriteF("prod/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs2
`)
	m := th.Run("prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/namespace.yaml
  name: myNs
---
apiVersion: v1
kind: Role
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/role.yaml
  name: p-b-myRole
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/service.yaml
  name: p-b-myService
---
apiVersion: v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/deployment.yaml
  name: p-b-myDep
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: p-myService2
---
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: namespace.yaml
  name: myNs2
`)
}

// This is a copy of TestGeneratorBasics in configmaps_test.go,
// except that we've enabled the addAnnoOrigin option
// (which doesn't do anything yet).
// TODO: Generated resources should receive the annotation
//      config.kubernetes.io/origin: |
//        generated-by: path/to/kustomization.yaml
func TestGeneratorWithAnnoOrigin(t *testing.T) {
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
