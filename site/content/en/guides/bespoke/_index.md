---
title: "Bespoke Application"
linkTitle: "Bespoke Application"
type: docs
weight: 1
description: >
    Workflow for bespoke applications
---

In this workflow, all configuration (resource YAML) files are owned by the user.
No content is incorporated from version control repositories owned by others.

![bespoke config workflow image][workflowBespoke]

#### 1) create a directory in version control

Speculate some overall cluster application called _ldap_;
we want to keep its configuration in its own repo.

> ```
> git init ~/ldap
> ```

#### 2) create a [base]

> ```
> mkdir -p ~/ldap/base
> ```

In this directory, create and commit a [kustomization]
file and a set of [resources].

#### 3) create [overlays]

> ```
> mkdir -p ~/ldap/overlays/staging
> mkdir -p ~/ldap/overlays/production
> ```

Each of these directories needs a [kustomization]
file and one or more [patches].

The _staging_ directory might get a patch
that turns on an experiment flag in a configmap.

The _production_ directory might get a patch
that increases the replica count in a deployment
specified in the base.

#### 4) bring up [variants]

Run kustomize, and pipe the output to [apply].

> ```
> kustomize build ~/ldap/overlays/staging | kubectl apply -f -
> kustomize build ~/ldap/overlays/production | kubectl apply -f -
> ```

You can also use [kubectl-v1.14.0] to apply your [variants].
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
