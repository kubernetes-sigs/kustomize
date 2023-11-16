---
title: "ObjectMeta"
linkTitle: "ObjectMeta"
weight: 4
date: 2023-11-15
description: >
  ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.
---

ObjectMeta partially copies [`k8s.io/apimachinery/pkg/apis/meta/v1`].

---

* **name** (string)

	Name must be unique within a namespace.

* **namespace** (string)

	Namespace defines the space within which each name must be unique.

* **labels** (map[string]string]

	Map of string keys and values that can be used to organize and categorize (scope and select) objects.

* **annotations** (map[string]string]

	Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata.



[`k8s.io/apimachinery/pkg/apis/meta/v1`]: https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/
