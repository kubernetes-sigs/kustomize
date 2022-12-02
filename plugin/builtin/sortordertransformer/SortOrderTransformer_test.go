// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestWrongOrder(t *testing.T) {
	resources := `
apiVersion: v1
kind: lalakis
metadata:
  name: lalakis
`
	transformer := `
  apiVersion: builtin
  kind: SortOrderTransformer
  metadata:
    name: notImportantHere
  sortOptions:
    order: invalid_value
`
	th := kusttest_test.MakeEnhancedHarness(t).PrepBuiltin("SortOrderTransformer")
	defer th.Reset()
	th.RunTransformerAndCheckError(
		transformer,
		resources,
		func(t *testing.T, err error) {
			t.Helper()
			require.EqualError(t, err, "the field 'sortOptions.order' must be one of [fifo, legacy]")
		},
	)
}

func TestSortOrderTransformer(t *testing.T) {
	resources := `
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
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apricot
`

	tests := []struct {
		name           string
		resources      string
		transformer    string
		expectedOutput string
	}{
		{
			name:      "default ordering",
			resources: resources,
			transformer: `
apiVersion: builtin
kind: SortOrderTransformer
metadata:
  name: notImportantHere
sortOptions:
  order: legacy
`,
			expectedOutput: `
apiVersion: v1
kind: Namespace
metadata:
  name: apple
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
kind: Service
metadata:
  name: papaya
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
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
`,
		},
		{
			name:      "webhooks first, order namespace and deployment last",
			resources: resources,
			transformer: `
apiVersion: builtin
kind: SortOrderTransformer
metadata:
  name: notImportantHere
sortOptions:
  order: legacy
  legacySortOptions:
    orderFirst:
    - MutatingWebhookConfiguration
    - ValidatingWebhookConfiguration
    - ResourceQuota
    - StorageClass
    - CustomResourceDefinition
    - ServiceAccount
    - PodSecurityPolicy
    - Role
    - ClusterRole
    - RoleBinding
    - ClusterRoleBinding
    - ConfigMap
    - Secret
    - Endpoints
    - Service
    - LimitRange
    - PriorityClass
    - PersistentVolume
    - PersistentVolumeClaim
    - StatefulSet
    - CronJob
    - PodDisruptionBudget
    orderLast:
    - Namespace
    - Deployment
`,
			expectedOutput: `
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
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
kind: Service
metadata:
  name: papaya
---
apiVersion: v1
kind: LimitRange
metadata:
  name: peach
---
apiVersion: v1
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Deployment
metadata:
  name: pear
`,
		},
		{
			name:      "fifo order",
			resources: resources,
			transformer: `
apiVersion: builtin
kind: SortOrderTransformer
metadata:
  name: notImportantHere
sortOptions:
  order: fifo
`,
			expectedOutput: resources,
		},
		{
			name:      "some first, some last, some in-between",
			resources: resources,
			transformer: `
apiVersion: builtin
kind: SortOrderTransformer
metadata:
  name: notImportantHere
sortOptions:
  order: legacy
  legacySortOptions:
    orderFirst:
    - ValidatingWebhookConfiguration
    - Service
    orderLast:
    - Namespace
    - Deployment
`,
			expectedOutput: `
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
---
apiVersion: v1
kind: Service
metadata:
  name: papaya
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apricot
---
apiVersion: v1
kind: Ingress
metadata:
  name: durian
---
apiVersion: v1
kind: LimitRange
metadata:
  name: peach
---
apiVersion: v1
kind: Role
metadata:
  name: banana
---
apiVersion: v1
kind: Secret
metadata:
  name: quince
---
apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Deployment
metadata:
  name: pear
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			th := kusttest_test.MakeEnhancedHarness(t).PrepBuiltin("SortOrderTransformer")
			defer th.Reset()
			th.AssertActualEqualsExpected(
				th.LoadAndRunTransformer(test.transformer, test.resources),
				test.expectedOutput,
			)
		})
	}
}
