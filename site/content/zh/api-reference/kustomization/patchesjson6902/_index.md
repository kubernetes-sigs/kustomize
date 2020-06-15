---
title: "patchesJson6902"
linkTitle: "patchesJson6902"
type: docs
description: >
    使用 [json 6902 标准](https://tools.ietf.org/html/rfc6902) Patch resources
---

patchesJson6902 列表中的每个条目都应可以解析为 kubernetes 对象和将应用于该对象的 [JSON patch](https://tools.ietf.org/html/rfc6902)。

目标字段指向的 kubernetes 对象的 group、 version、 kind、 name 和 namespace 在同一 kustomization 内 path 字段内容是 JSON patch 文件的相对路径。

patch 文件中的内容可以如下这种 JSON 格式：

```json
 [
   {"op": "add", "path": "/some/new/path", "value": "value"},
   {"op": "replace", "path": "/some/existing/path", "value": "new value"}
 ]
 ```

也可以使用 YAML 格式表示：

```yaml
- op: add
  path: /some/new/path
  value: value
- op: replace
  path: /some/existing/path
  value: new value
```

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  path: add_init_container.yaml
- target:
    version: v1
    kind: Service
    name: my-service
  path: add_service_annotation.yaml
```

patch 内容也可以是一个inline string：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  patch: |-
    - op: add
      path: /some/new/path
      value: value
    - op: replace
      path: /some/existing/path
      value: "new value"
```
