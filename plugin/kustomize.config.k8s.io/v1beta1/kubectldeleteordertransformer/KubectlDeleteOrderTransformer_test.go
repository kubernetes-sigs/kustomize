// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

func TestKubectlDeleteOrderTransformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"kustomize.config.k8s.io", "v1beta1", "KubectlDeleteOrderTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	rm := th.LoadAndRunTransformer(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: KubectlDeleteOrderTransformer
metadata:
  name: notImportantHere
`, `
apiVersion: v1
kind: Service
metadata:
  name: papaya
---
apiVersion: v1
kind: Role
metadata:
  name: banana
---
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
---
apiVersion: v1
kind: LimitRange
metadata:
  name: peach
---
apiVersion: v1
kind: Deployment
metadata:
  name: pear
---
apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Secret
metadata:
  name: quince
---
apiVersion: v1
kind: CRD1
metadata:
  name: durian
---
apiVersion: v1
kind: CRD2
metadata:
  name: durian
---
apiVersion: v1
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apricot
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
---
apiVersion: v1
kind: CRD2
metadata:
  name: durian
---
apiVersion: v1
kind: CRD1
metadata:
  name: durian
---
apiVersion: v1
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: Service
metadata:
  name: papaya
---
apiVersion: v1
kind: Deployment
metadata:
  name: pear
---
apiVersion: v1
kind: Role
metadata:
  name: banana
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apricot
---
apiVersion: v1
kind: Secret
metadata:
  name: quince
---
apiVersion: v1
kind: LimitRange
metadata:
  name: peach
---
apiVersion: v1
kind: Namespace
metadata:
  name: apple
`)
}

func TestKubectlDeleteOrderTransformerCustomOrder(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"kustomize.config.k8s.io", "v1beta1", "KubectlDeleteOrderTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	rm := th.LoadAndRunTransformer(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: KubectlDeleteOrderTransformer
metadata:
  name: notImportantHere
kindorder:
- ValidatingWebhookConfiguration
- Ingress
- Deployment
- Service
- Role
- ConfigMap
- Secret
- CRD2
- CRD1
- LimitRange
- Namespace
`, `
apiVersion: v1
kind: Service
metadata:
  name: papaya
---
apiVersion: v1
kind: Role
metadata:
  name: banana
---
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
---
apiVersion: v1
kind: LimitRange
metadata:
  name: peach
---
apiVersion: v1
kind: Deployment
metadata:
  name: pear
---
apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Secret
metadata:
  name: quince
---
apiVersion: v1
kind: CRD1
metadata:
  name: durian
---
apiVersion: v1
kind: CRD2
metadata:
  name: durian
---
apiVersion: v1
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apricot
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
---
apiVersion: v1
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: Deployment
metadata:
  name: pear
---
apiVersion: v1
kind: Service
metadata:
  name: papaya
---
apiVersion: v1
kind: Role
metadata:
  name: banana
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apricot
---
apiVersion: v1
kind: Secret
metadata:
  name: quince
---
apiVersion: v1
kind: CRD2
metadata:
  name: durian
---
apiVersion: v1
kind: CRD1
metadata:
  name: durian
---
apiVersion: v1
kind: LimitRange
metadata:
  name: peach
---
apiVersion: v1
kind: Namespace
metadata:
  name: apple
`)
}
