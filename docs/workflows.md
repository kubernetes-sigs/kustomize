[OTS]: glossary.md#off-the-shelf
[apply]: glossary.md#apply
[applying]: glossary.md#apply
[base]: glossary.md#base
[fork]: https://guides.github.com/activities/forking/
[instances]: glossary.md#instance
[manifest]: glossary.md#manifest
[off-the-shelf]: glossary.md#off-the-shelf
[overlays]: glossary.md#overlay
[patch]: glossary.md#patch
[patches]: glossary.md#patch
[rebase]: https://git-scm.com/docs/git-rebase
[resources]: glossary.md#resources
[workflowBespoke]: workflowBespoke.jpg
[workflowOts]: workflowOts.jpg

# workflows

A _workflow_ is the sequence of steps one takes to
use and maintain a configuration.

## Bespoke configuration

In this workflow, all configuration files are owned by
the user.  No content is incorporated from version
control repositories owned by others.

![bespoke config workflow image][workflowBespoke]

#### 1) create a directory in version control

> ```
> git init ~/ldap
> ```

#### 2) create a [base]

> ```
> mkdir -p ~/ldap/base
> ```

In this directory, create and commit a [manifest]
and a set of [resources].

#### 3) create [overlays]

> ```
> mkdir -p ~/ldap/overlays/staging
> mkdir -p ~/ldap/overlays/production
> ```

Each of these directories needs a [manifest]
and one or more [patches].

The _staging_ directory might get a patch
that turns on an experiment flag in a configmap.

The _production_ directory might get a patch
that increases the replica count in a deployment
specified in the base.

#### 4) bring up [instances]

Run kustomize, and pipe the output to [apply].

> ```
> kustomize ~/ldap/overlays/staging | kubectl apply -f -
> kustomize ~/ldap/overlays/production | kubectl apply -f -
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
https://github.com/kinflate/ldap.

> ```
> mkdir ~/ldap
> git clone https://github.com/$USER/ldap ~/ldap/base
> cd ~/ldap/base
> git remote add upstream git@github.com:kustomize/ldap
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


#### 4) bring up instances

> ```
> kustomize ~/ldap/overlays/staging | kubectl apply -f -
> kustomize ~/ldap/overlays/production | kubectl apply -f -
> ```

#### 5) (optionally) capture changes from upstream

The user can optionally [rebase] their [base] to
capture changes made in the upstream repository.

> ```
> cd ~/ldap/base
> git fetch upstream
> git rebase upstream/master
> ```
