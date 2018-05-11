# kustomize

[applied]: docs/glossary.md#apply
[base]: docs/glossary.md#base
[declarative configuration]: docs/glossary.md#declarative-application-management
[demo]: demos/README.md
[demos]: demos/README.md
[imageBase]: docs/base.jpg
[imageOverlay]: docs/overlay.jpg
[kubernetes style]: docs/glossary.md#kubernetes-style-object
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

## Installation

This assumes [Go](https://golang.org/) (v1.10.1 or higher)
is installed and your `PATH` contains `$GOPATH/bin`:

<!-- @installkustomize @test -->
```
go get k8s.io/kubectl/cmd/kustomize
```

## Usage

#### 1) Make a base

A [base] configuration is a [kustomization] file listing a set of
k8s [resources] - deployments, services, configmaps,
secrets that serve some common purpose.

![base image][imageBase]

#### 2) Customize it with overlays

An [overlay] customizes your base along different dimensions
for different purposes or different teams, e.g. for
_development, staging and production_.

![overlay image][imageOverlay]

#### 3) Run kustomize

Run `kustomize` on your overlay.  The result
is printed to `stdout` as a set of complete
resources, ready to be [applied] to a cluster.
See the [demos].


## About

This project sponsored by [sig-cli].
