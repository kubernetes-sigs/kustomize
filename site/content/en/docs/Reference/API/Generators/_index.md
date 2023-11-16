---
title: "Generators"
linkTitle: "Generators"
weight: 2
date: 2023-11-15
description: >
  Generators can create Kubernetes objects.
---

Generators can create Kubernetes objects with user provided arguments.

Besides `spec`, transformers require the same [fields](https://kubernetes.io/docs/concepts/overview/working-with-objects/#required-fields), listed below, as other Kubernetes objects:

* `apiVersion`
* `kind`
* `metadata`
  * `name`
