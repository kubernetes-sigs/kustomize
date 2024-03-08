---
title: "SuffixTransformer"
linkTitle: "SuffixTransformer"
weight: 5
date: 2024-02-12
description: >
  SuffixTransformer adds suffixes to the names of user-input resources.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

* **apiVersion**: builtin
* **kind**: SuffixTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **suffix** (string)

  Suffix is the value that SuffixTransformer will postfix to the names of resources.

  If not specified, SuffixTransformer leaves the names of resources unchanged.

* **fieldSpecs** (\[\][FieldSpec]({{< relref "../Common%20Definitions/FieldSpec.md" >}}))

  fieldSpecs specifies the field on each resource that SuffixTransformer should add the suffix to.
  It essentially allows the user to re-define the field path of the Kubernetes name field from `metadata/name` for different resources.

  If not specified, SuffixTransformer applies the suffix to the `metadata/name` field of all resources.
