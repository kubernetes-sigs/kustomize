---
title: "patchesStrategicMerge"
linkTitle: "patchesStrategicMerge"
type: docs
description: >
     使用 strategic merge patch 标准 Patch resources.
---

此列表中的每个条目都应可以解析为 [StrategicMergePatch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md).

这些（也可能是部分的）资源文件中的 name 必须与已经通过 `resources` 加载的 name 字段匹配，或者通过 `bases` 中的 name 字段匹配。这些条目将用于 _patch_（修改）已知资源。

推荐使用小的 patches，例如：修改内存的 request/limit，更改 ConfigMap 中的 env 变量等。小的 patches 易于维护和查看，并且易于在 overlays 中混合使用。

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesStrategicMerge:
- service_port_8888.yaml
- deployment_increase_replicas.yaml
- deployment_increase_memory.yaml
```

patch 内容也可以是一个inline string：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: nignx:latest
```

请注意，kustomize 不支持同一个 patch 对象中包含多个 _删除_ 指令。要从一个对象中删除多个字段或切片元素，需要创建一个单独的 patch，以执行所有需要的删除。
