# kustomize

[applied]: docs/glossary.md#apply
[base]: docs/glossary.md#base
[declarative configuration]: docs/glossary.md#declarative-application-management
[demo]: demos/README.md
[imageBase]: docs/base.jpg
[imageOverlay]: docs/overlay.jpg
[manifest]: docs/glossary.md#manifest
[overlay]: docs/glossary.md#overlay
[resources]: docs/glossary.md#resource
[workflows]: docs/workflows.md

`kustomize` is a command line tool supporting
template-free customization of declarative
configuration targetted to kubernetes.

## Installation

Assumes [Go](https://golang.org/) is installed
and your `PATH` contains `$GOPATH/bin`:

<!-- @installkustomize @test -->
```
go get k8s.io/kubectl/cmd/kustomize
```

## Usage

#### 1) Make a base

A [base] configuration is a [manifest] listing a set of
k8s [resources] - deployments, services, configmaps,
secrets that serve some common purpose.

![base image][imageBase]

#### 2) Customize it with overlays

An [overlay] customizes your base along different dimensions
for different purposes or different teams, e.g. for
_development, staging and production_.

![overlay image][imageOverlay]

#### 3) Run kustomize

Run kustomize on your overlay.  The result
is printed to `stdout` as a set of complete
resources, ready to be [applied] to a cluster.

For more details, try a [demo].
