// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestLegacyOrderTransformer(t *testing.T) {

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
kind: LegacyOrderTransformer
metadata:
  name: notImportantHere
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
			name:      "webhooks first",
			resources: resources,
			transformer: `
apiVersion: builtin
kind: LegacyOrderTransformer
metadata:
  name: notImportantHere
legacySortOptions:
  orderFirst:
  - MutatingWebhookConfiguration
  - ValidatingWebhookConfiguration
  - Namespace
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
  - Deployment
  - StatefulSet
  - CronJob
  - PodDisruptionBudget
`,
			expectedOutput: `
apiVersion: v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pomegranate
---
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
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			th := kusttest_test.MakeEnhancedHarness(t).
				PrepBuiltin("LegacyOrderTransformer")
			defer th.Reset()
			th.RunTransformerAndCheckResult(test.transformer, test.resources, test.expectedOutput)
		})
	}

}
