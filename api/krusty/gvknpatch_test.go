// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Regression test for https://github.com/kubernetes-sigs/kustomize/issues/3280
// GVKN shouldn't change with default options
func TestKeepOriginalGVKN(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
`)
	th.WriteK(".", `
resources:
- deployment.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

// https://github.com/kubernetes-sigs/kustomize/issues/3280
// These tests document behavior that will change
func TestChangeName(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
`)
	th.WriteK(".", `
resources:
- deployment.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
  options: 
    allowNameChange: true
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestChangeKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: old-name
`)
	th.WriteK(".", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowKindChange: true
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: StatefulSet
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestChangeNameAndKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
`)
	th.WriteK(".", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
    allowKindChange: true
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

// Should be able to refer to a resource with either its
// original GVKN or its current one
func TestPatchOriginalName(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("base/patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
`)
	th.WriteK("base", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
`)
	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  replicas: 999
`)
	options := th.MakeDefaultOptions()
	m := th.Run("overlay", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchNewName(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("base/patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
`)
	th.WriteK("base", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
`)
	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
spec:
  replicas: 999
`)
	options := th.MakeDefaultOptions()
	m := th.Run("overlay", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchOriginalNameAndKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("base/patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
`)
	th.WriteK("base", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
    allowKindChange: true
`)
	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  replicas: 999
`)
	options := th.MakeDefaultOptions()
	m := th.Run("overlay", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchNewNameAndKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("base/patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
`)
	th.WriteK("base", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
    allowKindChange: true
`)
	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
spec:
  replicas: 999
`)
	options := th.MakeDefaultOptions()
	m := th.Run("overlay", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

// Use original name, but new kind
// Should fail, even after #3280 is done, because this ID is invalid
func TestPatchOriginalNameAndNewKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: old-name
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)
	th.WriteF("base/patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: new-name
`)
	th.WriteK("base", `
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
    allowKindChange: true
`)
	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: new-name
spec:
  replicas: 999
`)
	options := th.MakeDefaultOptions()
	// depPatch cannot find target with kind `Deployment` and name `new-name`
	// because the resource never had this GVKN
	assert.Error(t, th.RunWithErr("overlay", options))
}

// Here is a structure of a kustomization of two components, component1
// and component2, that both use a shared deployment definition, which
// component2 adjusts by changing the kind via a patch. This test case
// checks that it does not have a conflicting definition.
// Currently documents broken behavior.
//
//                   root
//              /            \
//  component1/overlay  component2/overlay
//             |              |
//    component1/base    component2/base
//              \            /
//                   base
//
// This is the directory layout:
//
// ├── component1
// │   ├── base
// │   │   └── kustomization.yaml
// │   └── overlay
// │       └── kustomization.yaml
// ├── component2
// │   ├── base
// │   │   └── kustomization.yaml
// │   └── overlay
// │       └── kustomization.yaml
// ├── shared
// │   ├── kustomization.yaml
// │   └── deployment.yaml
// ├── kustomization.yaml

func TestBaseReuseNameAndKindConflict(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("shared", `
resources:
  - deployment.yaml
`)
	th.WriteF("shared/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: my-deploy
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
`)

	th.WriteK("component1/base", `
resources:
  - ../../shared
`)
	th.WriteK("component1/overlay", `
resources:
  - ../base
namePrefix: overlay-
`)

	th.WriteK("component2/base", `
resources:
  - ../../shared
patches:
- path: patch.yaml
  target:
    kind: Deployment
    name: my-deploy
  options:
    allowNameChange: true
    allowKindChange: true
`)
	th.WriteF("component2/base/patch.yaml", `
apiVersion: v1
kind: StatefulSet
metadata:
  name: my-stateful-set
`)
	th.WriteK("component2/overlay", `
resources:
  - ../base
namePrefix: overlay-
`)

	th.WriteK(".", `
resources:
  - component1/overlay
  - component2/overlay
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: overlay-my-deploy
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: v1
kind: StatefulSet
metadata:
  name: overlay-my-stateful-set
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestNameReferenceAfterGvknChange(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("configMap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: old-name
`)
	th.WriteF("patch.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: new-name
`)
	th.WriteF("deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - env:
        - valueFrom:
            configMapKeyRef:
              name: old-name
              key: somekey
        envFrom:
        - configMapRef:
            name: old-name
            key: somekey
`)
	th.WriteK(".", `
resources:
- configMap.yaml
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: ConfigMap
  options:
    allowNameChange: true
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: new-name
---
apiVersion: v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - env:
        - valueFrom:
            configMapKeyRef:
              key: somekey
              name: new-name
        envFrom:
        - configMapRef:
            key: somekey
            name: new-name
`)
}

func TestNameReferenceAfterJsonPatch(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("resources.yaml", `
apiVersion: v1
data:
  bar: bar
kind: ConfigMap
metadata:
  name: cm
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: foo
spec:
  selector:
    matchLabels:
      foo: foo
  template:
    metadata:
      labels:
        foo: foo
    spec:
      containers:
      - name: foo
        image: example
        volumeMounts:
        - mountPath: /path
          name: myvol
      volumes:
      - configMap:
          name: cm
        name: myvol
`)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: foo-
resources:
- resources.yaml
patches:
- target:
    group: apps
    version: v1
    name: foo
  patch: |
    - op: replace
      path: /kind
      value: Deployment
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  bar: bar
kind: ConfigMap
metadata:
  name: foo-cm
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-foo
spec:
  selector:
    matchLabels:
      foo: foo
  template:
    metadata:
      labels:
        foo: foo
    spec:
      containers:
      - image: example
        name: foo
        volumeMounts:
        - mountPath: /path
          name: myvol
      volumes:
      - configMap:
          name: foo-cm
        name: myvol
`)
}

func TestNameReferenceAfterJsonPatchConfigMapGenerator(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("statefulset.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: foo
spec:
  selector:
    matchLabels:
      foo: foo
  template:
    metadata:
      labels:
        foo: foo
    spec:
      containers:
      - name: foo
        image: example
        volumeMounts:
        - mountPath: /path
          name: myvol
      volumes:
      - configMap:
          name: cm
        name: myvol
`)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- statefulset.yaml
patches:
- target:
    group: apps
    version: v1
    name: foo
  patch: |
    - op: replace
      path: /kind
      value: Deployment
configMapGenerator:
- name: cm
  literals:
  - bar=bar
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  selector:
    matchLabels:
      foo: foo
  template:
    metadata:
      labels:
        foo: foo
    spec:
      containers:
      - image: example
        name: foo
        volumeMounts:
        - mountPath: /path
          name: myvol
      volumes:
      - configMap:
          name: cm-8hm8224gfd
        name: myvol
---
apiVersion: v1
data:
  bar: bar
kind: ConfigMap
metadata:
  name: cm-8hm8224gfd
`)
}
