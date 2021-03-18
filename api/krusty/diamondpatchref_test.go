// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

// In the following structure `top` patches `left-deploy` and `right-deploy` to reference the
// configMap in `bottom`. This test verifies `configMapRef.name` in `left-deploy` and
// `right-deploy` is modified correctly.
//
//                       top
//                 /            \
//           left                   right
//          /   \                  /    \
// left-service  bottom        bottom right-service
//         |          \       /          |
// left-deploy          bottom        right-deploy
//                        |
//                     configMap

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestConfMapNameResolutionInDiamondWithPatches(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteK("bottom", `
configMapGenerator:
- name: bottom
  literals:
  - KEY=value
`)

	th.WriteK("left-service", `
resources:
- deployment.yaml
`)

	th.WriteF("left-service/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: left-deploy
  labels:
    app.kubernetes.io/name: left-deploy
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: left-pod
  template:
    metadata:
      labels:
        app.kubernetes.io/name: left-pod
    spec:
      containers:
      - image: left-image:v1.0
        name: service
`)

	th.WriteK("right-service", `
resources:
- deployment.yaml
`)

	th.WriteF("right-service/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: right-deploy
  labels:
    app.kubernetes.io/name: right-deploy
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: right-pod
  template:
    metadata:
      labels:
        app.kubernetes.io/name: right-pod
    spec:
      containers:
      - image: right-image:v1.0
        name: service
`)

	th.WriteK("top", `
resources:
- ./left
- ./right
`)

	th.WriteK("top/left", `
resources:
- ../../left-service
- ./bottom

patches:
- target:
    kind: Deployment
    name: left-deploy
  patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ignored-by-kustomize
    spec:
      template:
        spec:
          containers:
          - name: service
            env:
            - name: MAPPED_CONFIG_ITEM
              valueFrom:
                configMapKeyRef:
                  name: left-bottom
                  key: KEY
`)

	th.WriteK("top/left/bottom", `
namePrefix: left-

resources:
- ../../../bottom
`)

	th.WriteK("top/right", `
resources:
- ../../right-service
- ./bottom

patches:
- target:
    kind: Deployment
    name: right-deploy
  patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ignored-by-kustomize
    spec:
      template:
        spec:
          containers:
          - name: service
            env:
            - name: MAPPED_CONFIG_ITEM
              valueFrom:
                configMapKeyRef:
                  name: right-bottom
                  key: KEY
`)

	th.WriteK("top/right/bottom", `
namePrefix: right-

resources:
- ../../../bottom
`)

	m := th.Run("top", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: left-deploy
  name: left-deploy
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: left-pod
  template:
    metadata:
      labels:
        app.kubernetes.io/name: left-pod
    spec:
      containers:
      - env:
        - name: MAPPED_CONFIG_ITEM
          valueFrom:
            configMapKeyRef:
              key: KEY
              name: left-bottom-9f2t6f5h6d
        image: left-image:v1.0
        name: service
---
apiVersion: v1
data:
  KEY: value
kind: ConfigMap
metadata:
  name: left-bottom-9f2t6f5h6d
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: right-deploy
  name: right-deploy
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: right-pod
  template:
    metadata:
      labels:
        app.kubernetes.io/name: right-pod
    spec:
      containers:
      - env:
        - name: MAPPED_CONFIG_ITEM
          valueFrom:
            configMapKeyRef:
              key: KEY
              name: right-bottom-9f2t6f5h6d
        image: right-image:v1.0
        name: service
---
apiVersion: v1
data:
  KEY: value
kind: ConfigMap
metadata:
  name: right-bottom-9f2t6f5h6d
`)
}
