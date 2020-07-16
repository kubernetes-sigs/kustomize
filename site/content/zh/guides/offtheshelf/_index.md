---
title: "通用配置（Off-the-shelf configuration）"
linkTitle: "通用配置（Off-the-shelf configuration）"
type: docs
weight: 2
description: >
    使用通用配置的工作流。
---

在这个工作流程中，所有文件都由用户拥有，并维护在他们控制的存储库中，但它们是基于一个现成的（[off-the-shelf]）配置，定期查询更新。

![off-the-shelf config workflow image][workflowOts]

#### 1) 寻找并且 [fork] 一个 [OTS] 配置

#### 2) 将其克隆为你自己的 [base]

这个 [base] 目录维护在上游为 [OTS] 配置的 repo ，在这个示例中使用 `ladp` 的 repo 。

> ```bash
> mkdir ~/ldap
> git clone https://github.com/$USER/ldap ~/ldap/base
> cd ~/ldap/base
> git remote add upstream git@github.com:$USER/ldap
> ```

#### 3) 创建 [overlays]

如配置定制方法一样，创建并完善 _overlays_ 目录中的内容。

所有的 [overlays] 都依赖于 [base] 。

> ```bash
> mkdir -p ~/ldap/overlays/staging
> mkdir -p ~/ldap/overlays/production
> ```

用户可以将 `overlays` 维护在不同的 repo 中。

#### 4) 生成 [variants]

> ```bash
> kustomize build ~/ldap/overlays/staging | kubectl apply -f -
> kustomize build ~/ldap/overlays/production | kubectl apply -f -
> ```

也可以在 [kubectl-v1.14.0] 版，使用 ```kubectl``` 命令发布你的 [variants] 。
>
> ```bash
> kubectl apply -k ~/ldap/overlays/staging
> kubectl apply -k ~/ldap/overlays/production
> ```

#### 5) （可选）从上游更新

用户可以定期从上游 repo 中 [rebase] 他们的 [base] 以保证及时更新。

> ```bash
> cd ~/ldap/base
> git fetch upstream
> git rebase upstream/master
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
