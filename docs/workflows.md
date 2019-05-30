[OTS]: glossary.md#off-the-shelf-configuration
[apply]: glossary.md#apply
[applying]: glossary.md#apply
[base]: glossary.md#base
[fork]: https://guides.github.com/activities/forking/
[variants]: glossary.md#variant
[kustomization]: glossary.md#kustomization
[off-the-shelf]: glossary.md#off-the-shelf-configuration
[overlays]: glossary.md#overlay
[patch]: glossary.md#patch
[patches]: glossary.md#patch
[rebase]: https://git-scm.com/docs/git-rebase
[resources]: glossary.md#resource
[workflowBespoke]: images/workflowBespoke.jpg
[workflowOts]: images/workflowOts.jpg
[kubectl-v1.14.0]:https://kubernetes.io/blog/2019/03/25/kubernetes-1-14-release-announcement/

# workflows

A _workflow_ is the sequence of steps one takes to
use and maintain a configuration.

## Bespoke configuration

In this workflow, all configuration (resource YAML) files
are owned by the user.  No content is incorporated from version
control repositories owned by others.

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
> ```
> kubectl apply -k ~/ldap/overlays/staging
> kubectl apply -k ~/ldap/overlays/production
> ```

## Off-the-shelf configuration

In this workflow, all files are owned by the user and
maintained in a repository under their control, but
they are based on an [off-the-shelf] configuration that
is periodically consulted for updates.


![off-the-shelf config workflow image][workflowOts]

#### 1) find and [fork] an [OTS] config

#### 2) clone it as your [base]

The [base] directory is maintained in a repo whose
upstream is an [OTS] configuration, in this case
some user's `ldap` repo:

> ```
> mkdir ~/ldap
> git clone https://github.com/$USER/ldap ~/ldap/base
> cd ~/ldap/base
> git remote add upstream git@github.com:$USER/ldap
> ```

#### 3) create [overlays]

As in the bespoke case above, create and populate
an _overlays_ directory.

The [overlays] are siblings to each other and to the
[base] they depend on.

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
