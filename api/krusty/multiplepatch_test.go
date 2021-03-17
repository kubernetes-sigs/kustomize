// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestRemoveEmptyDirWithNullFieldInSmp(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- deployment.yaml
patchesStrategicMerge:
- patch.yaml
`)
	th.WriteF("deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: fancyDisk
        emptyDir: {}
`)
	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: fancyDisk
        emptyDir: null
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: fancyDisk
`)
}

func TestRemoveEmptyDirAddPersistentDisk(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- deployment.yaml
patchesStrategicMerge:
- patch.yaml
`)
	th.WriteF("deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: fancyDisk
        emptyDir: {}
`)
	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: fancyDisk
        emptyDir: null
        gcePersistentDisk:
          pdName: fancyDisk
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - gcePersistentDisk:
          pdName: fancyDisk
        name: fancyDisk
`)
}

func TestVolumeRemoveEmptyDirInOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- deployment.yaml
configMapGenerator:
- name: baseCm
  literals:
  - foo=bar
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
        volumeMounts:
        - name: fancyDisk
          mountPath: /tmp/ps
      volumes:
      - name: fancyDisk
        emptyDir: {}
      - configMap:
          name: baseCm
        name: baseCm
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: fancyDisk
      volumes:
      - emptyDir: {}
        name: fancyDisk
      - configMap:
          name: baseCm-798k5k7g9f
        name: baseCm
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: baseCm-798k5k7g9f
`)

	th.WriteK("overlay", `
patchesStrategicMerge:
- patch.yaml
resources:
- ../base
configMapGenerator:
- name: overlayCm
  literals:
  - hello=world
`)
	th.WriteF("overlay/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: fancyDisk
        emptyDir: null
        gcePersistentDisk:
          pdName: fancyDisk
      - configMap:
          name: overlayCm
        name: overlayCm
`)
	m = th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: fancyDisk
      volumes:
      - gcePersistentDisk:
          pdName: fancyDisk
        name: fancyDisk
      - configMap:
          name: overlayCm-dc6fm46dhm
        name: overlayCm
      - configMap:
          name: baseCm-798k5k7g9f
        name: baseCm
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: baseCm-798k5k7g9f
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  name: overlayCm-dc6fm46dhm
`)
}

// Goal is to remove "  emptyDir: {}" with a patch.
func TestRemoveEmptyDirWithPatchesAtSameLevel(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
      - name: sidecar
        image: sidecar:latest
      volumes:
      - name: nginx-persistent-storage
        emptyDir: {}
`)
	th.WriteK("overlay", `
patchesStrategicMerge:
- deployment-patch1.yaml
- deployment-patch2.yaml
resources:
- ../base
`)
	th.WriteF("overlay/deployment-patch1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      volumes:
      - name: nginx-persistent-storage
        emptyDir: null
        gcePersistentDisk:
          pdName: nginx-persistent-storage
`)
	th.WriteF("overlay/deployment-patch2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        env:
        - name: ANOTHERENV
          value: FOO
      volumes:
      - name: nginx-persistent-storage
`)
	opts := th.MakeDefaultOptions()
	m := th.Run("overlay", opts)
	th.AssertActualEqualsExpected(
		m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - env:
        - name: ANOTHERENV
          value: FOO
        image: nginx
        name: nginx
      - image: sidecar:latest
        name: sidecar
      volumes:
      - gcePersistentDisk:
          pdName: nginx-persistent-storage
        name: nginx-persistent-storage
`)
}

func TestSimpleMultiplePatches(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: b-
commonLabels:
  team: foo
resources:
- deployment.yaml
- service.yaml
configMapGenerator:
- name: configmap-in-base
  literals:
  - foo=bar
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
        volumeMounts:
        - name: nginx-persistent-storage
          mountPath: /tmp/ps
      - name: sidecar
        image: sidecar:latest
      volumes:
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
	th.WriteF("base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  ports:
  - port: 80
`)
	th.WriteK("overlay", `
namePrefix: a-
commonLabels:
  env: staging
patchesStrategicMerge:
- deployment-patch1.yaml
- deployment-patch2.yaml
resources:
- ../base
configMapGenerator:
- name: configmap-in-overlay
  literals:
  - hello=world
`)
	th.WriteF("overlay/deployment-patch1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        env:
        - name: ENVKEY
          value: ENVVALUE
      volumes:
      - name: nginx-persistent-storage
        gcePersistentDisk:
          pdName: nginx-persistent-storage
      - configMap:
          name: configmap-in-overlay
        name: configmap-in-overlay
`)
	th.WriteF("overlay/deployment-patch2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ANOTHERENV
          value: FOO
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    env: staging
    team: foo
  name: a-b-nginx
spec:
  selector:
    matchLabels:
      env: staging
      team: foo
  template:
    metadata:
      labels:
        env: staging
        team: foo
    spec:
      containers:
      - env:
        - name: ANOTHERENV
          value: FOO
        - name: ENVKEY
          value: ENVVALUE
        image: nginx:latest
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      - image: sidecar:latest
        name: sidecar
      volumes:
      - gcePersistentDisk:
          pdName: nginx-persistent-storage
        name: nginx-persistent-storage
      - configMap:
          name: a-configmap-in-overlay-dc6fm46dhm
        name: configmap-in-overlay
      - configMap:
          name: a-b-configmap-in-base-798k5k7g9f
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  labels:
    env: staging
    team: foo
  name: a-b-nginx
spec:
  ports:
  - port: 80
  selector:
    env: staging
    team: foo
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    env: staging
    team: foo
  name: a-b-configmap-in-base-798k5k7g9f
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  labels:
    env: staging
  name: a-configmap-in-overlay-dc6fm46dhm
`)
}

func makeCommonFilesForMultiplePatchTests(th kusttest_test.Harness) {
	th.WriteK("base", `
namePrefix: team-foo-
commonLabels:
  app: mynginx
  org: example.com
  team: foo
commonAnnotations:
  note: This is a test annotation
resources:
  - deployment.yaml
  - service.yaml
configMapGenerator:
- name: configmap-in-base
  literals:
  - foo=bar
`)
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
      - name: sidecar
        image: sidecar:latest
      volumes:
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
`)
	th.WriteK("overlay/staging", `
namePrefix: staging-
commonLabels:
  env: staging
patchesStrategicMerge:
  - deployment-patch1.yaml
  - deployment-patch2.yaml
resources:
  - ../../base
configMapGenerator:
- name: configmap-in-overlay
  literals:
  - hello=world
`)
}

func TestMultiplePatchesNoConflict(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFilesForMultiplePatchTests(th)
	th.WriteF("overlay/staging/deployment-patch1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        env:
        - name: ENVKEY
          value: ENVVALUE
      volumes:
      - name: nginx-persistent-storage
        gcePersistentDisk:
          pdName: nginx-persistent-storage
      - configMap:
          name: configmap-in-overlay
        name: configmap-in-overlay
`)
	th.WriteF("overlay/staging/deployment-patch2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ANOTHERENV
          value: FOO
`)
	m := th.Run("overlay/staging", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      env: staging
      org: example.com
      team: foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        env: staging
        org: example.com
        team: foo
    spec:
      containers:
      - env:
        - name: ANOTHERENV
          value: FOO
        - name: ENVKEY
          value: ENVVALUE
        image: nginx:latest
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      - image: sidecar:latest
        name: sidecar
      volumes:
      - gcePersistentDisk:
          pdName: nginx-persistent-storage
        name: nginx-persistent-storage
      - configMap:
          name: staging-configmap-in-overlay-dc6fm46dhm
        name: configmap-in-overlay
      - configMap:
          name: staging-team-foo-configmap-in-base-798k5k7g9f
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
    env: staging
    org: example.com
    team: foo
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-configmap-in-base-798k5k7g9f
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  labels:
    env: staging
  name: staging-configmap-in-overlay-dc6fm46dhm
`)
}

func TestNonCommutablePatches(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFilesForMultiplePatchTests(th)
	th.WriteF("overlay/staging/deployment-patch1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ENABLE_FEATURE_FOO
          value: TRUE
      volumes:
      - name: nginx-persistent-storage
        gcePersistentDisk:
          pdName: nginx-persistent-storage
      - configMap:
          name: configmap-in-overlay
        name: configmap-in-overlay
`)
	th.WriteF("overlay/staging/deployment-patch2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ENABLE_FEATURE_FOO
          value: FALSE
`)
	// kyaml doesn't try to detect conflicts in patches
	// (so ENABLE_FEATURE_FOO FALSE wins).
	m := th.Run("overlay/staging", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      env: staging
      org: example.com
      team: foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        env: staging
        org: example.com
        team: foo
    spec:
      containers:
      - env:
        - name: ENABLE_FEATURE_FOO
          value: false
        image: nginx
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      - image: sidecar:latest
        name: sidecar
      volumes:
      - gcePersistentDisk:
          pdName: nginx-persistent-storage
        name: nginx-persistent-storage
      - configMap:
          name: staging-configmap-in-overlay-dc6fm46dhm
        name: configmap-in-overlay
      - configMap:
          name: staging-team-foo-configmap-in-base-798k5k7g9f
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
    env: staging
    org: example.com
    team: foo
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-configmap-in-base-798k5k7g9f
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  labels:
    env: staging
  name: staging-configmap-in-overlay-dc6fm46dhm
`)
}

func TestMultiplePatchesWithOnePatchDeleteDirective(t *testing.T) {
	additivePatch := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: SOME_NAME
          value: somevalue
`
	deletePatch := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - $patch: delete
        name: sidecar
`
	cases := map[string]struct {
		patch1      string
		patch2      string
		expectError bool
	}{
		"Patch with delete directive first": {
			patch1: deletePatch,
			patch2: additivePatch,
		},
		"Patch with delete directive second": {
			patch1: additivePatch,
			patch2: deletePatch,
		},
	}
	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			th := kusttest_test.MakeHarness(t)
			makeCommonFilesForMultiplePatchTests(th)
			th.WriteF("overlay/staging/deployment-patch1.yaml", c.patch1)
			th.WriteF("overlay/staging/deployment-patch2.yaml", c.patch2)
			m := th.Run("overlay/staging", th.MakeDefaultOptions())
			th.AssertActualEqualsExpected(m, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      env: staging
      org: example.com
      team: foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        env: staging
        org: example.com
        team: foo
    spec:
      containers:
      - env:
        - name: SOME_NAME
          value: somevalue
        image: nginx
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      volumes:
      - configMap:
          name: staging-team-foo-configmap-in-base-798k5k7g9f
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
    env: staging
    org: example.com
    team: foo
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-configmap-in-base-798k5k7g9f
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  labels:
    env: staging
  name: staging-configmap-in-overlay-dc6fm46dhm
`)
		})
	}
}

func TestMultiplePatchesBothWithPatchDeleteDirective(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFilesForMultiplePatchTests(th)
	th.WriteF("overlay/staging/deployment-patch1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - $patch: delete
        name: sidecar
`)
	th.WriteF("overlay/staging/deployment-patch2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - $patch: delete
        name: nginx
`)
	// kyaml doesn't fail on conflicts in patches; both containers
	// (nginx and sidecar) are deleted per this patching instruction.
	m := th.Run("overlay/staging", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      env: staging
      org: example.com
      team: foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        env: staging
        org: example.com
        team: foo
    spec:
      containers: []
      volumes:
      - configMap:
          name: staging-team-foo-configmap-in-base-798k5k7g9f
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
    env: staging
    org: example.com
    team: foo
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: foo
  name: staging-team-foo-configmap-in-base-798k5k7g9f
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  labels:
    env: staging
  name: staging-configmap-in-overlay-dc6fm46dhm
`)
}

// test for #3513
func TestSmpWithDifferentKeysOnDifferentPorts(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
  - resource.yaml
patches:
  - path: patch.yaml
    target:
      kind: StatefulSet
      name: myapp
`)
	th.WriteF("resource.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
        - name: consul
          image: "dashicorp/consul:1.9.1"
          ports:
            - containerPort: 8500
              name: http
            - containerPort: 8501
              name: https
            - containerPort: 8301
              protocol: "TCP"
              name: serflan-tcp
            - containerPort: 8301
              protocol: "UDP"
              name: serflan-udp
            - containerPort: 8302
              name: serfwan
            - containerPort: 8300
              name: server
            - containerPort: 8600
              name: dns-tcp
              protocol: "TCP"
            - containerPort: 8600
              name: dns-udp
              protocol: "UDP"`)
	th.WriteF("patch.yaml", `
kind: StatefulSet
metadata:
  name: myapp
  labels:
    test: label
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    test: label
  name: myapp
spec:
  template:
    spec:
      containers:
      - image: dashicorp/consul:1.9.1
        name: consul
        ports:
        - containerPort: 8301
          name: serflan-tcp
          protocol: TCP
        - containerPort: 8301
          name: serflan-udp
          protocol: UDP
        - containerPort: 8600
          name: dns-tcp
          protocol: TCP
        - containerPort: 8600
          name: dns-udp
          protocol: UDP
        - containerPort: 8500
          name: http
        - containerPort: 8501
          name: https
        - containerPort: 8302
          name: serfwan
        - containerPort: 8300
          name: server
`)
}

// test for #3616
func TestSmpDeleteOnResource(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- workloads.yaml
patches:
- patch: |
    apiVersion: monitoring.coreos.com/v1
    kind: PrometheusRule
    metadata:
      name: rule1
    $patch: delete
`)
	th.WriteF("workloads.yaml", `
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    role: alert-rules
  name: rule1
spec:
  groups:
  - name: rabbitmq.rules
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    role: alert-rules
  name: rule2
spec:
  groups:
  - name: rabbitmq.rules
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    role: alert-rules
  name: rule2
spec:
  groups:
  - name: rabbitmq.rules
`)
}
