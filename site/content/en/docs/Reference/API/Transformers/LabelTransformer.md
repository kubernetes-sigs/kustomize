---
title: "LabelTransformer"
linkTitle: "LabelTransformer"
weight: 2
date: 2024-02-12
description: >
  LabelTransformer adds labels to user-input resources.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

* **apiVersion**: builtin
* **kind**: LabelTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **labels** (map[string]string)

  Map of labels that LabelTransformer will add to resources.

  If not specified, LabelTransformer leaves the resources unchanged.

* **fieldSpecs** (\[\][FieldSpec]({{< relref "../Common%20Definitions/FieldSpec.md" >}}))

  fieldSpecs specifies the field on each resource that LabelTransformer should add the labels to.
  It essentially allows the user to re-define the field path of the Kubernetes labels field from `metadata/labels` for different resources.

  If not specified, LabelTransformer applies the labels to the `metadata/labels` field of all resources.
