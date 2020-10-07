---
title: "patches"
linkTitle: "patches"
type: docs
description: >
    Patch resources
---

[strategic merge]: /kustomize/zh/api-reference/glossary#patchstrategicmerge
[JSON]: /kustomize/zh/api-reference/glossary#patchjson6902

Patches 在资源上添加或覆盖字段，Kustomization 使用 `patches` 字段来提供该功能。

`patches` 字段包含要按指定顺序应用的 patch 列表。

patch 可以:

- 是一个 [strategic merge] patch，或者是一个 [JSON] patch。
- 也可以是 patch 文件或 inline string
- 针对单个资源或多个资源

目标选择器可以通过 group、version、kind、name、namespace、标签选择器和注释选择器来选择资源，选择一个或多个匹配所有**指定**字段的资源来应用 patch。

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patches:
- path: patch.yaml
  target:
    group: apps
    version: v1
    kind: Deployment
    name: deploy.*
    labelSelector: "env=dev"
    annotationSelector: "zone=west"
- patch: |-
    - op: replace
      path: /some/existing/path
      value: new value
  target:
    kind: MyKind
    labelSelector: "env=dev"
```

patch 目标选择器的 `name` 和 `namespace` 字段是自动锚定的正则表达式。这意味着 `myapp` 的值相当于 `^myapp$`。

示例 1 和示例 2 都将使用以下 `deployment.yaml`：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
      - name: the-container
        image: registry/conatiner:latest
```

## 示例1

### 目的

将容器镜像指向特定版本，代替 latest 版本。

### 文件输入

```yaml
# kustomization.yaml
resources:
- deployment.yaml

patches:
- path: patch.yaml
```

```yaml
# patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  template:
    containers:
      - name: the-container
        image: registry/conatiner:1.0.0
```

### 构建输出

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
    - image: registry/conatiner:1.0.0
      name: the-container
```

## 示例2

### 目的

同上。

### 文件输入

```yaml
# kustomization.yaml
resources:
- deployment.yaml

patches:
- target:
    kind: Deployment
    name: the-deployment
  path: patch.json
```

```yaml
# patch.json
[
   {"op": "replace", "path": "/spec/template/containers/0/image", "value": "registry/conatiner:1.0.0"}
]

```

### 构建输出

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
    - image: registry/conatiner:1.0.0
      name: the-container
```
