---
title: "commonLabels"
linkTitle: "commonLabels"
type: docs
description: >
    Add labels and selectors to add all resources.
---

Add labels and selectors to all resources.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```
