---
title: "commonAnnotations"
linkTitle: "commonAnnotations"
type: docs
description: >
    Add annotations to add all resources.
---

Add annotations (non-identifying metadata) to all resources.  These are key value pairs.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-555-1212
```
