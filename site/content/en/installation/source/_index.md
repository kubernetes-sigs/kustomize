---
title: "Go Source"
linkTitle: "Go Source"
weight: 2
type: docs
description: >
    Install Kustomize from the Go source code
---

Requires [Go] to be installed.

## Install the kustomize CLI from source without cloning the repo

```
GOBIN=$(pwd)/ GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3
```

## Install the kustomize CLI from local source with cloning the repo

```
# Need go 1.13 or higher
unset GOPATH
# see https://golang.org/doc/go1.13#modules
unset GO111MODULES

# clone the repo
git clone git@github.com:kubernetes-sigs/kustomize.git
# get into the repo root
cd kustomize

# Optionally checkout a particular tag if you don't
# want to build at head
git checkout kustomize/v3.2.3

# build the binary
(cd kustomize; go install .)

# run it
~/go/bin/kustomize version
```

[Go]: https://golang.org
