---
title: "ReplacementTransformer"
linkTitle: "ReplacementTransformer"
weight: 6
date: 2026-05-07
description: >
  ReplacementTransformer copies values from a source field or literal value into target fields.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

`ReplacementTransformer` uses the same replacement schema as the
[`replacements` field]({{< relref "../Kustomization File/replacements.md" >}}) in a
Kustomization file. It can be invoked explicitly through the `transformers` field
when the replacement configuration should live in a separate transformer file.

* **apiVersion**: builtin
* **kind**: ReplacementTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **replacements** ([]Replacement)

  List of replacements to apply. Each replacement copies a value from `source` or
  `sourceValue` into one or more `targets`.

  Each item may also be specified as a `path` to a file containing a single
  replacement or a list of replacements.

Example transformer file:

```yaml
apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: replacement-transformer
replacements:
- source:
    kind: Deployment
    name: source-deployment
    fieldPath: metadata.name
  targets:
  - select:
      kind: Service
      name: target-service
    fieldPaths:
    - metadata.annotations.[example.com/source-name]
    options:
      create: true
```

Example Kustomization using the transformer file:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
- service.yaml
transformers:
- replacement-transformer.yaml
```

For the full replacement schema, field descriptions, selector behavior, and
field path syntax, see the [`replacements` field documentation]({{< relref "../Kustomization File/replacements.md" >}}).
