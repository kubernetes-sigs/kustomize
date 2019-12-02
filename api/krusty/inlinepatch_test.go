// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func makeResourcesForPatchTest(th kusttest_test.Harness) {
	th.WriteF("/app/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        volumeMounts:
        - name: nginx-persistent-storage
          mountPath: /tmp/ps
      volumes:
      - name: nginx-persistent-storage
        emptyDir: {}
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
}

func TestStrategicMergePatchInline(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeResourcesForPatchTest(th)
	th.WriteK("/app/base", `
resources:
- deployment.yaml

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: image1
`)
	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: image1
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      volumes:
      - emptyDir: {}
        name: nginx-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
}

func TestJSONPatchInline(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeResourcesForPatchTest(th)
	th.WriteK("/app/base", `
resources:
- deployment.yaml

patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: nginx
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/image
      value: image1
`)
	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: image1
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      volumes:
      - emptyDir: {}
        name: nginx-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
}

func TestExtendedPatchInlineJSON(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeResourcesForPatchTest(th)
	th.WriteK("/app/base", `
resources:
- deployment.yaml

patches:
- target:
    kind: Deployment
    name: nginx
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/image
      value: image1
`)
	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: image1
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      volumes:
      - emptyDir: {}
        name: nginx-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
}

func TestExtendedPatchInlineYAML(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeResourcesForPatchTest(th)
	th.WriteK("/app/base", `
resources:
- deployment.yaml

patches:
- target:
    kind: Deployment
    name: nginx
  patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx
    spec:
      template:
        spec:
          containers:
            - name: nginx
              image: image1
`)
	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: image1
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      volumes:
      - emptyDir: {}
        name: nginx-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
}
