// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Reproduce issue #3732
func TestValidatingWebhookCombinedNamespaces(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- service.yaml
- validatingwebhook.yaml
`)
	th.WriteF("base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: admission
  namespace: base-namespace
spec:
  type: ClusterIP
  ports:
    - name: https-webhook
      port: 443
      targetPort: webhook
`)
	th.WriteF("base/validatingwebhook.yaml", `
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: validatingwebhook
webhooks:
  - name: validate
    matchPolicy: Equivalent
    rules:
      - apiGroups:
          - networking.k8s.io
        apiVersions:
          - v1beta1
        operations:
          - CREATE
          - UPDATE
        resources:
          - ingresses
    failurePolicy: Fail
    sideEffects: None
    admissionReviewVersions:
      - v1
      - v1beta1
    clientConfig:
      service:
        namespace: base-namespace
        name: admission
        path: /networking/v1beta1/ingresses
`)
	th.WriteK("overlay", `
namespace: merge-namespace
resources:
- ../base
patchesStrategicMerge:
- validatingwebhookdelete.yaml
`)
	th.WriteF("overlay/validatingwebhookdelete.yaml", `
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: validatingwebhook
$patch: delete
`)
	th.WriteK("combined", `
resources:
- ../base
- ../overlay
`)
	m := th.Run("combined", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  name: admission
  namespace: base-namespace
spec:
  ports:
  - name: https-webhook
    port: 443
    targetPort: webhook
  type: ClusterIP
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: validatingwebhook
webhooks:
- admissionReviewVersions:
  - v1
  - v1beta1
  clientConfig:
    service:
      name: admission
      namespace: merge-namespace
      path: /networking/v1beta1/ingresses
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: validate
  rules:
  - apiGroups:
    - networking.k8s.io
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - ingresses
  sideEffects: None
---
apiVersion: v1
kind: Service
metadata:
  name: admission
  namespace: merge-namespace
spec:
  ports:
  - name: https-webhook
    port: 443
    targetPort: webhook
  type: ClusterIP
`)
}
