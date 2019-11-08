# Pseudo Modules

This package contains dependencies copied from kubernetes/kubernetes repos which
are synced out of staging.

The long term plan is to move off of the staging libraries entirely in favor of
more suitable libraries developed in the Kustomize repo.

## Why?

1. Vendoring the Kustomize API in other tools

The Kubernetes staging packages do not have stable APIs, and frequently break compatibility.
This makes it difficult for other tools to vendor the Kustomize APIs, as they may depend
on incompatible versions of the staging APIs.  By forking the staging libraries, we
ensure that we are using our own copy which will not conflict with other versions.

2. Vendoring into kubectl

Packages that depend upon staging may not be vendored into kubernetes/kubernetes.  By forking
the staging packages, we break this circular dependency so that the kustomize packages may
be vendored into kubernetes/kubernetes without depending on code originating out of
kubernetes/kubernetes.

## Who?

While it is possible to depend upon them from modules outside the Kustomize repository,
there is not guarantee that this will continue to work in the future.

The pseudo modules may be removed at anytime in the future without warning and no
support will be given for these modules.

## How?

These libraries were forked by running `git clone` to clone the repos.

### Automated Creation Steps

1. Remove the current existing psuedo modules
  - `$ rm -rf psuedo/k8s`
2. Run the [init-pseudo-module.sh](init-pseudo-module.sh) script to clone and configure pseudo deps
  - From the root directory -- `$ psuedo/init-pseudo-module.sh`

### Using the Pseudo Modules in Kustomize

TODO(pwittrock): Write this once it has been done successfully
