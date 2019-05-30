[OTS]: ../glossary.md#off-the-shelf-configuration
[apply]: ../glossary.md#apply
[applying]: ../glossary.md#apply
[base]: ../glossary.md#base
[fork]: https://guides.github.com/activities/forking/
[variants]: ../glossary.md#variant
[kustomization]: ../glossary.md#kustomization
[off-the-shelf]: ../glossary.md#off-the-shelf-configuration
[overlays]: ../glossary.md#overlay
[patch]: ../glossary.md#patch
[patches]: ../glossary.md#patch
[rebase]: https://git-scm.com/docs/git-rebase
[resources]: ../glossary.md#resource
[workflowBespoke]: ../images/workflowBespoke.jpg
[workflowOts]: ../images/workflowOts.jpg
[kubectl-v1.14.0]:https://kubernetes.io/blog/2019/03/25/kubernetes-1-14-release-announcement/

# 工作流

工作流是 kustomize 运行和维护配置的步骤。

## 配置定制（Bespoke configuration）

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

在这个目录中创建并提交 [kustomization] 文件及一组资源配置。

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
> ```
> kubectl apply -k ~/ldap/overlays/staging
> kubectl apply -k ~/ldap/overlays/production
> ```

## 使用现成的配置（Off-the-shelf configuration）

在这个工作流方式中，可从别人的 repo 中 fork kustomize 配置，并根据自己的需求来配置。


![off-the-shelf config workflow image][workflowOts]

#### 1) 寻找并且 [fork] 一个 [OTS] 配置

#### 2) 将其克隆为你自己的 [base]

这个 [base] 目录维护在上游为 [OTS] 配置的 repo ，在这个例子使用 `ladp` 的 repo 。

> ```
> mkdir ~/ldap
> git clone https://github.com/$USER/ldap ~/ldap/base
> cd ~/ldap/base
> git remote add upstream git@github.com:$USER/ldap
> ```

#### 3) 创建 [overlays]

如配置定制方法一样，创建并完善 _overlays_ 目录中的内容。

所有的 [overlays] 都依赖于 [base] 。

> ```
> mkdir -p ~/ldap/overlays/staging
> mkdir -p ~/ldap/overlays/production
> ```

用户可以将 `overlays` 维护在不同的 repo 中。

#### 4) 生成 [variants]

> ```
> kustomize build ~/ldap/overlays/staging | kubectl apply -f -
> kustomize build ~/ldap/overlays/production | kubectl apply -f -
> ```

也可以在 [kubectl-v1.14.0] 版，使用 ```kubectl``` 命令发布你的 [variants] 。
> ```
> kubectl apply -k ~/ldap/overlays/staging
> kubectl apply -k ~/ldap/overlays/production
> ```

#### 5) （可选）从上游更新

用户可以定期从上游 repo 中 [rebase] 他们的 [base] 以保证及时更新。

> ```
> cd ~/ldap/base
> git fetch upstream
> git rebase upstream/master
> ```
