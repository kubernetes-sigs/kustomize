---
title: "replicas"
linkTitle: "replicas"
type: docs
description: >
    修改资源的副本数。
---

对于如下 kubernetes Deployment 片段：

```yaml
# deployment.yaml
kind: Deployment
metadata:
  name: deployment-name
spec:
  replicas: 3
```

在 kustomization 中添加以下内容，将副本数更改为 5：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

replicas:
- name: deployment-name
  count: 5
```

该字段内容为列表，所以可以同时修改许多资源。

由于这个声明无法设置 `kind:` 或 `group:`，所以他只能匹配如下资源中的一种：

- `Deployment`
- `ReplicationController`
- `ReplicaSet`
- `StatefulSet`

对于更复杂的用例，请使用 patch 。
