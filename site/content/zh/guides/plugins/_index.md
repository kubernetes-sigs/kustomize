---
title: "Kustomize 插件"
linkTitle: "Kustomize 插件"
type: docs
weight: 98
description: >
    Kustomize 插件指南
---

Kustomize 提供一个插件框架，允许用户开发自己的 _生成器_ 和 _转化器_。

[generator options]: https://github.com/kubernetes-sigs/kustomize/tree/master/examples/generatorOptions.md
[transformer configs]: https://github.com/kubernetes-sigs/kustomize/tree/master/examples/transformerconfigs

通过插件，实现 [generatorOptions] 和 [transformerconfigs] 无法满足的需求。

[12-factor]: https://12factor.net

* _generator_ 插件生成 k8s 资源，比如 [helm chart inflator] 是一个 generator 插件，基于少量自由变量生成一个 [12-factor] 应用所包含的全部组件 deployment，service，scaler，ingress 等）也是一个 generator 插件。
* _transformer_ 插件转化（修改）k8s 资源，比如可能会执行对特殊容器命令行的编辑，或为其他内置转换器（`namePrefix`、`commonLabels` 等）无法转换的内容提供转换。

## `kustomization.yaml` 的格式

从为添加 `generators` 或 `transformers` 字段开始。

字段内容为一个 string list：

> ```yaml
> generators:
> - relative/path/to/some/file.yaml
> - relative/path/to/some/kustomization
> - /absolute/path/to/some/kustomization
> - https://github.com/org/repo/some/kustomization
>
> transformers:
> - {as above}
> ```

格式要求与 `resources` 字段相同，`generators` 或 `transformers` 列表的每一列内容都必须是一个 YAML 文件的相对路径或者指向 [kustomization] 的 URL。

[kustomization]: /kustomize/zh/api-reference/glossary#kustomization

从磁盘上读取 YAML 文件，kustomization 的路径或 URL 会触发 kustomization 的运行。由此产生的每个的对象都会被 kustomize 进一步解析为 _plugin configuration_ 对象。

## 配置

kustomization 文件可以包含如下内容：

```yaml
generators:
- chartInflator.yaml
```

像这样，kustomization 进程将在 [kustomization root](glossary.md#kustomization-root) 下寻找到一个名为 `chartInflator.yaml` 的文件。

`chartInflator.yaml` 为插件配置文件，该文件包含 YAML 配置对象，内容如下：

```yaml
apiVersion: someteam.example.com/v1
kind: ChartInflator
metadata:
  name: notImportantHere
chartName: minecraft
```

__`apiVersion` 和 `kind` 字段用于定位插件。__

[k8s 对象]: /kustomize/zh/api-reference/glossary#Kubernetes-风格的对象

同时由于 kustomize 插件配置对象也是一个 [k8s 对象]，因此这些字段是必要的。

为了让插件准备好生成或转换，它包含了配置文件的全部内容。

[NameTransformer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin/prefixsuffixtransformer/PrefixSuffixTransformer_test.go
[ChartInflator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/chartinflator/ChartInflator_test.go
[plugins]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin

更多关于插件配置 YAML 的例子，请浏览根目录下 [plugins] 中的单元测试，例如 [ChartInflator] 或 [NameTransformer]。

## 植入

每个插件都有自己的专用目录，名为：

[`XDG_CONFIG_HOME`]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

```bash
$XDG_CONFIG_HOME/kustomize/plugin
    /${apiVersion}/LOWERCASE(${kind})
```

The default value of [`XDG_CONFIG_HOME`] is
`$HOME/.config`.

[`XDG_CONFIG_HOME`] 的默认值为 `$HOME/.config`。

为了便于插件包（源码、测试、插件数据文件等）的共享，要求每个目录存放一个插件。

在 [Go 插件](#go-插件)中，还可以为单个插件提供一个 `go.mod` 文件，可以缓解包版本依赖性偏移的问题。

加载时，kustomize 首先会寻找一个 _可执行_ 文件，名为：

```bash
$XDG_CONFIG_HOME/kustomize/plugin
    /${apiVersion}/LOWERCASE(${kind})/${kind}
```

如果没有找到这个文件，或者这个文件不是可执行的，kustomize 会在同一目录下寻找一个名为 `${kind}.so` 的文件，并尝试将其作为 [Go插件](#go-插件) 加载。

如果这两项检查都失败，则插件加载失败，`kustomize build` 失败。

## 执行情况

插件只有在运行 `kustomize build` 命令时使用。

生成器插件是在处理完 `resources` 字段后运行的（`resources` 字段本身也可以看成是一个简单地从磁盘上读取对象的生成器）。

之后所有资源将被传递到转换管道中，由其中内置的转换器，如 `namePrefix` 和 `commonLabel` 等先转换应用（如果 kustomization 文件中指定了他们），然后再转换用户指定的 `transformers` 字段。

由于无法指定转化的顺序，所以需要遵守 `transformers` 字段中指定的顺序。

#### No Security

Kustomize 插件不会在任何形式的 kustomize 提供的沙盒中运行。不存在 _"plugin security"_ 的概念。

`kustomize build` 会尝试使用插件，但如果省略了 `--enable_alpha_plugins`，将导致插件无法加载，并且会有一个关于插件使用的警告。

使用这个 flag 就是承认使用不稳定的插件 API（alpha）、承认使用缺少出处插件，以及插件不属于 kustomize 的事实。

简单的说，一些从网上下载的 kustomize 插件可能会奇妙地将 k8s 的配置以理想的方式进行改造，同时也会悄悄地对运行 `kustomize build` 的系统做任何用户可以做的事情。

## 编写插件

插件有 [exec](#exec-插件) 和 [Go](#go-插件) 两种.

### Exec 插件

_exec 插件_ 是一个可以在命令行中接收参数可执行文件，该参数指向包含 kustomization 配置的 YAML 文件。

> TODO: 对插件的限制，允许同一个 _exec 插件_ 同时被 `generators` 和 `transformers` 字段所触发。
>
> - 第一个参数可以是固定的字符串 `generate` 或 `transform`，（配置文件的名称移动到第2个参数）
> - 默认情况下，exec plugin 会作为一个转化器,除非提供了标志 `-g`，将 exec 插件切换为生成器。

[helm chart inflator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/chartinflator
[bashed config map]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/bashedconfigmap
[sed transformer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/sedtransformer
[hashicorp go-getter]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/gogetter

#### 示例

* [helm chart inflator] - helm chart inflates 生成器。
* [bashed config map] - 使用 bash 生成十分简单的 configMap。
* [sed transformer] - 使用插件来定义非结构化的编辑。
* [hashicorp go-getter] - 下载 kustomize layes 并通过构建它来生成资源。

生成器插件无需在 `stdin` 上输入任何东西，就会将生成的资源输出到 `stdout`。

转化器插件需要在 `stdin` 上输入资源的 YAML，并转化后的资源输出到 `stdout`。

kustomize 会使用 exec 插件适配器，为 `stdin` 提供的资源，并获取 `stdout` 以进行下一步的处理。

#### Generator 选项

生成器 exec 插件可以通过设置以下内部注释中的一个来调整生成器选项。

> 注意：这些注释只会在本地的 kustomize 中，不会出现在最终输出中。

**`kustomize.config.k8s.io/needs-hash`**

通过包含 `needs-hash` 注释，可以将资源标记为需要由内部哈希转换器处理的资源。当设置注释的有效值为 `"true"` 和 `"false"` 时，分别启用或禁用资源的哈希后缀。忽略该注解相当于将值设置为 `"false"`。

如果此注释被设置在不受哈希转换器支持的资源上，将导致构建将失败。

示例：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
  annotations:
    kustomize.config.k8s.io/needs-hash: "true"
data:
  foo: bar
```

**`kustomize.config.k8s.io/behavior`**

`behavior` 注释为当资源发生冲突时插件的处理方式，有效值包括："create"、"merge "和 "replace"，默认为 "create"。

示例：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
  annotations:
    kustomize.config.k8s.io/behavior: "merge"
data:
  foo: bar
```

### Go 插件

请务必阅读 [Go 插件注意事项](goPluginCaveats.md)。

[Go 插件]: https://golang.org/pkg/plugin/

如果一个 `.go` 文件声明 `package main`，并附加了有用的功能标志，那么它就可以成为一个 [Go 插件]。

如果标志被命名为 “KustomizePlugin”，并且附加的函数实现了 `Configurable`、`Generator` 和 `Transformer` 接口，那么它可以进一步作为 _kustomize_ 插件使用。

kustomize 的一个 Go 插件看起来是这样的：

> ```go
> package main
>
> import (
> "sigs.k8s.io/kustomize/api/ifc"
> "sigs.k8s.io/kustomize/api/resmap"
>   ...
> )
>
> type plugin struct {...}
>
> var KustomizePlugin plugin
>
> func (p *plugin) Config(
>    ldr ifc.Loader,
>    rf *resmap.Factory,
>    c []byte) error {...}
>
> func (p *plugin) Generate() (resmap.ResMap, error) {...}
>
> func (p *plugin) Transform(m resmap.ResMap) error {...}
> ```

需要使用标识符 `plugin`，`KustomizePlugin` 并且需要实现方法签名 `Config`。

实现 `Generatoror` 或 `Transformer` 方法允许（分别）将插件的配置文件添加到 kustomization 文件的 `generatorsor` 或 `transformers` 字段中，并根据需要执行。

[secret generator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/secretsfromdatabase
[service generator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/someservicegenerator
[string prefixer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/stringprefixer
[date prefixer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/dateprefixer
[sops encoded secrets]: https://github.com/monopole/sopsencodedsecrets

#### 示例

* [service generator] - 使用 name 和 port 参数生成一个 service。
* [string prefixer] - 使用 `metadata/name` 值作为前缀。这个特殊的示例是为了展示插件的转化行为。详见 `target` 包中的 `TestTransformedTransformers` 测试。
* [date prefixer] - 将当前日期作为前缀添加到资源名称上，这是一个用于修改刚才提到的字符串前缀插件的简单示例。
* [secret generator] - 从 toy 数据库生成 secret。
* [sops encoded secrets] - 一个更复杂的 secret 生成器。
* [All the builtin plugins](https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin).
   用户自制的插件与内置插件是一样的。

Go 插件既可以是生成器，也可以是转化器。`Generate` 方法将在 `Transform` 方法运行之前与所有其他生成器一起运行。

如下的构建命令，假设插件源代码位于 kustomize 期望查找 `.so` 文件的目录中：

```bash
d=$XDG_CONFIG_HOME/kustomize/plugin\
/${apiVersion}/LOWERCASE(${kind})

go build -buildmode plugin \
   -o $d/${kind}.so $d/${kind}.go
```
