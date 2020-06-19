---
title: "nameSuffix"
linkTitle: "nameSuffix"
type: docs
description: >
    为所有资源和引用的名称添加后缀。
---

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

nameSuffix: -v2
```

deployment 名称从 `wordpress` 变为 `wordpress-v2`。

**注意:** 如果资源类型是 ConfigMap 或 Secret，则在哈希值之前添加后缀。
