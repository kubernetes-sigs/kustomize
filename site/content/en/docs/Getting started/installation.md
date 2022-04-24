---
title: "Install Kustomize"
linkTitle: "Install Kustomize"
date: 2022-02-27
weight: 10
description: >
  Kustomize can be installed in a variety of ways
---

## Binaries

Binaries at various versions for Linux, macOS and Windows are published on the [releases page].

The following [script] detects your OS and downloads the appropriate kustomize binary to your
current working directory.

```bash
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
```

## Homebrew / MacPorts

For [Homebrew] users:

```bash
brew install kustomize
```

For [MacPorts] users:

```bash
sudo port install kustomize
```

## Chocolatey

```bash
choco install kustomize
```

For support on the chocolatey package
and prior releases, see:

- [Choco Package]
- [Package Source]

## Docker Images

Starting with Kustomize v3.8.7, docker images are available to run Kustomize.
The image artifacts are hosted on Google Container Registry (GCR).

See [GCR page] for available images.

The following commands are how to pull and run kustomize {{<example-semver-version>}} docker image.

```bash
docker pull k8s.gcr.io/kustomize/kustomize:{{< example-version >}}
docker run k8s.gcr.io/kustomize/kustomize:{{< example-version >}} version
```

## Go Source

Requires [Go] to be installed.

### Install the kustomize CLI from source without cloning the repo

```bash
go install sigs.k8s.io/kustomize/kustomize/{{< example-major-version >}}
```

### Install the kustomize CLI from local source

```bash
# Clone the repo
git clone git@github.com:kubernetes-sigs/kustomize.git
# Get into the repo root
cd kustomize

# Optionally checkout a particular tag if you don't want to build at head
git checkout kustomize/{{< example-version >}}

# Build the binary
(cd kustomize; go install .)

# Run it - this assumes your Go bin (generally GOBIN or GOPATH/bin) is on your PATH
# See the Go documentation for more details: https://go.dev/doc/code
kustomize version
```

[Go]: https://golang.org
[releases page]: https://github.com/kubernetes-sigs/kustomize/releases
[script]: https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh
[GCR page]: https://us.gcr.io/k8s-artifacts-prod/kustomize/kustomize
[Homebrew]: https://brew.sh
[MacPorts]: https://www.macports.org
[Choco Package]: https://chocolatey.org/packages/kustomize
[Package Source]: https://github.com/kenmaglio/choco-kustomize
