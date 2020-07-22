// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestOrderPreserved(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
namePrefix: b-
resources:
- namespace.yaml
- role.yaml
- service.yaml
- deployment.yaml
`)
	th.WriteF("/app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
`)
	th.WriteF("/app/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs
`)
	th.WriteF("/app/base/role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("/app/base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: myDep
`)
	th.WriteK("/app/prod", `
namePrefix: p-
resources:
- ../base
- service.yaml
- namespace.yaml
`)
	th.WriteF("/app/prod/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService2
`)
	th.WriteF("/app/prod/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs2
`)
	m := th.Run("/app/prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs
---
apiVersion: v1
kind: Role
metadata:
  name: p-b-myRole
---
apiVersion: v1
kind: Service
metadata:
  name: p-b-myService
---
apiVersion: v1
kind: Deployment
metadata:
  name: p-b-myDep
---
apiVersion: v1
kind: Service
metadata:
  name: p-myService2
---
apiVersion: v1
kind: Namespace
metadata:
  name: myNs2
`)
}

func TestBaseInResourceList(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/prod", `
namePrefix: b-
resources:
- ../base
`)
	th.WriteK("/app/base", `
namePrefix: a-
resources:
- service.yaml
`)
	th.WriteF("/app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    backend: bungie
`)
	m := th.Run("/app/prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  name: b-a-myService
spec:
  selector:
    backend: bungie
`)
}

func TestTinyOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: a-
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	th.WriteK("overlay", `
namePrefix: b-
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 999
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-a-myDeployment
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: whatever
`)
}

func writeSmallBase(th kusttest_test.Harness) {
	th.WriteK("/app/base", `
namePrefix: a-
commonLabels:
  app: myApp
resources:
- deployment.yaml
- service.yaml
`)
	th.WriteF("/app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    backend: bungie
  ports:
    - port: 7002
`)
	th.WriteF("/app/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
}

func TestSmallBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeSmallBase(th)
	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
  name: a-myDeployment
spec:
  selector:
    matchLabels:
      app: myApp
  template:
    metadata:
      labels:
        app: myApp
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
  name: a-myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    backend: bungie
`)
}

func TestSmallOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeSmallBase(th)
	th.WriteK("/app/overlay", `
namePrefix: b-
commonLabels:
  env: prod
  quotedFruit: "peach"
  quotedBoolean: "true"
resources:
- ../base
patchesStrategicMerge:
- deployment/deployment.yaml
images:
- name: whatever
  newTag: 1.8.0
`)

	th.WriteF("/app/overlay/configmap/app.env", `
DB_USERNAME=admin
DB_PASSWORD=somepw
`)
	th.WriteF("/app/overlay/configmap/app-init.ini", `
FOO=bar
BAR=baz
`)
	th.WriteF("/app/overlay/deployment/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 1000
`)
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
    env: prod
    quotedBoolean: "true"
    quotedFruit: peach
  name: b-a-myDeployment
spec:
  replicas: 1000
  selector:
    matchLabels:
      app: myApp
      env: prod
      quotedBoolean: "true"
      quotedFruit: peach
  template:
    metadata:
      labels:
        app: myApp
        backend: awesome
        env: prod
        quotedBoolean: "true"
        quotedFruit: peach
    spec:
      containers:
      - image: whatever:1.8.0
        name: whatever
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
    env: prod
    quotedBoolean: "true"
    quotedFruit: peach
  name: b-a-myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    backend: bungie
    env: prod
    quotedBoolean: "true"
    quotedFruit: peach
`)
}

func TestSharedPatchDisAllowed(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeSmallBase(th)
	th.WriteK("/app/overlay", `
commonLabels:
  env: prod
resources:
- ../base
patchesStrategicMerge:
- ../shared/deployment-patch.yaml
`)
	th.WriteF("/app/shared/deployment-patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 1000
`)
	err := th.RunWithErr("/app/overlay", func() Options {
		o := th.MakeDefaultOptions()
		o.LoadRestrictions = types.LoadRestrictionsRootOnly
		return o
	}())
	if !strings.Contains(
		err.Error(),
		"security; file '/app/shared/deployment-patch.yaml' is not in or below '/app/overlay'") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestSharedPatchAllowed(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeSmallBase(th)
	th.WriteK("/app/overlay", `
commonLabels:
  env: prod
resources:
- ../base
patchesStrategicMerge:
- ../shared/deployment-patch.yaml
`)
	th.WriteF("/app/shared/deployment-patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 1000
`)
	m := th.Run("/app/overlay", func() Options {
		o := th.MakeDefaultOptions()
		o.LoadRestrictions = types.LoadRestrictionsNone
		return o
	}())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
    env: prod
  name: a-myDeployment
spec:
  replicas: 1000
  selector:
    matchLabels:
      app: myApp
      env: prod
  template:
    metadata:
      labels:
        app: myApp
        backend: awesome
        env: prod
    spec:
      containers:
      - image: whatever
        name: whatever
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
    env: prod
  name: a-myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    backend: bungie
    env: prod
`)
}

func TestSmallOverlayJSONPatch(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeSmallBase(th)
	th.WriteK("/app/overlay", `
resources:
- ../base
patchesJson6902:
- target:
    version: v1
    kind: Service
    name: a-myService
  path: service-patch.yaml
`)

	th.WriteF("/app/overlay/service-patch.yaml", `
- op: add
  path: /spec/selector/backend
  value: beagle
`)
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
  name: a-myDeployment
spec:
  selector:
    matchLabels:
      app: myApp
  template:
    metadata:
      labels:
        app: myApp
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
  name: a-myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    backend: beagle
`)
}
