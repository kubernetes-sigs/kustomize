# Forked K8S Dependencies

This package contains dependencies forked from kubernetes/kubernetes/staging

The long term plan is to move off of the staging libraries entirely in favor of
more suitable libraries developed in the Kustomize repo.

## Why

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

## How

These libraries were forked by running `git clone` to clone the repos.  From the `forked` directory:

1. Clone the repos
  - `git clone https://github.com/kubernetes/api`
  - `git clone https://github.com/kubernetes/apimachinery`
  - `git clone https://github.com/kubernetes/client-go`
2. Remove spurious OWNERS files
  - `find . -name OWNERS | xargs rm`
3. Remove the `.git` dirs
  - `rm -rf api/.git`
  - `rm -rf apimachinery/.git`
  - `rm -rf client-go/.git`
4. Rename the packages
  - `find . -name go.mod | xargs sed -i -e 's$k8s.io/client-go$sigs.k8s.io/kustomize/forked/client-go$g'`
  - `find . -name go.mod | xargs sed -i -e 's$k8s.io/api$sigs.k8s.io/kustomize/forked/api$g'`
  - `find . -name *.go | xargs sed -i -e 's$k8s.io/api$sigs.k8s.io/kustomize/forked/api$g'`
  - `find . -name *.go | xargs sed -i -e 's$k8s.io/client-go$sigs.k8s.io/kustomize/forked/client-go$g'`
