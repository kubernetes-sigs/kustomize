---
title: "Off The Shelf Application"
linkTitle: "Off The Shelf Application"
type: docs
weight: 2
description: >
    Workflow for off the shelf applications
---

In this workflow, all files are owned by the user and maintained in a repository under their control, but
they are based on an [off-the-shelf] configuration that is periodically consulted for updates.

![off-the-shelf config workflow image][workflowOts]

#### 1) find and [fork] an [OTS] config

#### 2) clone it as your [base]

The [base] directory is maintained in a repo whose upstream is an [OTS] configuration, in this case
some user's `ldap` repo:

> ```
> mkdir ~/ldap
> git clone https://github.com/$USER/ldap ~/ldap/base
> cd ~/ldap/base
> git remote add upstream git@github.com:$USER/ldap
> ```

#### 3) create [overlays]

As in the bespoke case above, create and populate an _overlays_ directory.

The [overlays] are siblings to each other and to the [base] they depend on.

> ```
> mkdir -p ~/ldap/overlays/staging
> mkdir -p ~/ldap/overlays/production
> ```

The user can maintain the `overlays` directory in a
distinct repository.

#### 4) bring up [variants]

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

#### 5) (optionally) capture changes from upstream

The user can periodically [rebase] their [base] to
capture changes made in the upstream repository.

> ```
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
