---
title: "namePrefix"
linkTitle: "namePrefix"
type: docs
description: >
    为所有资源和引用的名称添加前缀。
---

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: alices-
```

deployment 名称从 `wordpress` 变为 `alices-wordpress`。
