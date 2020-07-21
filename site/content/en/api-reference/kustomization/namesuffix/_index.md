---
title: "nameSuffix"
linkTitle: "nameSuffix"
type: docs
description: >
    Appends the value to the names of all resources and references.
---

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

nameSuffix: -v2
```

A deployment named `wordpress` would become `wordpress-v2`.

**Note:** The suffix is appended before the content hash if the resource type is ConfigMap or Secret.
