---
title: "namespace"
linkTitle: "namespace"
type: docs
description: >
    Adds namespace to all resources.
---

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: my-namespace
```

Will override the existing namespace if it is set on a resource, or add it
if it is not set on a resource.
