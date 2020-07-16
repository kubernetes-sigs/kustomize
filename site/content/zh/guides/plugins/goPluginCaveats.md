---
title: "Go 插件注意事项"
linkTitle: "Go 插件注意事项"
type: docs
weight: 3
description: >
    Go 插件注意事项
---

[plugin package]: https://golang.org/pkg/plugin
[Go modules]: https://github.com/golang/go/wiki/Modules
[ELF]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format
[tensorflow plugin]: https://www.tensorflow.org/guide/extend/op

_Go 插件_ 是一个编译产品/组件，其定义见 [plugin package]，需要特殊的构建标志，不能单独运行，必须加载到正在运行的 Go 程序中。

> 用 Go 编写的普通程序可以作为 _exec 插件_，但是不能作为 _Go 插件_。

Go 插件允许运行 kustomize 扩展，而无需在每次运行时将资源分配到子流程或从子流程中解封所有资源数据。Go 插件 API 确保一定程度的一致性，以避免混淆下游转换器。

Go 插件的工作方式与 [plugin package] 中所述的相同，但与 _plugin_ 一词相关的常见概念不同。

## The skew problem

Go 插件编译会创建一个 [ELF] 格式的 `.so` 文件，根据定义，该文件不包含有关目标代码来源的信息。

主程序 ELF 和插件 ELF 的编译条件（软件包依赖项的版本 `GOOS`，`GOARCH`）之间的偏移会导致插件加载失败，并带有无用的错误消息。

Exec 插件也会缺乏来源，但不会因编译不正确而失败。

在任何情况下，共享插件的最好方法是使用某种 _捆绑包_（git repo URL、git 存档文件、tar 包等），其中包含可解包至 `$XDG_CONFIG_HOME/kustomize/plugin` 的源代码，测试和相关数据。

对于 Go 插件，使用共享插件的最终用户 _必须同时编译 kustomize 和 plugin_。

这意味着一次性运行

```bash
# Or whatever is appropriate at time of reading
GOPATH=${whatever} GO111MODULE=on go get sigs.k8s.io/kustomize/api
```

然后使用一个正常的开发周期

```bash
go build -buildmode plugin \
    -o ${wherever}/${kind}.so ${wherever}/${kind}.go
```

并根据需要调整路径和发行版本标记（例如 `v3.0.0`）。

为了进行比较，可以参考编写 [tensorflow plugin] 必须做的事情。

## 为什么支持 Go 插件

### 安全

Go 插件开发者可以操作与原生 kustomize 操作相同的 API，可确保某些语义、变量和检查等一致。exec 插件子进程通过 stdin/stdout 来处理这些问题，但对于下游的转化器和使用者来说，会更容易把事情搞砸。

关键点：如果插件通过 kustomize 提供的文件 `Loader` 接口读取文件，则会受到 kustomize 文件加载限制的约束。当然，除了代码审计之外，没有什么可以阻止 Go 插件导入 io 包并执行其所需的任何操作。

### Debugging

Go 插件开发者可以在功能测试中运行插件时，在 _本地_ 调试插件，并在插件内部和其他位置设置断点。

为了获得两全其美的方式（共享性和安全性），开发人员可以编写一个 `.go` 程序作为 _exec 插件_，同时可以被 `go generate` 程序处理生成 Go 插件（反之亦然）。

### 贡献单元化

所有内置的生成器和转换器本身都是 Go 插件。这意味着 kustomize 维护者可以将贡献的插件升级为内置插件，而无需更改代码（超出常规代码审阅要求的范围）。

### 围绕生态系统发展

工具可以简化 Go 插件的 _共享_，但是这需要大量的 Go 插件的创作，而这又会导致围绕共享插件的混乱。[Go modules] 一旦被更广泛地采用，将解决共享插件最大的难题：含糊不清的插件 vs 主机依赖性。
