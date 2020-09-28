// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSimpleBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: team-foo-
commonLabels:
  app: mynginx
  org: example.com
  team: foo
commonAnnotations:
  note: This is a test annotation
resources:
  - service.yaml
  - deployment.yaml
  - networkpolicy.yaml
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
	th.WriteF("/app/base/networkpolicy.yaml", `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: nginx
spec:
  podSelector:
    matchExpressions:
      - {key: app, operator: In, values: [test]}
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: nginx
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
`)
	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    org: example.com
    team: foo
  name: team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
    org: example.com
    team: foo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    org: example.com
    team: foo
  name: team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      org: example.com
      team: foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        org: example.com
        team: foo
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    org: example.com
    team: foo
  name: team-foo-nginx
spec:
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: mynginx
          org: example.com
          team: foo
  podSelector:
    matchExpressions:
    - key: app
      operator: In
      values:
      - test
`)
}

func makeBaseWithGenerators(th kusttest_test.Harness) {
	th.WriteK("/app", `
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
secretGenerator:
- name: secret-in-base
  literals:
  - username=admin
  - password=somepw
`)
	th.WriteF("/app/deployment.yaml", `
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
	th.WriteF("/app/service.yaml", `
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
}

func TestBaseWithGeneratorsAlone(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeBaseWithGenerators(th)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    org: example.com
    team: foo
  name: team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      org: example.com
      team: foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        org: example.com
        team: foo
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
          name: team-foo-configmap-in-base-798k5k7g9f
        name: configmap-in-base
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    org: example.com
    team: foo
  name: team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
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
    org: example.com
    team: foo
  name: team-foo-configmap-in-base-798k5k7g9f
---
apiVersion: v1
data:
  password: c29tZXB3
  username: YWRtaW4=
kind: Secret
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    org: example.com
    team: foo
  name: team-foo-secret-in-base-bgd6bkgdm2
type: Opaque
`)
}

func TestMergeAndReplaceGenerators(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeBaseWithGenerators(th)
	th.WriteF("/overlay/deployment.yaml", `
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
      - configMap:
          name: configmap-in-overlay
        name: configmap-in-overlay
`)
	th.WriteK("/overlay", `
namePrefix: staging-
commonLabels:
  env: staging
  team: override-foo
patchesStrategicMerge:
- deployment.yaml
resources:
- ../app
configMapGenerator:
- name: configmap-in-overlay
  literals:
  - hello=world
- name: configmap-in-base
  behavior: replace
  literals:
  - foo=override-bar
secretGenerator:
- name: secret-in-base
  behavior: merge
  literals:
  - proxy=haproxy
`)
	m := th.Run("/overlay", th.MakeDefaultOptions())
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
    team: override-foo
  name: staging-team-foo-nginx
spec:
  selector:
    matchLabels:
      app: mynginx
      env: staging
      org: example.com
      team: override-foo
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: mynginx
        env: staging
        org: example.com
        team: override-foo
    spec:
      containers:
      - image: nginx
        name: nginx
        volumeMounts:
        - mountPath: /tmp/ps
          name: nginx-persistent-storage
      volumes:
      - gcePersistentDisk:
          pdName: nginx-persistent-storage
        name: nginx-persistent-storage
      - configMap:
          name: staging-configmap-in-overlay-dc6fm46dhm
        name: configmap-in-overlay
      - configMap:
          name: staging-team-foo-configmap-in-base-hc6g9dk6g9
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
    team: override-foo
  name: staging-team-foo-nginx
spec:
  ports:
  - port: 80
  selector:
    app: mynginx
    env: staging
    org: example.com
    team: override-foo
---
apiVersion: v1
data:
  foo: override-bar
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: override-foo
  name: staging-team-foo-configmap-in-base-hc6g9dk6g9
---
apiVersion: v1
data:
  password: c29tZXB3
  proxy: aGFwcm94eQ==
  username: YWRtaW4=
kind: Secret
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mynginx
    env: staging
    org: example.com
    team: override-foo
  name: staging-team-foo-secret-in-base-k2k4692t9g
type: Opaque
---
apiVersion: v1
data:
  hello: world
kind: ConfigMap
metadata:
  labels:
    env: staging
    team: override-foo
  name: staging-configmap-in-overlay-dc6fm46dhm
`)
}

func TestGeneratingIntoNamespaces(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
configMapGenerator:
- name: test
  namespace: default
  literals:
    - key=value
- name: test
  namespace: kube-system
  literals:
    - key=value
secretGenerator:
- name: test
  namespace: default
  literals:
  - username=admin
  - password=somepw
- name: test
  namespace: kube-system
  literals:
  - username=admin
  - password=somepw
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  name: test-t757gk2bmf
  namespace: default
---
apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  name: test-t757gk2bmf
  namespace: kube-system
---
apiVersion: v1
data:
  password: c29tZXB3
  username: YWRtaW4=
kind: Secret
metadata:
  name: test-bgd6bkgdm2
  namespace: default
type: Opaque
---
apiVersion: v1
data:
  password: c29tZXB3
  username: YWRtaW4=
kind: Secret
metadata:
  name: test-bgd6bkgdm2
  namespace: kube-system
type: Opaque
`)
}

// Valid that conflict is detected is the name are identical
// and namespace left to default
func TestConfigMapGeneratingIntoSameNamespace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
configMapGenerator:
- name: test
  namespace: default
  literals:
  - key=value
- name: test
  literals:
  - key=value
`)
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "behavior must be merge or replace") {
		t.Fatalf("unexpected error %v", err)
	}
}

// Valid that conflict is detected is the name are identical
// and namespace left to default
func TestSecretGeneratingIntoSameNamespace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
secretGenerator:
- name: test
  namespace: default
  literals:
  - username=admin
  - password=somepw
- name: test
  literals:
  - username=admin
  - password=somepw
`)
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "behavior must be merge or replace") {
		t.Fatalf("unexpected error %v", err)
	}
}
