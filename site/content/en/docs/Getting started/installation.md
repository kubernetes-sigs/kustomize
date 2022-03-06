---
title: "Install Kustomize"
linkTitle: "Install Kustomize"
date: 2022-02-27
weight: 10
description: >
  Installing Kustomize
---

Kustomize can be installed in a variety of ways.

## Binaries
Binaries are available for Linux, MacOS and Windows, across a variety of architectures.

You can see the full list of releases here on the [Github releases page](https://github.com/kubernetes-sigs/kustomize/releases).

### Quick install
Get the latest build of Kustomize for your platform.
```bash
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
```
You can also pass optional `version` and `target_dir` arguments to the script:
```bash
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" 4.4.1 $HOME/bin  | bash
```

❗**This script doesn’t work for ARM architecture**. If you want to install ARM binaries, please refer to the [releases page](https://github.com/kubernetes-sigs/kustomize/releases).

## Packages
Kustomize is also available in some package repositories.

### Debian/Ubuntu
```bash
sudo apt-get install kustomize
```
### Arch
```bash
pacman -S kustomize
```

### Mac
[Homebrew](https://brew.sh/):
```bash
brew install kustomize
```

[MacPorts](https://www.macports.org/):
```bash
sudo port install kustomize
```

### Windows
[Chocolatey](https://community.chocolatey.org/packages/kustomize)
```bash
choco install kustomize
```

## Docker
Docker images for kustomize are published on the [GCR Container Registry](https://console.cloud.google.com/gcr/images/k8s-artifacts-prod/US/kustomize/kustomize).
```bash
docker run k8s.gcr.io/kustomize/kustomize:v4.5.1 version
```

## go get
<!--
TODO: is this still the way to do this? v3 seems old and my go module knowledge isn't great
-->
Requires [Go](https://go.dev/) to be installed.
```bash
GOBIN=$(pwd)/ GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3
```

## Source
<!--
TODO: once again here, are these instructions up-to-date? We probably should bump to a later version of kustomize but I'm also not sure if these go env variables are modern practice
-->
Clone the kustomize Github repo and build using go.
```bash
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