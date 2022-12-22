// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func makeCommonFileForExtendedPatchTest(th kusttest_test.Harness) {
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  labels:
    app: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: busybox
        image: busybox
        volumeMounts:
        - name: busybox-persistent-storage
          mountPath: /tmp/ps
      volumes:
      - name: busybox-persistent-storage
        emptyDir: {}
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
	th.WriteF("base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
    - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  name: busybox
  labels:
    app: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchNameSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    name: busybox
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchGvkSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchLabelSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    labelSelector: app=nginx
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  annotations:
    new-key: new-value
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    new-key: new-value
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchNameGvkSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    name: busybox
    kind: Deployment
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchNameLabelSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    name: .*
    labelSelector: app=busybox
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchGvkLabelSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
    labelSelector: app=busybox
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchNameGvkLabelSelector(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    name: busybox
    kind: Deployment
    labelSelector: app=busybox
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchNoMatch(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    name: no-match
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
`)
	err := th.RunWithErr("base", th.MakeDefaultOptions())
	assert.Contains(t, err.Error(), "patches target not found for [noKind].[noVer].[noGrp]/no-match.[noNs]")
}

func TestExtendedPatchWithoutTarget(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestExtendedPatchNoMatchMultiplePatch(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch.yaml
  target:
    name: no-match
- path: patch.yaml
  target:
    name: busybox
    kind: Job
`)
	th.WriteF("base/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key: new-value
`)
	err := th.RunWithErr("base", th.MakeDefaultOptions())
	assert.Contains(t, err.Error(), "patches target not found for [noKind].[noVer].[noGrp]/no-match.[noNs]")
}

func TestExtendedPatchMultiplePatchOverlapping(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- deployment.yaml
- service.yaml
patches:
- path: patch1.yaml
  target:
    labelSelector: app=busybox
- path: patch2.yaml
  target:
    name: busybox
    kind: Deployment
`)
	th.WriteF("base/patch1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key-from-patch1: new-value
`)
	th.WriteF("base/patch2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
  annotations:
    new-key-from-patch2: new-value
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
      - image: nginx
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    new-key-from-patch1: new-value
    new-key-from-patch2: new-value
  labels:
    app: busybox
  name: busybox
spec:
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/ps
          name: busybox-persistent-storage
      volumes:
      - emptyDir: {}
        name: busybox-persistent-storage
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    new-key-from-patch1: new-value
  labels:
    app: busybox
  name: busybox
spec:
  ports:
  - port: 8080
  selector:
    app: busybox
`)
}

func TestTargetMissingPatchJson6902Error(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForExtendedPatchTest(th)
	th.WriteK("base", `
resources:
- servicemonitor.yaml
patchesJson6902:
- target:
    group: monitoring.coreos.com
    kind: ServiceMonitor
    name: starboard-exporter
    namespace: starboard
    version: v2
  path: patch.0.yaml
`)
	th.WriteF("base/servicemonitor.yaml", `
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app: starboard-exporter
  name: starboard-exporter
  namespace: starboard
spec:
  endpoints:
  - path: /metrics
    port: metrics
  selector:
    matchLabels:
      app.kubernetes.io/instance: starboard-exporter
      app.kubernetes.io/name: starboard-exporter
`)
	th.WriteF("base/patch.0.yaml", `
- op: add
  path: /metadata/labels/release
  value: kube-prometheus-stack
`)
	err := th.RunWithErr("base", th.MakeDefaultOptions())
	assert.Contains(t, err.Error(), "patchesJson6902 target not found for ServiceMonitor.v2.monitoring.coreos.com/starboard-exporter.starboard")
}
