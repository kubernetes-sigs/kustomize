---
title: "mergeKeys"
linkTitle: "mergeKeys"
type: docs
weight: 16
description: >
    Declare custom merge keys for CRD list fields that lack OpenAPI schemas.
---

`apiVersion: kustomize.config.k8s.io/v1beta1`

## Problem

Kustomize's strategic merge patch uses OpenAPI schema information to decide
whether a list field should be **merged** (by a merge key such as `name`) or
**replaced** entirely. For Kubernetes built-in types this works automatically.
For Custom Resources (CRDs) — such as a Flux `HelmRelease` with arbitrary
values nested under `spec.values` — there is no registered schema, so patches
that add items to those lists silently **wipe** the base items.

## Solution

The `mergeKeys` field lets you declare "for resources of this type, treat this
list field as an associative list merged on this key" — without registering an
OpenAPI schema or using the deprecated `configurations` feature.

### mergeKeys

`mergeKeys` accepts a list of `MergeKeySpec` objects:

| Field     | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `kind`    | string | no       | Resource kind to match (e.g. `HelmRelease`). Omit to match any kind. |
| `group`   | string | no       | API group to match (e.g. `helm.toolkit.fluxcd.io`). Omit to match any group. |
| `version` | string | no       | API version to match (e.g. `v2beta1`). Omit to match any version. |
| `path`    | string | yes      | Slash-separated path to the list field (e.g. `spec/values/myapp/env`). |
| `key`     | string | yes      | Field name within each list item to use as the merge key (e.g. `name`). |

## Example

### Problem without mergeKeys

Given a base `HelmRelease`:

```yaml
# base/helmrelease.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: myapp
spec:
  values:
    env:
    - name: LOG_LEVEL
      value: info
```

And an overlay patch that adds a new env var:

```yaml
# overlay/patch.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: myapp
spec:
  values:
    env:
    - name: DEBUG
      value: "true"
```

Without `mergeKeys`, the overlay **replaces** the entire `env` list, and
`LOG_LEVEL` is lost.

### Fix with mergeKeys

```yaml
# overlay/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../base

patches:
- path: patch.yaml

mergeKeys:
- kind: HelmRelease
  group: helm.toolkit.fluxcd.io
  path: spec/values/env
  key: name
```

Now the patch **merges** by `name`, and both env vars appear in the output:

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: myapp
spec:
  values:
    env:
    - name: DEBUG
      value: "true"
    - name: LOG_LEVEL
      value: info
```

## Notes

### `path` uses `/` as separator

`path` uses `/` as a separator (not `.`), matching the same convention as
`configurations` field specs.

### `key` must be a direct field on list items

`key` must be the name of a **direct, top-level field** on each list item — it
cannot be a nested path. For example, if items look like `{name: foo, value:
bar}`, use `key: name`. If the identifier is nested (e.g. `metadata.name` on a
PersistentVolumeClaim), `mergeKeys` cannot be used to merge that list.

### Paths reset at associative-list boundaries

When a target list is nested inside another associative list, the walker resets
its path counter at each list-element boundary. This means `path` must be
expressed **relative to the element that contains the target list**, not from
the document root.

For example, for a list at `spec.volumes[*].configMap.items` (where `volumes`
is itself an associative list), declare:

```yaml
mergeKeys:
- path: configMap/items   # relative to inside a volume element
  key: key
```

**not** `spec/volumes/configMap/items`.

### Scope

- `version` is optional — most CRD users only know group and kind.
- Multiple entries can target different paths or different resource types.
- `mergeKeys` affects strategic merge patches (`patches` with SM patch content)
  and `patchesStrategicMerge`. JSON 6902 patches are unaffected.
- This feature works for both CRDs (no registered schema) and native Kubernetes
  resource fields that have a schema but no merge key annotation.
- This feature does not modify or replace the `configurations` field.
