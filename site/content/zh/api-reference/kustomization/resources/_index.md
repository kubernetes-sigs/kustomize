---
title: "resources"
linkTitle: "resources"
type: docs
description: >
    包含的资源。
---

该条目可以是指向本地目录的相对路径，也可以是指向远程仓库中的目录的 URL，例如：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- myNamespace.yaml
- sub-dir/some-deployment.yaml
- ../../commonbase
- github.com/kubernetes-sigs/kustomize/examples/multibases?ref=v1.0.6
- deployment.yaml
- github.com/kubernets-sigs/kustomize/examples/helloWorld?ref=test-branch
```

将以深度优先的顺序读取和处理资源。

文件应包含 YAML 格式的 k8s 资源。一个资源描述文件可以含有多个由（`---`）分隔的资源。
应该包含 `resources` 字段的 kustomization 文件的指定文件目录的相对路径。

[hashicorp URL]: https://github.com/hashicorp/go-getter#url-format

目录规范可以是相对、绝对或部分的 URL。URL 规范应遵循 [hashicorp URL] 格式。该目录必须包含 `kustomization.yaml` 文件。
