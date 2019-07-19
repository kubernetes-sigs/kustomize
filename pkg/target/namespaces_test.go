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

// This serie of constants is used to prove the need of
// the namespace field in the objref field of the var declaration.
// The following tests demonstrate that it creates umbiguous variable
// declaration if two entities of the kind with the same name
// but in different namespaces are declared.
// This is tracking the following issue:
// https://github.com/kubernetes-sigs/kustomize/issues/1298
const namespaceNeedInVarMyApp string = `
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

const namespaceNeedInVarDevResources string = `
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

const namespaceNeedInVarTestResources string = `
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

const namespaceNeedInVarExpectedOutput string = `
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

// TestVariablesAmbiguous demonstrates how two variables pointing at two different resources
// using the same name in different namespaces are treated as ambiguous
// The fundamental reason is that objRef field in the variable does not contain a namespace
// qualifier.
// Once the namespace qualifer is added, as for the other resources lookup in the coder,
// the ResId.GvknEquals method usage will have to be phased out and replaced by ResId.Equals.
func TestVariablesAmbiguous(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/namespaceNeedInVar/myapp")
	th.WriteK("/namespaceNeedInVar/myapp", namespaceNeedInVarMyApp)
	th.WriteF("/namespaceNeedInVar/myapp/elasticsearch-dev-service.yaml", namespaceNeedInVarDevResources)
	th.WriteF("/namespaceNeedInVar/myapp/elasticsearch-test-service.yaml", namespaceNeedInVarTestResources)
	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "unable to disambiguate") {
		t.Fatalf("unexpected error %v", err)
	}
}

const namespaceNeedInVarDevFolder string = `
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

const namespaceNeedInVarTestFolder string = `
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

// TestVariablesAmbiguousWorkaround demonstrates a possible workaround
// to TestVariablesAmbiguous problem. It requires to separate the variables
// and resources into multiple kustomization context/folders instead of one.
func TestVariablesAmbiguousWorkaround(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/namespaceNeedInVar/workaround")
	th.WriteK("/namespaceNeedInVar/dev", namespaceNeedInVarDevFolder)
	th.WriteF("/namespaceNeedInVar/dev/elasticsearch-dev-service.yaml", namespaceNeedInVarDevResources)
	th.WriteK("/namespaceNeedInVar/test", namespaceNeedInVarTestFolder)
	th.WriteF("/namespaceNeedInVar/test/elasticsearch-test-service.yaml", namespaceNeedInVarTestResources)
	th.WriteK("/namespaceNeedInVar/workaround", `
resources:
- ../dev
- ../test
`)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, namespaceNeedInVarExpectedOutput)
}
