// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

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
      - name: nginx-persistent-storage
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
        emptyDir: null
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
      volumes:
      - name: nginx-persistent-storage
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

func makeCommonFileForMultiplePatchTest(th kusttest_test.Harness) {
	th.WriteK("/app/base", `
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
      - name: sidecar
        image: sidecar:latest
      volumes:
      - name: nginx-persistent-storage
        emptyDir: {}
      - configMap:
          name: configmap-in-base
        name: configmap-in-base
`)
	th.WriteF("/app/base/service.yaml", `
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
	th.WriteK("/app/overlay/staging", `
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
	makeCommonFileForMultiplePatchTest(th)
	th.WriteF("/app/overlay/staging/deployment-patch1.yaml", `
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
        emptyDir: null
        gcePersistentDisk:
          pdName: nginx-persistent-storage
      - configMap:
          name: configmap-in-overlay
        name: configmap-in-overlay
`)
	th.WriteF("/app/overlay/staging/deployment-patch2.yaml", `
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
      volumes:
      - name: nginx-persistent-storage
`)
	m := th.Run("/app/overlay/staging", th.MakeDefaultOptions())
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

func TestMultiplePatchesWithConflict(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForMultiplePatchTest(th)
	th.WriteF("/app/overlay/staging/deployment-patch1.yaml", `
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
        emptyDir: null
        gcePersistentDisk:
          pdName: nginx-persistent-storage
      - configMap:
          name: configmap-in-overlay
        name: configmap-in-overlay
`)
	th.WriteF("/app/overlay/staging/deployment-patch2.yaml", `
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
	err := th.RunWithErr("/app/overlay/staging", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected conflict")
	}
	if !strings.Contains(
		err.Error(), "conflict between ") {
		t.Fatalf("Unexpected err: %v", err)
	}
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
	cases := []struct {
		name        string
		patch1      string
		patch2      string
		expectError bool
	}{
		{
			name:   "Patch with delete directive first",
			patch1: deletePatch,
			patch2: additivePatch,
		},
		{
			name:   "Patch with delete directive second",
			patch1: additivePatch,
			patch2: deletePatch,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			th := kusttest_test.MakeHarness(t)

			makeCommonFileForMultiplePatchTest(th)
			th.WriteF("/app/overlay/staging/deployment-patch1.yaml", c.patch1)
			th.WriteF("/app/overlay/staging/deployment-patch2.yaml", c.patch2)
			m := th.Run("/app/overlay/staging", th.MakeDefaultOptions())
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
      - emptyDir: {}
        name: nginx-persistent-storage
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
	makeCommonFileForMultiplePatchTest(th)
	th.WriteF("/app/overlay/staging/deployment-patch1.yaml", `
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
	th.WriteF("/app/overlay/staging/deployment-patch2.yaml", `
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
	err := th.RunWithErr("/app/overlay/staging", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("Expected error")
	}
	if !strings.Contains(
		err.Error(), "both containing ") {
		t.Fatalf("Unexpected err: %v", err)
	}
}
