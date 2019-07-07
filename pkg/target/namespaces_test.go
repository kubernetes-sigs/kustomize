// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"strings"
	"testing"
)

func TestNamespacedSecrets(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")

	th.WriteF("/app/secrets.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: dummy
  namespace: default
type: Opaque
data:
  dummy: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: dummy
  namespace: kube-system
type: Opaque
data:
  dummy: ""
`)

	// This should find the proper secret.
	th.WriteF("/app/role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: dummy
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["dummy"]
  verbs: ["get"]
`)

	th.WriteK("/app", `
resources:
- secrets.yaml
- role.yaml
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	// This validates Fix #1444. This should not be an error anymore -
	// the secrets have the same name but are in different namespaces.
	// The ClusterRole (by def) is not in a namespace,
	// an in this case applies to *any* Secret resource
	// named "dummy"
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  dummy: ""
kind: Secret
metadata:
  name: dummy
  namespace: default
type: Opaque
---
apiVersion: v1
data:
  dummy: ""
kind: Secret
metadata:
  name: dummy
  namespace: kube-system
type: Opaque
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: dummy
rules:
- apiGroups:
  - ""
  resourceNames:
  - dummy
  resources:
  - secrets
  verbs:
  - get
`)
}

const recreate1298_dev string = `
resources:
- elasticsearch-dev-service.yaml
vars:
- name: elasticsearch-dev-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-dev-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`

const recreate1298_test string = `
resources:
- elasticsearch-test-service.yaml
vars:
- name: elasticsearch-test-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-test-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`

const recreate1298_onefolder string = `
resources:
- elasticsearch-test-service.yaml
- elasticsearch-dev-service.yaml
vars:
- name: elasticsearch-test-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-test-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
- name: elasticsearch-dev-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-dev-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`

const recreate1298_twofolders string = `
resources:
- ../dev
- ../test
`

const recreate1298_devresources string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: dev
spec:
  template:
    spec:
      containers:
        - name: elasticsearch
          env:
            - name: DISCOVERY_SERVICE
              value: "$(elasticsearch-dev-service-name).monitoring.svc.cluster.local"
            - name: DISCOVERY_PROTOCOL
              value: "$(elasticsearch-dev-protocol)"
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: dev
spec:
  ports:
    - name: transport
      port: 9300
      protocol: TCP
  clusterIP: None
`

const recreate1298_testresources string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: test
spec:
  template:
    spec:
      containers:
        - name: elasticsearch
          env:
            - name: DISCOVERY_SERVICE
              value: "$(elasticsearch-test-service-name).monitoring.svc.cluster.local"
            - name: DISCOVERY_PROTOCOL
              value: "$(elasticsearch-test-protocol)"
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: test
spec:
  ports:
    - name: transport
      port: 9300
      protocol: UDP
  clusterIP: None
`

const expectedOutput string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: dev
spec:
  template:
    spec:
      containers:
      - env:
        - name: DISCOVERY_SERVICE
          value: elasticsearch.monitoring.svc.cluster.local
        - name: DISCOVERY_PROTOCOL
          value: TCP
        name: elasticsearch
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: dev
spec:
  clusterIP: None
  ports:
  - name: transport
    port: 9300
    protocol: TCP
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: test
spec:
  template:
    spec:
      containers:
      - env:
        - name: DISCOVERY_SERVICE
          value: elasticsearch.monitoring.svc.cluster.local
        - name: DISCOVERY_PROTOCOL
          value: UDP
        name: elasticsearch
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: test
spec:
  clusterIP: None
  ports:
  - name: transport
    port: 9300
    protocol: UDP
`

func TestVariablesInTwoFolder(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/recreate1298/twofolders")
	th.WriteK("/recreate1298/twofolders", recreate1298_twofolders)
	th.WriteK("/recreate1298/dev", recreate1298_dev)
	th.WriteK("/recreate1298/test", recreate1298_test)
	th.WriteF("/recreate1298/dev/elasticsearch-dev-service.yaml", recreate1298_devresources)
	th.WriteF("/recreate1298/test/elasticsearch-test-service.yaml", recreate1298_testresources)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, expectedOutput)
}

func TestVariablesInOneFolder(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/recreate1298/onefolder")
	th.WriteK("/recreate1298/onefolder", recreate1298_onefolder)
	th.WriteF("/recreate1298/onefolder/elasticsearch-dev-service.yaml", recreate1298_devresources)
	th.WriteF("/recreate1298/onefolder/elasticsearch-test-service.yaml", recreate1298_testresources)
	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	// This requires the namespace to be added to the variable declaration.
	// The expected output would then be identical regardless if the
	// the files are in one or two folders.
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "unable to disambiguate") {
		t.Fatalf("unexpected error %v", err)
	}

}
