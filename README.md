# kustomize

[applied]: docs/glossary.md#apply
[base]: docs/glossary.md#base
[declarative configuration]: docs/glossary.md#declarative-application-management
[demo]: demos/README.md
[demos]: demos/README.md
[imageBase]: docs/base.jpg
[imageOverlay]: docs/overlay.jpg
[kubernetes style]: docs/glossary.md#kubernetes-style-object
[KEP]: https://github.com/kubernetes/community/blob/master/keps/sig-cli/0008-kustomize.md
[kustomization]: docs/glossary.md#kustomization
[overlay]: docs/glossary.md#overlay
[resources]: docs/glossary.md#resource
[sig-cli]: https://github.com/kubernetes/community/blob/master/sig-cli/README.md
[workflows]: docs/workflows.md


`kustomize` is a command line tool supporting
template-free customization of YAML (or JSON) objects
that conform to the [kubernetes style].  If your
objects have a `kind` and a `metadata` field,
`kustomize` can patch them to support configuration
sharing and re-use.

For more details, try a [demo].

[![Build Status](https://travis-ci.org/kubernetes-sigs/kustomize.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/kustomize)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-sigs/kustomize)](https://goreportcard.com/report/github.com/kubernetes-sigs/kustomize)


## Installation

This assumes [Go](https://golang.org/) (v1.10.1 or higher)
is installed and your `PATH` contains `$GOPATH/bin`:

<!-- @installkustomize @test -->
```
go get github.com/kubernetes-sigs/kustomize
```


## Usage

#### 1) Make a customized base

A [base] configuration is a [kustomization] file listing a set of
k8s [resources] - deployments, services, configmaps,
secrets that serve some common purpose.

![base image][imageBase]

File structure:

> ```
> ~/yourApp
> └── base
>     ├── deployment.yaml
>     ├── kustomization.yaml
>     └── service.yaml
> ```

Your base could be a fork of someone else's
configuration, that your occasionally rebase from to
capture improvements.

#### 2) Further customize with overlays

An [overlay] customizes your base along different dimensions
for different purposes or different teams, e.g. for
_development_ and _production_.

![overlay image][imageOverlay]

File structure:
> ```
> ~/yourApp
> ├── base
> │   ├── deployment.yaml
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── development
>     │   ├── cpu_count.yaml
>     │   ├── kustomization.yaml
>     │   └── replica_count.yaml
>     └── production
>         ├── cpu_count.yaml
>         ├── kustomization.yaml
>         └── replica_count.yaml
> ```

Your overlays could sit in your own repository.

#### 3) Run kustomize

Run `kustomize` on an overlay, e.g.

```
kustomize build ~/yourApp/overlays/production
```

The result is printed to `stdout` as a set of complete
resources, ready to be [applied] to a cluster.  See the
[demos].



## About

This project is sponsored by [sig-cli] ([KEP]).
