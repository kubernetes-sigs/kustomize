---
title: "sortOptions"
linkTitle: "sortOptions"
type: docs
weight: 22
description: >
    Change the strategy used to sort resources at the end of the Kustomize build.
---

The `sortOptions` field is used to sort the resources kustomize outputs. It is
available in kustomize v5.0.0+.

IMPORTANT:
- Currently, this field is respected only in the top-level Kustomization (that
  is, the immediate target of `kustomize build`). Any instances of the field in
  Kustomizations further down the build chain (for example, in bases included
  through the `resources` field) will be ignored.
- This field is the endorsed way to sort resources. It should be used instead of
  the `--reorder` CLI flag, which is deprecated.

Currently, we support the following sort options:
- `legacy`
- `fifo`

```yaml
kind: Kustomization
sortOptions:
  order: legacy | fifo # "legacy" is the default
```

## FIFO Sorting

In `fifo` order, kustomize does not change the order of resources. They appear
in the order they are loaded in `resources`.

### Example 1: FIFO Sorting

```yaml
kind: Kustomization
sortOptions:
  order: fifo
```

## Legacy Sorting

The `legacy` sort is the default order, and is used when the sortOrder field is
unspecified.

In `legacy` order, kustomize sorts resources by using two priority lists:
- An `orderFirst` list for resources which should be first in the output.
- An `orderLast` list for resources which should be last in the output.
- Resources not on the lists will appear in between, sorted using their apiVersion and kind fields.

### Example 2: Legacy Sorting with orderFirst / orderLast lists

In this example, we use the `legacy` sort order to output `Namespace` objects
first and `Deployment` objects last.

```yaml
kind: Kustomization
sortOptions:
  order: legacy
  legacySortOptions:
    orderFirst:
    - Namespace
    orderLast:
    - Deployment
```

### Example 3: Default Legacy Sorting

If you specify `legacy` sort order without any arguments for the lists,
kustomize will fall back to the lists we were using before introducing this
feature. Since legacy sort is the default, this is also equivalent to not
specifying the field at all.

These two configs are equivalent:

```yaml
kind: Kustomization
sortOptions:
  order: legacy
```

is equivalent to:

```yaml
kind: Kustomization
sortOptions:
  order: legacy
  legacySortOptions:
    orderFirst:
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
    orderLast:
    - MutatingWebhookConfiguration
    - ValidatingWebhookConfiguration
```
