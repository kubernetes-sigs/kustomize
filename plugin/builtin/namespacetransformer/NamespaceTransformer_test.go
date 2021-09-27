// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamespaceTransformer1(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("NamespaceTransformer")
	defer th.Reset()
	th.RunTransformerAndCheckResult(`
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
  group: rbac.authorization.k8s.io
- path: subjects
  kind: ClusterRoleBinding
  group: rbac.authorization.k8s.io
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
apiVersion: admissionregistration.k8s.io/v1
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd
`,
		// Import note: The namespace transformer is in charge of
		// the metadata.namespace field. The namespace transformer SHOULD
		// NOT modify neither the "namespace" subfield within the
		// ClusterRoleBinding.subjects field nor the "namespace"
		// subfield in the ValidatingWebhookConfiguration.webhooks field.
		// This is the role of the namereference Transformer to handle
		// object reference changes (prefix/suffix and namespace).
		// For use cases involving simultaneous change of name and namespace,
		// refer to namespaces tests in pkg/target test suites.
		`
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
apiVersion: admissionregistration.k8s.io/v1
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd
`)
}

func TestNamespaceTransformerClusterLevelKinds(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("NamespaceTransformer")
	defer th.Reset()

	const noChangeExpected = `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd1
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
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv1
`

	th.RunTransformerAndCheckResult(`
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
  group: rbac.authorization.k8s.io
- path: subjects
  kind: ClusterRoleBinding
  group: rbac.authorization.k8s.io
`, noChangeExpected, noChangeExpected)
}

func TestNamespaceTransformerObjectConflict(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("NamespaceTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
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
`,
		func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "ID conflict") {
				t.Fatalf("unexpected error: %s", err.Error())
			}
		})
}
