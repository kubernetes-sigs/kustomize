// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Here is a structure of a kustomization of two components, component1
// and component2, that both use a shared postgres definition, which
// they would individually adjust. This test case checks that the name
// prefix does not cause a name reference conflict.
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
// │   └── resources.yaml
// ├── kustomization.yaml

func TestBaseReuseNameConflict(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/component1/base", `
resources:
  - ../../shared

namePrefix: component1-
`)
	th.WriteK("/app/component1/overlay", `
resources:
  - ../base

namePrefix: overlay-
`)

	th.WriteK("/app/component2/base", `
resources:
  - ../../shared

namePrefix: component2-
`)
	th.WriteK("/app/component2/overlay", `
resources:
  - ../base

namePrefix: overlay-
`)

	th.WriteK("/app/shared", `
resources:
  - resources.yaml
`)
	th.WriteF("/app/shared/resources.yaml", `
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: postgres
spec:
  resources:
    requests:
      storage: 1Gi
  accessModes:
    - ReadWriteOnce
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
spec:
  selector:
    matchLabels: {}
  strategy:
    type: Recreate
  template:
    spec:
      containers:
        - name: postgres
          image: postgres
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - mountPath: /var/lib/postgresql
              name: data
          ports:
            - name: postgres
              containerPort: 5432
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: postgres
`)

	th.WriteK("/app", `
resources:
  - component1/overlay
  - component2/overlay
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: overlay-component1-postgres
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: overlay-component1-postgres
spec:
  selector:
    matchLabels: {}
  strategy:
    type: Recreate
  template:
    spec:
      containers:
      - image: postgres
        imagePullPolicy: IfNotPresent
        name: postgres
        ports:
        - containerPort: 5432
          name: postgres
        volumeMounts:
        - mountPath: /var/lib/postgresql
          name: data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: overlay-component1-postgres
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: overlay-component2-postgres
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: overlay-component2-postgres
spec:
  selector:
    matchLabels: {}
  strategy:
    type: Recreate
  template:
    spec:
      containers:
      - image: postgres
        imagePullPolicy: IfNotPresent
        name: postgres
        ports:
        - containerPort: 5432
          name: postgres
        volumeMounts:
        - mountPath: /var/lib/postgresql
          name: data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: overlay-component2-postgres
`)
}
