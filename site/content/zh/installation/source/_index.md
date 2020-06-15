---
title: "Go 源码"
linkTitle: "Go 源码"
weight: 2
type: docs
description: >
    使用 Go 源码安装 Kustomize。
---

需要先安装 [Go]。

## 无需克隆源码库直接构建 kustomize CLI

```bash
GOBIN=$(pwd)/ GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3
```

## 在本地克隆源码库构建 kustomize CLI

```bash
# 需要 go 1.13 或更高版本
unset GOPATH
# 详见 https://golang.org/doc/go1.13#modules
unset GO111MODULES

# 拉取 repo
git clone git@github.com:kubernetes-sigs/kustomize.git
# 进入目录
cd kustomize

# 如果您不想从 HEAD 开始构建， 则可以选择切换特定的标签
git checkout kustomize/v3.2.3

# 开始构建
(cd kustomize; go install .)

# 运行
~/go/bin/kustomize version
```

[Go]: https://golang.org
