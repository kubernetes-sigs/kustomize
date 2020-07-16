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
