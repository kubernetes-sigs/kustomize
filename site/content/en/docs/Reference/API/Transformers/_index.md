---
title: "Transformers"
linkTitle: "Transformers"
weight: 3
date: 2023-07-28
description: >
  Transformers have the ability to modify user-input resources.
---

Transformers are Kubernetes objects that dictate how Kustomize changes other Kubernetes objects that users provide.
The following required Kubernetes fields are also required on transformers:

* `apiVersion`
* `kind`
* `metadata`
  * `name`
