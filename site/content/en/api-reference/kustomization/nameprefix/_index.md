---
title: "namePrefix"
linkTitle: "namePrefix"
type: docs
description: >
    Prepends the value to the names of all resources and references.
---

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: alices-
```

A deployment named `wordpress` would become `alices-wordpress`.
