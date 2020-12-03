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

	th.WriteF("apps/deployment.yaml", `
apiVersion: apps/v1
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

	th.WriteF("apps/patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: new-name
`)

	th.WriteK("apps", `
resources:
- deployment.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
`)

	m := th.Run("apps", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
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

	th.WriteF("apps/deployment.yaml", `
apiVersion: apps/v1
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

	th.WriteF("apps/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-name
`)

	th.WriteK("apps", `
resources:
- deployment.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// name should become `new-name`
	m := th.Run("apps", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
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

func TestChangeKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("apps/deployment.yaml", `
apiVersion: apps/v1
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

	th.WriteF("apps/patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: old-name
`)

	th.WriteK("apps", `
resources:
- deployment.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// kind should become `StatefulSet`
	m := th.Run("apps", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
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

func TestChangeNameAndKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("apps/deployment.yaml", `
apiVersion: apps/v1
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

	th.WriteF("apps/patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: new-name
`)

	th.WriteK("apps", `
resources:
- deployment.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// kind should become `StatefulSet`
	// name should become `new-name`
	m := th.Run("apps", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
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
// Should be able to refer to a resource with either its
// original GVKN or its current one
func TestPatchOriginalName(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
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
apiVersion: apps/v1
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
`)

	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: old-name
spec:
  replicas: 999
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// name should become `new-name`
	m := th.Run("overlay", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: old-name
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
apiVersion: apps/v1
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
apiVersion: apps/v1
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
`)

	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-name
spec:
  replicas: 999
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// depPatch cannot find target with the name `new-name`
	// because base/patch.yaml can't yet edit the name
	assert.Error(t, th.RunWithErr("overlay", options))
}

func TestPatchOriginalNameAndKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
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
apiVersion: apps/v1
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
`)

	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: old-name
spec:
  replicas: 999
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// kind should become `StatefulSet`
	// name should become `new-name`
	m := th.Run("overlay", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: old-name
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
apiVersion: apps/v1
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
apiVersion: apps/v1
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
`)

	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: new-name
spec:
  replicas: 999
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// depPatch cannot find target with kind `StatefulSet` and name `new-name`
	// because base/patch.yaml can't yet edit the kind or name
	assert.Error(t, th.RunWithErr("overlay", options))
}

// Use original name, but new kind
// Should fail, even after #3280 is done, because this ID is invalid
func TestPatchOriginalNameAndNewKind(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
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
apiVersion: apps/v1
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
`)

	th.WriteK("overlay", `
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-name
spec:
  replicas: 999
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

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

	th.WriteK("/app/shared", `
resources:
  - deployment.yaml
`)
	th.WriteF("/app/shared/deployment.yaml", `
apiVersion: apps/v1
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

	th.WriteK("/app/component1/base", `
resources:
  - ../../shared
`)
	th.WriteK("/app/component1/overlay", `
resources:
  - ../base

namePrefix: overlay-
`)

	th.WriteK("/app/component2/base", `
resources:
  - ../../shared

patches:
- path: patch.yaml
  target:
    kind: Deployment
    name: my-deploy
`)
	th.WriteF("/app/component2/base/patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-stateful-set
`)
	th.WriteK("/app/component2/overlay", `
resources:
  - ../base

namePrefix: overlay-
`)

	th.WriteK("/app", `
resources:
  - component1/overlay
  - component2/overlay
`)

	options := th.MakeDefaultOptions()
	options.AllowResourceIdChanges = true

	// Error occurs when app/component2/base tries to load the shared resources
	// because the kind is not (yet) allowed to change yet
	// so it loads a second Deployment with the name my-deploy
	// instead of a StatefulSet as specified by the patch.
	// Will be fixed by #3280.
	assert.Error(t, th.RunWithErr("overlay", options))
}
