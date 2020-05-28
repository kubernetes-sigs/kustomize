// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Test kustomization.yaml files with  kind: Component
func writeComponentBase(th kusttest_test.Harness) {
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

func writeComponentPatch(th kusttest_test.Harness) {
	th.WriteF("/app/patch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
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

func writeComponentProd(th kusttest_test.Harness) {
	th.WriteK("/app/prod", `
resources:
- ../base
- db

components:
- ../patch
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

// Components are inserted into the resource hierarchy as the parent of those
// resources that come before it in the resources list of the parent Kustomization.
func TestBasicComponent(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	writeComponentProd(th)
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
  name: patched-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
`)
}

func TestMultipleComponents(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	writeComponentProd(th)
	th.WriteF("/app/additionalpatch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - otherValue=9
`)
	th.WriteK("/app/prod", `
resources:
- ../base
- db

components:
- ../patch
- ../additionalpatch
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
  name: patched-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
`)
}

func TestNestedComponents(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	writeComponentProd(th)
	th.WriteF("/app/additionalpatch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
components:
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
- db

components:
- ../additionalpatch
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
  name: patched-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
`)
}

// If a patch sets a name prefix on a base, then that base can also be separately included
// without being affected by the patch in another branch of the resource tree
func TestBasicComponentWithRepeatedBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	writeComponentProd(th)
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
  name: patched-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: patched-stub
spec:
  replicas: 1
`)
}

func TestApplyingComponentDirectlySameAsKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	th.WriteF("/app/solopatch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
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

func TestComponentsCannotBeAddedToResources(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	th.WriteF("/app/custinres/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../base
- ../patch
`)
	err := th.RunWithErr("/app/custinres", th.MakeDefaultOptions())
	if !strings.Contains(
		err.Error(),
		"expected kind != 'Component' for path '/app/patch'") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestKustomizationsCannotBeAddedToComponents(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentPatch(th)
	th.WriteF("/app/kustincomponents/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
components:
- ../base
- ../patch
`)
	err := th.RunWithErr("/app/kustincomponents", th.MakeDefaultOptions())
	if !strings.Contains(
		err.Error(),
		"accumulating components: accumulateDirectory: \"expected kind 'Component' for path '/app/base' but got 'Kustomization'") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestFilesCannotBeAddedToComponentsList(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)

	th.WriteF("/app/filesincomponents/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)

	th.WriteF("/app/filesincomponents/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
components:
- stub.yaml
- ../patch
`)
	err := th.RunWithErr("/app/filesincomponents", th.MakeDefaultOptions())
	if !strings.Contains(
		err.Error(),
		"'/app/filesincomponents/stub.yaml' must be a directory to be a root") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestMissingOptionalComponentApiVersion(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentProd(th)
	th.WriteF("/app/patch/kustomization.yaml", `
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - otherValue=9
`)

	m := th.Run("/app/prod", th.MakeDefaultOptions())
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
  otherValue: "9"
  testValue: "1"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: my-configmap-72cfg2mg5d
---
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`)
}

func TestInvalidComponentApiVersion(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeComponentBase(th)
	writeComponentProd(th)
	th.WriteF("/app/patch/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
    - otherValue=9
`)
	err := th.RunWithErr("/app/prod", th.MakeDefaultOptions())
	if !strings.Contains(
		err.Error(),
		"apiVersion for Component should be kustomize.config.k8s.io/v1alpha1") {
		t.Fatalf("unexpected error: %s", err)
	}
}
