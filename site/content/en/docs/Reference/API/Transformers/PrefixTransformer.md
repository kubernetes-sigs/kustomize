---
title: "PrefixTransformer"
linkTitle: "PrefixTransformer"
weight: 4
date: 2024-02-12
description: >
  PrefixTransformer adds prefixes to the names of user-input resources.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

* **apiVersion**: builtin
* **kind**: PrefixTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **prefix** (string)

  Prefix is the value that PrefixTransformer will prepend to the names of resources.

  If not specified, PrefixTransformer leaves the names of resources unchanged.

* **fieldSpecs** (\[\][FieldSpec]({{< relref "../Common%20Definitions/FieldSpec.md" >}}))

  fieldSpecs specifies the field on each resource that PrefixTransformer should add the prefix to.
  It essentially allows the user to re-define the field path of the Kubernetes name field from `metadata/name` for different resources.

  If not specified, PrefixTransformer applies the prefix to the `metadata/name` field of all resources.
