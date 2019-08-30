// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"sigs.k8s.io/kustomize/v3/pkg/plugins/testenv"
)

func TestNamespaceTransformer1(t *testing.T) {
	tc := testenv.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "NamespaceTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: NamespaceTransformer
metadata:
  name: notImportantHere
  namespace: test
fieldSpecs:
# replace or add namespace field
# on all entities by default
- path: metadata/namespace
  create: true

# Update namespace if necessary
# in the subjects fields
- path: subjects
  kind: RoleBinding
- path: subjects
  kind: ClusterRoleBinding

# skip those ClusterWide entities
- path: metadata/namespace
  kind: APIService
  skip: true
- path: metadata/namespace
  kind: CSIDriver
  skip: true
- path: metadata/namespace
  kind: CSINode
  skip: true
- path: metadata/namespace
  kind: CertificateSigningRequest
  skip: true
- path: metadata/namespace
  kind: ClusterRole
  skip: true
- path: metadata/namespace
  kind: ClusterRoleBinding
  skip: true
- path: metadata/namespace
  kind: ComponentStatus
  skip: true
- path: metadata/namespace
  kind: CustomResourceDefinition
  skip: true
- path: metadata/namespace
  kind: MutatingWebhookConfiguration
  skip: true
- path: metadata/namespace
  kind: Namespace
  skip: true
- path: metadata/namespace
  kind: Node
  skip: true
- path: metadata/namespace
  kind: PersistentVolume
  skip: true
- path: metadata/namespace
  kind: PodSecurityPolicy
  skip: true
- path: metadata/namespace
  kind: PriorityClass
  skip: true
- path: metadata/namespace
  kind: RuntimeClass
  skip: true
- path: metadata/namespace
  kind: SelfSubjectAccessReview
  skip: true
- path: metadata/namespace
  kind: SelfSubjectRulesReview
  skip: true
- path: metadata/namespace
  kind: StorageClass
  skip: true
- path: metadata/namespace
  kind: SubjectAccessReview
  skip: true
- path: metadata/namespace
  kind: TokenReview
  skip: true
- path: metadata/namespace
  kind: ValidatingWebhookConfiguration
  skip: true
- path: metadata/namespace
  kind: VolumeAttachment
  skip: true
`, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
  namespace: foo
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
---
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: test
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
  namespace: system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
subjects:
- kind: ServiceAccount
  name: default
  namespace: system
- kind: ServiceAccount
  name: service-account
  namespace: system
- kind: ServiceAccount
  name: another
  namespace: random
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: example
webhooks:
  - name: example1
    clientConfig:
      service:
        name: svc1
        namespace: system
  - name: example2
    clientConfig:
      service:
        name: svc2
        namespace: system
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
`)

	// Import note: The namespace transformer is in charge of
	// the metadata.namespace field. The namespace transformer SHOULD
	// NOT modify neither the "namespace" subfield within the
	// ClusterRoleBinding.subjects field nor the "namespace"
	// subfield in the ValidatingWebhookConfiguration.webhooks field.
	// This is the role of the namereference Transformer to handle
	// object reference changes (prefix/suffix and namespace).
	// For use cases involving simultaneous change of name and namespace,
	// refer to namespaces tests in pkg/target test suites.
	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
  namespace: test
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
  namespace: test
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: test
---
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: test
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
subjects:
- kind: ServiceAccount
  name: default
  namespace: test
- kind: ServiceAccount
  name: service-account
  namespace: system
- kind: ServiceAccount
  name: another
  namespace: random
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: example
webhooks:
- clientConfig:
    service:
      name: svc1
      namespace: system
  name: example1
- clientConfig:
    service:
      name: svc2
      namespace: system
  name: example2
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
`)
}

func TestNamespaceTransformerClusterLevelKinds(t *testing.T) {
	tc := testenv.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "NamespaceTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	const noChangeExpected = `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
kind: CustomResourceDefinition
metadata:
  name: crd1
---
kind: ClusterRole
metadata:
  name: cr1
---
kind: ClusterRoleBinding
metadata:
  name: crb1
---
kind: PersistentVolume
metadata:
  name: pv1
`
	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: NamespaceTransformer
metadata:
  name: notImportantHere
  namespace: test
fieldSpecs:
  - path: metadata/namespace
    create: true
  - path: subjects
    kind: RoleBinding
  - path: subjects
    kind: ClusterRoleBinding
  - path: metadata/namespace
    kind: Namespace
    skip: true
  - path: metadata/namespace
    kind: ClusterRole
    skip: true
  - path: metadata/namespace
    kind: ClusterRoleBinding
    skip: true
  - path: metadata/namespace
    kind: ComponentStatus
    skip: true
  - path: metadata/namespace
    kind: CustomResourceDefinition
    skip: true
  - path: metadata/namespace
    kind: PersistentVolume
    skip: true
`, noChangeExpected)

	th.AssertActualEqualsExpected(rm, noChangeExpected)
}

func TestNamespaceTransformerObjectConflict(t *testing.T) {
	tc := testenv.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "NamespaceTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	err := th.ErrorFromLoadAndRunTransformer(`
apiVersion: builtin
kind: NamespaceTransformer
metadata:
  name: notImportantHere
  namespace: test
fieldSpecs:
- path: metadata/namespace
  create: true
`, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  namespace: foo
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  namespace: bar
`)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ID conflict") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
