// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

//nolint:gochecknoglobals
var sortOrderResources = `
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

//nolint:gochecknoglobals
var legacyOrderResources = `
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
`

func TestCustomOrdering(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

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
`)
	th.WriteF("base/resources.yaml", sortOrderResources)

	th.AssertActualEqualsExpected(
		th.Run("base", th.MakeDefaultOptions()),
		`
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
`)
}

func TestDefaultLegacyOrdering(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

sortOptions:
  order: legacy
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	th.AssertActualEqualsExpected(
		th.Run("base", th.MakeDefaultOptions()), legacyOrderResources)
}

func TestFIFOOrdering(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

sortOptions:
  order: fifo
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	th.AssertActualEqualsExpected(
		th.Run("base", th.MakeDefaultOptions()), sortOrderResources)
}

func TestInvalidLegacySortOptionsWithoutOrderKey(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

sortOptions:
  legacySortOptions: {}
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	err := th.RunWithErr("base", th.MakeDefaultOptions())
	require.ErrorContains(t, err, "the field 'sortOptions.order' must be one of [fifo, legacy]")
}

func TestInvalidLegacySortOptionsWithFIFOOrder(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

sortOptions:
  order: fifo
  legacySortOptions: {}
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	err := th.RunWithErr("base", th.MakeDefaultOptions())
	require.ErrorContains(t, err, "the field 'sortOptions.legacySortOptions' is set but the selected sort order is 'fifo', not 'legacy'")
}

// If the sort order is defined both in a CLI flag and the kustomization file,
// the kustomization file takes precedence.
func TestCLIAndKustomizationSet(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

sortOptions:
  order: fifo
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	kustOptions := th.MakeDefaultOptions()
	kustOptions.Reorder = krusty.ReorderOptionLegacy
	th.AssertActualEqualsExpected(th.Run("base", kustOptions), sortOrderResources)
}

// If no sort order is set in the kustomization file, validate that we fallback
// to the default legacy sort ordering.
func TestKustomizationSortOrderNotSet(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	kustOptions := th.MakeDefaultOptions()
	kustOptions.Reorder = krusty.ReorderOptionUnspecified
	th.AssertActualEqualsExpected(th.Run("base", kustOptions), legacyOrderResources)
}

func TestChildKustomizationSortOrder(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

sortOptions:
  order: legacy
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	th.WriteK("overlay", `
resources:
- ../base

sortOptions:
  order: fifo
`)

	kustOptions := th.MakeDefaultOptions()
	// If child kustomization ordering takes effect, the order will be legacy,
	// otherwise fifo.
	th.AssertActualEqualsExpected(th.Run("overlay", kustOptions), sortOrderResources)
}

func TestSortOrderGivenAsTransformer(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- resources.yaml

transformers:
- sort_order.yaml
`)
	th.WriteF("base/resources.yaml", sortOrderResources)
	th.WriteF("base/sort_order.yaml", `
apiVersion: builtin
kind: SortOrderTransformer
metadata:
  name: notImportantHere
sortOptions:
  order: fifo`)

	kustOptions := th.MakeDefaultOptions()
	err := th.RunWithErr("base", kustOptions)
	require.ErrorContains(t, err, "unable to load builtin SortOrderTransformer.builtin.[noGrp]")
}
