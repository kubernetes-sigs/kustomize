---
title: "commonAnnotations"
linkTitle: "commonAnnotations"
type: docs
description: >
    Add annotations to add all resources.
---

Add annotations to all resources.  If the annotation key already is present on the resource,
the value will be overridden.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-555-1212
```

## Example

### File Input

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-555-1212

resources:
- deploy.yaml
```

```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  ...
```

### Build Output

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  annotations:
    oncallPager: 800-555-1212
spec:
  ...
```
