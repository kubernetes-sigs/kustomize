---
title: "Transformers"
linkTitle: "Transformers"
weight: 3
date: 2023-07-28
description: >
  Transformers have the ability to modify user-input resources.
---

Transformers are Kubernetes objects that dictate how Kustomize changes other Kubernetes objects that users provide.
Besides `spec`, transformers require the same [fields](https://kubernetes.io/docs/concepts/overview/working-with-objects/#required-fields), listed below, as other Kubernetes objects:

* `apiVersion`
* `kind`
* `metadata`
  * `name`
