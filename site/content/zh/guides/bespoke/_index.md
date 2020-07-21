---
title: "配置定制（Bespoke configuration）"
linkTitle: "配置定制（Bespoke configuration）"
type: docs
weight: 1
description: >
    自定义配置的工作流。
---

在这个工作流方式中，所有的配置文件（ YAML 资源）都为用户所有，存储在用户的私有 repo 中。其他用户是无法使用的。

![bespoke config workflow image][workflowBespoke]

#### 1) 创建一个目录用于版本控制

我们希望将一个名为 _ldap_ 的 Kubernetes 集群应用的配置保存在自己的 repo 中。
这里使用 git 进行版本控制。

> ```
> git init ~/ldap
> ```

#### 2) 创建一个 [base]

> ```
> mkdir -p ~/ldap/base
> ```

在这个目录中创建并提交 [kustomization] 文件及一组资源 [resources] 配置。

#### 3) 创建 [overlays]

> ```
> mkdir -p ~/ldap/overlays/staging
> mkdir -p ~/ldap/overlays/production
> ```

每个目录都包含需要一个 [kustomization] 文件以及一或多个 [patches]。

在 _staging_ 目录可能会有一个用于在 configmap 中打开一个实验标记的补丁。

在 _production_ 目录可能会有一个在 deployment 中增加副本数的补丁。

#### 4) 生成 [variants]

运行 kustomize，将生成的配置用于 kubernetes 应用发布。

> ```
> kustomize build ~/ldap/overlays/staging | kubectl apply -f -
> kustomize build ~/ldap/overlays/production | kubectl apply -f -
> ```

也可以在 [kubectl-v1.14.0] 版，使用 ```kubectl``` 命令发布你的 [variants] 。
>
> ```
> kubectl apply -k ~/ldap/overlays/staging
> kubectl apply -k ~/ldap/overlays/production
> ```

[OTS]: /kustomize/api-reference/glossary#off-the-shelf-configuration
[apply]: /kustomize/api-reference/glossary#apply
[applying]: /kustomize/api-reference/glossary#apply
[base]: /kustomize/api-reference/glossary#base
[fork]: https://guides.github.com/activities/forking/
[variants]: /kustomize/api-reference/glossary#variant
[kustomization]: /kustomize/api-reference/glossary#kustomization
[off-the-shelf]: /kustomize/api-reference/glossary#off-the-shelf-configuration
[overlays]: /kustomize/api-reference/glossary#overlay
[patch]: /kustomize/api-reference/glossary#patch
[patches]: /kustomize/api-reference/glossary#patch
[rebase]: https://git-scm.com/docs/git-rebase
[resources]: /kustomize/api-reference/glossary#resource
[workflowBespoke]: /kustomize/images/workflowBespoke.jpg
[workflowOts]: /kustomize/images/workflowOts.jpg
[kubectl-v1.14.0]:https://kubernetes.io/blog/2019/03/25/kubernetes-1-14-release-announcement/
