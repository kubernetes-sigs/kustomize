---
title: "crds"
linkTitle: "crds"
type: docs
description: >
    增加对 CRD 的支持。
---

此列表中的每个条目都应该是自定义资源定义（CRD）文件的相对路径。

该字段的存在是为了让 kustomize 识别用户自定义的 CRD ，并对这些类型中的对象应用适当的转换。

典型用例：CRD 引用 ConfigMap 对象

在 kustomization 中，ConfigMap 对象名称可能会通过 `namePrefix` 、`nameSuffix` 或 `hashing` 来更改 CRD 对象中该 ConfigMap 对象的名称，
引用时需要以相同的方式使用 `namePrefix` 、 `nameSuffix` 或 `hashing` 来进行更新。

Annotations 可以放入 openAPI 的定义中：

- "x-kubernetes-annotation": ""
- "x-kubernetes-label-selector": ""
- "x-kubernetes-identity": ""
- "x-kubernetes-object-ref-api-version": "v1",
- "x-kubernetes-object-ref-kind": "Secret",
- "x-kubernetes-object-ref-name-key": "name",

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

crds:
- crds/typeA.yaml
- crds/typeB.yaml
```
