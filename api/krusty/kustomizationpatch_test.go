// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Test kustomization.yaml files with  kind: KustomizationPatch
func writeKustomizationPatchBase(th kusttest_test.Harness) {
	th.WriteK("/app/base", `
resources:
- deploy.yaml
configMapGenerator:
- name: my-configmap
  literals:	
    - testValue=1
    - otherValue=10
`)
	th.WriteF("/app/base/deploy.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
`)
}

func writeKustomizationPatchPatch(th kusttest_test.Harness) {
	th.WriteF("/app/patch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: KustomizationPatch
namePrefix: patched-
replicas:
- name: storefront
  count: 3
resources:
  - stub.yaml
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - testValue=2
    - patchValue=5
`)
	th.WriteF("/app/patch/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)
}

func writeKustomizationPatchProd(th kusttest_test.Harness) {
	th.WriteK("/app/prod", `
resources:
- ../base
- ../patch
- db
`)
	th.WriteF("/app/prod/db", `
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`)
}

// KustomizationPatch are inserted into the resource hierarchy as the parent of those
// resources that come before it in the resources list of the parent Kustomization.
func TestBasicKustomizationPatch(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeKustomizationPatchBase(th)
	writeKustomizationPatchPatch(th)
	writeKustomizationPatchProd(th)
	m := th.Run("/app/prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: patched-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  otherValue: "10"
  patchValue: "5"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: patched-my-configmap-5g55h28cdh
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
---
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`)
}

func TestMultipleKustomizationPatches(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeKustomizationPatchBase(th)
	writeKustomizationPatchPatch(th)
	writeKustomizationPatchProd(th)
	th.WriteF("/app/additionalpatch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: KustomizationPatch
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - otherValue=9
`)
	th.WriteK("/app/prod", `
resources:
- ../base
- ../patch
- ../additionalpatch
- db
`)
	m := th.Run("/app/prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: patched-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  otherValue: "9"
  patchValue: "5"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: patched-my-configmap-9fddc87cdk
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
---
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`)
}

func TestNestedKustomizationPatches(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeKustomizationPatchBase(th)
	writeKustomizationPatchPatch(th)
	writeKustomizationPatchProd(th)
	th.WriteF("/app/additionalpatch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: KustomizationPatch
resources:
- ../patch
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - otherValue=9
`)
	th.WriteK("/app/prod", `
resources:
- ../base
- ../additionalpatch
- db
`)
	m := th.Run("/app/prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: patched-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  otherValue: "9"
  patchValue: "5"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: patched-my-configmap-9fddc87cdk
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
---
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`)
}

// If a patch sets a name prefix on a base, then that base can also be separately included
// without being affected by the patch in another branch of the resource tree
func TestBasicKustomizationPatchWithRepeatedBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeKustomizationPatchBase(th)
	writeKustomizationPatchPatch(th)
	writeKustomizationPatchProd(th)
	th.WriteK("/app/repeated", `
resources:
- ../base
- ../prod
`)
	m := th.Run("/app/repeated", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
---
apiVersion: v1
data:
  otherValue: "10"
  testValue: "1"
kind: ConfigMap
metadata:
  name: my-configmap-7k9t4h74f8
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  otherValue: "10"
  patchValue: "5"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: patched-my-configmap-5g55h28cdh
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
---
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`)
}

func TestApplyingKustomizationPatchDirectlySameAsKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeKustomizationPatchBase(th)
	writeKustomizationPatchPatch(th)
	th.WriteF("/app/solopatch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: KustomizationPatch
resources:
- ../base
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - patchValue=5
    - testValue=2
`)
	m := th.Run("/app/solopatch", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
---
apiVersion: v1
data:
  otherValue: "10"
  patchValue: "5"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: my-configmap-t86ktk6tdk
`)
}
