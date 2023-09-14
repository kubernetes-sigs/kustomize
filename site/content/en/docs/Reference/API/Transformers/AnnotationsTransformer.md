---
title: "AnnotationsTransformer"
linkTitle: "AnnotationsTransformer"
weight: 1
date: 2023-07-28
description: >
  AnnotationsTransformer adds annotations to user-input resources.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

* **apiVersion**: builtin
* **kind**: AnnotationsTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **annotations** (map[string]string)

  Map of annotations that AnnotationsTransformer will add to resources.

  If not specified, AnnotationsTransformer leaves the resources unchanged.

* **fieldSpecs** (\[\][FieldSpec]({{< relref "../Common%20Definitions/FieldSpec.md" >}}))

  fieldSpecs specifies the field on each resource that AnnotationsTransformer should add the annotations to.
  It essentially allows the user to re-define the field path of the Kubernetes annotations field from `metadata/annotations` for different resources.

  If not specified, AnnotationsTransformer applies the annotations to the `metadata/annotations` field of all resources.
