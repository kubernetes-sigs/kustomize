---
title: "ReplacementTransformer"
linkTitle: "ReplacementTransformer"
weight: 6
date: 2026-07-23
description: >
  ReplacementTransformer copies values from a source field or literal value into target fields.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

`ReplacementTransformer` uses the same replacement schema as the
[`replacements` field]({{< relref "../Kustomization%20File/replacements.md" >}})
in a Kustomization file. Use it through the `transformers` field when the
replacement configuration should live in a separate transformer file.

* **apiVersion**: builtin
* **kind**: ReplacementTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **replacements** ([]ReplacementField)

  List of replacements to apply. Each item is either an inline replacement or a
  `path` to a file that contains a single replacement or a list of replacements.
  An item must not set both `path` and an inline replacement.

  Each inline replacement copies a value from `source` or `sourceValue` into one
  or more `targets`. It must specify exactly one of `source` or `sourceValue`,
  and at least one target.

Example transformer file:

```yaml
apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: replacement-transformer
replacements:
- source:
    kind: Deployment
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.1.image
```

Example Kustomization that references the transformer file:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
transformers:
- replacement-transformer.yaml
```

For the full replacement schema, field descriptions, selector behavior, and
field path syntax, see the
[`replacements` field documentation]({{< relref "../Kustomization%20File/replacements.md" >}}).
