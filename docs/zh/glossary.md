# 词汇表

[CRD spec]: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/
[CRD]: #custom-resource-definition
[DAM]: #声明式应用程序管理
[Declarative Application Management]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/declarative-application-management.md
[JSON]: https://www.json.org/
[JSONPatch]: https://tools.ietf.org/html/rfc6902
[JSONMergePatch]: https://tools.ietf.org/html/rfc7386
[Resource]: #resource
[YAML]: http://www.yaml.org/start.html
[application]: #application
[apply]: #apply
[apt]: https://en.wikipedia.org/wiki/APT_(Debian)
[base]: #base
[bases]: #base
[bespoke]: #bespoke-configuration
[gitops]: #gitops
[k8s]: #kubernetes
[kubernetes]: #kubernetes
[kustomize]: #kustomize
[kustomization]: #kustomization
[kustomizations]: #kustomization
[off-the-shelf]: #off-the-shelf-configuration
[overlay]: #overlay
[overlays]: #overlay
[patch]: #patch
[patches]: #patch
[patchJson6902]: #patchjson6902
[patchExampleJson6902]: https://github.com/kubernetes-sigs/kustomize/blob/master/examples/jsonpatch.md
[patchesJson6902]: #patchjson6902
[proposal]: https://github.com/kubernetes/community/pull/1629
[rebase]: https://git-scm.com/docs/git-rebase
[资源]: #资源
[resources]: #resource
[root]: #kustomization-root
[rpm]: https://en.wikipedia.org/wiki/Rpm_(software)
[strategic-merge]: https://git.k8s.io/community/contributors/devel/sig-api-machinery/strategic-merge-patch.md
[target]: #target
[transformer]: #transformer
[variant]: #variant
[variants]: #variant
[workflow]: workflows.md

## 应用

**应用**是为某种目的关联起来的一组 Kubernetes 资源，例如一个前有负载均衡器，后有数据库的 Web 服务器。用标签、命名和元数据将[资源]组织起来，可以进行**添加**或**删除**等操作。

有提案（[Declarative Application Management]）描述了一种称为应用的新的 Kubernetes 资源。更加正式的描述了这一思路，并在应用程序级别提供了运维和仪表盘的支持。

[Kustomize] 对 Kubernetes 资源进行配置，其中描述的应用程序资源只是另一种普通的资源。

## Apply

**Apply** 这个动词在 Kubernetes 的上下文中，指的是一个 Kubernetes 命令以及能够对集群施加影响的进程内 [API 端点](https://goo.gl/UbCRuf)。

用户可以将对集群的运行要求用一组完整的资源列表的形式进行表达，通过 **apply** 命令进行提交。

集群把新提交的资源和之前提交的状态以及当前的实际状态进行合并，形成新的状态。这就是 Kubernetes 的状态管理过程。

## Base

**Base** 指的是会被其它 [Kustomization] 引用的 [Kustomization]。

包括 [Overlay] 在内的任何 Kustomization，都可以作为其它 Kustomization 的 Base。

Base 对引用自己的 Overlay 并无感知。

Base 和 [Overlay] 可以作为 Git 仓库中的唯一内容，用于简单的 [GitOps] 管理。对仓库的变更可以触发构建、测试以及部署过程。

## 定制配置

**定制**配置是由组织为满足自身需要，在内部创建和管理的 [Kustomization] 和[资源]。

和**定制配置**关联的 [Workflow] 比关联到通用配置的 [Workflow] 要简单一些，原因是通用配置是共享的，需要周期性的跟踪他人作出的变更。

## Custom resource definition

可以通过定制 CRD ([CRD spec]) 的方式对 Kubernetes API 进行扩展。

CRD 定义的[资源]是一种全新的资源，可以和 ConfigMap、Deployment 之类的原生资源以相同的方式来使用。

Kustomize 能够生成自定义资源，但是要完成这个目标，必须给出对应的 CRD，这样才能正确的对这种结构进行处理。

## 声明式应用程序管理

Kustomize 鼓励对声明式应用程序管理（[Declarative Application Management]）的支持，这种方式是一系列 Kubernetes 集群管理的最佳实践。Kustomize 应该可以：

- 适用于任何配置，例如自有配置、共享配置、无状态、有状态等。
- 支持通用配置，以及创建变体（例如开发、预发布、生产）。
- 开放使用原生 Kubernetes API，而不是隐藏它们。
- 不会给版本控制系统和集成的评审和审计工作造成困难。
- 用 Unix 的风格和其它工具进行协作。
- 避免使用模板、领域特定的语言等额外的学习内容。

## 生成器

生成器生成的资源，可以直接使用，也可以输出给转换器（[Transformer]）。

## GitOps

一种 DevOps 或者 CICD 流程，这种流程以 Git 作为唯一的事实，并且在这种事实发生变化时采取措施（例如构建、测试和部署）。

## Kustomization

**Kustomization** 这个词可以指 `kustomization.yaml` 这个文件，更常见的情况是一个包含了 `kustomization.yaml` 及其所有直接引用文件的相对路径（所有不需要 URL 的本地数据）。

也就是说，如果在 [Kustomize] 的上下文中说到 **Kustomization**，可能是以下的情况之一：

- 一个叫做 `kustomization.yaml` 的文件。
- 一个压缩包（包含 YAML 文件以及它的引用文件）。
- 一个 Git 压缩包。
- 一个 Git 仓库的 URL。

一个 Kustomization 文件包含的[字段](fields.md)，分为四个类别：

- `resources`：待定制的现存[资源]，示例字段：`resources`、`crds`。
- `generator`：将要创建的**新**资源，示例字段：`configMapGenerator`（传统）、`secretGenerator`（传统）、`generators`（v2.1）
- `transformer`：对前面提到的新旧资源进行**处理**的方式。示例字段：`namePrefix`、`nameSuffix`、`images`、`commonLabels`、`patchesJson6902` 等。在 v2.1 中还有更多的 `transformer`。
- `meta`：会对上面几种字段产生影响。示例字段：`vars`、`namespace`、`apiVersion`、`kind` 等。

## Kustomization root

直接包含 `kustomization.yaml` 文件的目录。

处理 Kustomization 文件时，可能访问到该目录以内或以外的文件。

像 YAML 资源这样的数据文件，或者用于生成 ConfigMap 或 Secret 的包含 `name=value` 的文本文件，或者用于补丁转换的补丁文件，必须**在这个目录的内部**，需要显式的使用**相对路径**来表达。

v2.1 中有一个特殊选项 `--load_restrictions none` 能够放宽这个限制，从而让不同的 Kustomization 可以共享补丁文件。

可以用 URL、绝对路径或者相对路径引用其它的 Kustomization（包含 `kustomization.yaml` 文件的其它目录）。

如果 `A` Kustomization 依赖 `B` Kustomization，那么：

- `B` 不能包含 `A`。
- `B` 不能依赖 `A`，间接依赖也不可以。

`A` 可以包含 `B`，但是这样的话，最简单的方式可能是让 `A` 直接依赖 `B` 的资源，并去除 `B` 的 `kustomization.yaml` 文件（就是把 `B` 合并到 `A`）。

通常情况下，`B` 和 `A` 处于同级目录，或者 `B` 放在一个完全独立的 Git 仓库里，可以从任意的 Kustomization 进行引用。

常见布局大致如下：

> ```
> ├── base
> │   ├── deployment.yaml
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── dev
>     │   ├── kustomization.yaml
>     │   └── patch.yaml
>     ├── prod
>     │   ├── kustomization.yaml
>     │   └── patch.yaml
>     └── staging
>         ├── kustomization.yaml
>         └── patch.yaml
> ```

`dev`、`prod` 以及 `staging` 是否依赖于 `base`，要根据 `kustomization.yaml` 具体判断。

## Kubernetes

[Kubernetes](https://kubernetes.io) 是一个开源软件，为容器化应用提供了自动部署、伸缩和管理的能力。

它经常会被简写为 `k8s`。

## Kubernetes 风格的对象

[必要字段]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

用 YAML 或者 JSON 文件表达一个对象，其中包含一些[必要字段]。`kind` 字段用于标识对象类型，`metadata/name` 字段用于区分实例，`apiVersion` 表示的是版本（如果有多个版本的话）。

## Kustomize

`kustomize` 是一个面向 Kubernetes 的命令行工具，用一种无模板、结构化的的方式为为声明式配置提供定制支持。

`面向 Kubernetes` 的意思是 Kustomize 对 API 资源、Kubernetes 概念（例如名称、标签、命名空间等）、以及资源补丁是有支持的。

Kustomize 是 [DAM] 的一个实现。

## 通用配置

通用配置是一种用于共享的 Kustomization 以及资源。

例如创建一个这样的 Github 仓库：

> ```
> github.com/username/someapp/
>   kustomization.yaml
>   deployment.yaml
>   configmap.yaml
>   README.md
> ```

其他人可以 `fork` 这个仓库，并把它们的 Fork `clone` 到本地进行定制。

用户可以用这个克隆回来的版本作为 [Base]，在此基础上定制 [Overlay] 来满足自身需求。

## Overlay

`Overlay` 是一个 依赖于其它 Kustomization 的 Kustomization。

Overlay 引用（通过文件路径、URI 或者别的什么方式）的 [Kustomization] 被称为 [Base]。

Overlay 无法脱离 Base 独立生效。

Overlay 可以作为其它 Overlay 的 Base。

通常 Overlay 都是不止一个的，因为实际情况中就需要为单一 Base 创建不同的[变体]，例如 `development`、`QA`、`production` 等。

总的说来，这些变体使用的资源基本是一致的，只有一些简单的差异，例如 Deployment 的实例数量、特定 Pod 的 CPU 资源、ConfigMap 中的数据源定义等。

可以这样把配置提交到集群：

> ```
>  kustomize build someapp/overlays/staging |\
>      kubectl apply -f -
>
>  kustomize build someapp/overlays/production |\
>      kubectl apply -f -
> ```

对 Base 的使用是隐性的——Overlay 的依赖是指向 Base 的。

请参考 [root]。

## 包

在 Kustomize 中，`包`是没有意义的，Kustomize 并无意成为 [apt]、[rpm] 那样的传统包管理工具。

## Patch

修改资源的通用指令。

有两种功能类似但是实现不同的补丁方式：[strategic merge patch](#patchstrategicmerge) 和 [JSON patch](#patchjson6902)。

## patchStrategicMerge

`patchStrategicMerge` 是 [strategic-merge] 风格的补丁（SMP）。

SMP 看上去像个不完整的 Kubernetes 资源 YAML。SMP 中包含 `TypeMeta` 字段，用于表明这个补丁的目标[资源]的 `group/version/kind/name`，剩余的字段是一个嵌套的结构，用于指定新的字段值，例如镜像标签。

缺省情况下，SMP 会**替换**目标值。如果目标值是一个字符串，这种行为是合适的，但是如果目标值是个列表，可能就不合适了。

可以加入 `directive` 来修改这种行为，，可以接受的 `directive` 包括 `replace`（缺省）、`merge`（不替换列表）、`delete` 等（[相关说明][strategic-merge]）。

注意对自定义资源来说，SMP 会被当作 [json merge patches][JSONMergePatch].

有趣的事实：所有的资源文件都可以当作 SMP 使用，相同 `group/version/kind/name` 资源中的匹配字段会被替换，其它内容则保持不变。

## patchJson6902

`patchJson6902` 引用一个 Kubernetes 资源，并用 [JSONPatch] 指定了修改这一资源的方法。

`patchJson6902` 几乎可以做到所有 `patchStrategicMerge` 的功能，但是语法更加简单，参考[示例][patchExampleJson6902]

## 插件

Kustomize 可以使用的一段代码，但是无需编译到 Kustomize 内部，可以作为 Kustomization 的一部分，生成或转换 Kubernetes 资源。

[插件](../plugins)的细节。

## 资源

在 REST-ful API 的上下文中，资源是 `GET`、`PUT` 或者 `POST` 等 HTTP 操作的目标。Kubernetes 提供了 REST-ful API 界面，用于和客户端进行交互。

在 Kustomization 的上下文中，资源是一个相对于 [root] 的相对路径，指向 [YAML] 或者 [JSON] 文件，描述了一个 Kubernetes API 对象，例如 Deployment 或者 ConfigMap，或者一个 Kustomization、或者一个指向 Kustomization 的 URL。

或者说任何定义了[对象](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields)的格式正确的 YAML 文件，其中包含了 `kind` 和 `metadata/name` 字段，都是资源。

## Root

参看 [kustomization root][root].

## sub-target / sub-application / sub-package

不存在 `sub-xxx`，只有 [Base] 和 [Overlay]。

## Target

`target` 是 `kustomize build` 的参数，例如：

> ```
>  kustomize build $target
> ```

`$target` 必须是一个指向 [Kustomization] 的路径或者 URL。

要创建用于进行 [Apply] 操作的资源，`target` 中必须包含或者引用所有相关信息。

[Base] 或者 [Overlay] 都可以作为 `target`。

## Transformer

转换器能够修改资源，或者在 `kustomize build` 的过程中获取资源的信息。

## 变体

在集群中把 [Overlay] 应用到 [Base] 上的产物称为**变体**。

比如 `staging` 和 `production` 两个 Overlay，都修改了同样的 Base，来创建各自的变体。

`staging` 变体包含了一组用来保障测试过程的资源，或者一些想要看到生产环境下一个版本的外部用户。

`production` 变体用于承载生产流量，可能使用大量的副本，分配更多的 CPU 和内存。
