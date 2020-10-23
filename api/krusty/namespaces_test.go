// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamespacedSecrets(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
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
	// This validates Fix #1444. This should not be an error anymore -
	// the secrets have the same name but are in different namespaces.
	// The ClusterRole (by def) is not in a namespace,
	// an in this case applies to *any* Secret resource
	// named "dummy"
	m := th.Run("/app", th.MakeDefaultOptions())
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

// TestNameAndNsTransformation validates that NamespaceTransformer,
// PrefixSuffixTransformer and namereference transformers are
// able to deal with simultaneous change of namespace and name.
func TestNameAndNsTransformation(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteK("/nameandns", `
namePrefix: p1-
nameSuffix: -s1
namespace: newnamespace
resources:
- resources.yaml
`)

	th.WriteF("/nameandns/resources.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
  namespace: ns1
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: ns1
---
apiVersion: v1
kind: Service
metadata:
  name: svc2
  namespace: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa1
  namespace: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa2
  namespace: ns1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
subjects:
- kind: ServiceAccount
  name: sa1
  namespace: ns1
- kind: ServiceAccount
  name: sa2
  namespace: ns1
- kind: ServiceAccount
  name: sa3
  namespace: random
- kind: ServiceAccount
  name: default
  namespace: irrelevant
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: example
webhooks:
  - name: example1
    clientConfig:
      service:
        name: svc1
        namespace: ns1
  - name: example2
    clientConfig:
      service:
        name: svc2
        namespace: ns1
  - name: example3
    clientConfig:
      service:
        name: svc3
        namespace: random
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crds.my.org
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cr1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb1
subjects:
- kind: ServiceAccount
  name: default
  namespace: irrelevant
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv1
`)
	m := th.Run("/nameandns", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: p1-cm1-s1
  namespace: newnamespace
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: p1-cm2-s1
  namespace: newnamespace
---
apiVersion: v1
kind: Service
metadata:
  name: p1-svc1-s1
  namespace: newnamespace
---
apiVersion: v1
kind: Service
metadata:
  name: p1-svc2-s1
  namespace: newnamespace
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: p1-sa1-s1
  namespace: newnamespace
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: p1-sa2-s1
  namespace: newnamespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: p1-manager-rolebinding-s1
subjects:
- kind: ServiceAccount
  name: p1-sa1-s1
  namespace: newnamespace
- kind: ServiceAccount
  name: p1-sa2-s1
  namespace: newnamespace
- kind: ServiceAccount
  name: sa3
  namespace: random
- kind: ServiceAccount
  name: default
  namespace: newnamespace
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: p1-example-s1
webhooks:
- clientConfig:
    service:
      name: p1-svc1-s1
      namespace: newnamespace
  name: example1
- clientConfig:
    service:
      name: p1-svc2-s1
      namespace: newnamespace
  name: example2
- clientConfig:
    service:
      name: svc3
      namespace: random
  name: example3
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crds.my.org
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: p1-cr1-s1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: p1-crb1-s1
subjects:
- kind: ServiceAccount
  name: default
  namespace: newnamespace
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: p1-pv1-s1
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
- elasticsearch-dev-service.yaml
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
// using the same name in different namespaces are treated as ambiguous if the namespace is
// not specified
func TestVariablesAmbiguous(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/namespaceNeedInVar/myapp", namespaceNeedInVarMyApp)
	th.WriteF("/namespaceNeedInVar/myapp/elasticsearch-dev-service.yaml", namespaceNeedInVarDevResources)
	th.WriteF("/namespaceNeedInVar/myapp/elasticsearch-test-service.yaml", namespaceNeedInVarTestResources)
	err := th.RunWithErr("/namespaceNeedInVar/myapp", th.MakeDefaultOptions())
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
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/namespaceNeedInVar/dev", namespaceNeedInVarDevFolder)
	th.WriteF("/namespaceNeedInVar/dev/elasticsearch-dev-service.yaml", namespaceNeedInVarDevResources)
	th.WriteK("/namespaceNeedInVar/test", namespaceNeedInVarTestFolder)
	th.WriteF("/namespaceNeedInVar/test/elasticsearch-test-service.yaml", namespaceNeedInVarTestResources)
	th.WriteK("/namespaceNeedInVar/workaround", `
resources:
- ../dev
- ../test
`)
	m := th.Run("/namespaceNeedInVar/workaround", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, namespaceNeedInVarExpectedOutput)
}

const namespaceNeedInVarMyAppWithNamespace string = `
resources:
- elasticsearch-dev-service.yaml
- elasticsearch-test-service.yaml
vars:
- name: elasticsearch-test-service-name
  objref:
    kind: Service
    name: elasticsearch
    namespace: test
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-test-protocol
  objref:
    kind: Service
    name: elasticsearch
    namespace: test
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
- name: elasticsearch-dev-service-name
  objref:
    kind: Service
    name: elasticsearch
    namespace: dev
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-dev-protocol
  objref:
    kind: Service
    name: elasticsearch
    namespace: dev
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`

// TestVariablesDisambiguatedWithNamespace demonstrates that adding the namespace
// to the variable declarations allows to disambiguate the variables.
func TestVariablesDisambiguatedWithNamespace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/namespaceNeedInVar/myapp", namespaceNeedInVarMyAppWithNamespace)
	th.WriteF("/namespaceNeedInVar/myapp/elasticsearch-dev-service.yaml", namespaceNeedInVarDevResources)
	th.WriteF("/namespaceNeedInVar/myapp/elasticsearch-test-service.yaml", namespaceNeedInVarTestResources)
	m := th.Run("/namespaceNeedInVar/myapp", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, namespaceNeedInVarExpectedOutput)
}
