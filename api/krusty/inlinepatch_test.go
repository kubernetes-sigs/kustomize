// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func makeResourcesForPatchTest(th kusttest_test.Harness) {
	th.WriteF("base/deployment.yaml", `
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
	th.WriteK("base", `
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
	m := th.Run("base", th.MakeDefaultOptions())
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
	th.WriteK("base", `
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
	m := th.Run("base", th.MakeDefaultOptions())
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
	th.WriteK("base", `
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
	m := th.Run("base", th.MakeDefaultOptions())
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
	th.WriteK("base", `
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
	m := th.Run("base", th.MakeDefaultOptions())
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

func TestPathWithCronJobV1(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- resources.yaml
patches:
- patch: |
    apiVersion: batch/v1
    kind: CronJob
    metadata:
      name: test
    spec:
      jobTemplate:
        spec:
          template:
            spec:
              containers:
              - name: test
                env:
                - name: ENV_NEW
                  value: val_new
  target:
    kind: CronJob
    name: test
`)
	th.WriteF("resources.yaml", `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: test
spec:
  schedule: "5 10 * * 1"
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      backoffLimit: 3
      template:
        spec:
          restartPolicy: Never
          containers:
          - name: test
            image: bash
            command:
            - /bin/sh
            - -c
            - echo "test"
            env:
            - name: ENV1
              value: val1
            - name: ENV2
              value: val2
            - name: ENV3
              value: val3`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: test
spec:
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      backoffLimit: 3
      template:
        spec:
          containers:
          - command:
            - /bin/sh
            - -c
            - echo "test"
            env:
            - name: ENV_NEW
              value: val_new
            - name: ENV1
              value: val1
            - name: ENV2
              value: val2
            - name: ENV3
              value: val3
            image: bash
            name: test
          restartPolicy: Never
  schedule: 5 10 * * 1
`)
}
