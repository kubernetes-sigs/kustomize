[KEP]: https://github.com/kubernetes/community/blob/master/keps/sig-cli/0008-kustomize.md
[`make`]: https://www.gnu.org/software/make
[`sed`]: https://www.gnu.org/software/sed
[applied]: docs/glossary.md#apply
[base]: docs/glossary.md#base
[declarative configuration]: docs/glossary.md#declarative-application-management
[demo]: demos/README.md
[demos]: demos/README.md
[imageBase]: docs/base.jpg
[imageOverlay]: docs/overlay.jpg
[installation]: INSTALL.md
[kubernetes style]: docs/glossary.md#kubernetes-style-object
[kustomization]: docs/glossary.md#kustomization
[overlay]: docs/glossary.md#overlay
[overlays]: docs/glossary.md#overlay
[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[resource]: docs/glossary.md#resource
[resources]: docs/glossary.md#resource
[sig-cli]: https://github.com/kubernetes/community/blob/master/sig-cli/README.md
[variant]: docs/glossary.md#variant
[variants]: docs/glossary.md#variant
[workflows]: docs/workflows.md

# kustomize

`kustomize` lets you customize raw, template-free YAML
files for multiple purposes, leaving the original YAML
untouched and usable as is.

`kustomize` targets kubernetes; it understands and can
patch [kubernetes style] API objects.  It's like
[`make`], in that what it does is declared in a file,
and it's like [`sed`], in that it emits editted text.

[![Build Status](https://travis-ci.org/kubernetes-sigs/kustomize.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/kustomize)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-sigs/kustomize)](https://goreportcard.com/report/github.com/kubernetes-sigs/kustomize)

**Installation**:
Download a binary from the [release page], or
see these [installation] alternatives.

Be sure to try one of the tested [demos].

## Usage


### Make a [base]

In some directory containing your YAML [resource]
files (deployments, services, configmaps, etc.), create a
[kustomization] file.

This file should declare those resources, and any
common customization to apply to them, e.g. _add a
common label_.

![base image][imageBase]

File structure:

> ```
> ~/yourApp
> └── base
>     ├── deployment.yaml
>     ├── kustomization.yaml
>     └── service.yaml
> ```

This is your [base].  The resources in it could be a
fork of someone else's configuration.  If so, you can
easily rebase from the source material to capture
improvements, because you don't modify the resources
directly.

Generate customized YAML with:

```
kustomize build ~/yourApp/base
```

The YAML can be directly [applied] to a cluster:

> ```
> kustomize build ~/yourApp/base | kubectl apply -f -
> ```


###  Create [variants] of a common base using [overlays]

Manage traditional [variants] of a configuration like
_development_, _staging_ and _production_ using
[overlays].

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

Store your overlays in your own repository.  On disk,
the overlay can reference a base in a sibling
directory.  This avoids trouble with nesting git
repositories.

Generate YAML with

```
kustomize build ~/yourApp/overlays/production
```

The YAML can be directly [applied] to a cluster:

> ```
> kustomize build ~/yourApp/overlays/production | kubectl apply -f -
> ```

## About

This tool is sponsored by [sig-cli] ([KEP]).
