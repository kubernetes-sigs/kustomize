---
title: "Overview"
linkTitle: "Overview"
weight: 1
description: >
  Introduction to Kustomize.
---

Kustomize provides a solution for customizing Kubernetes resource configuration free from templates and DSLs.

Kustomize lets you customize raw, template-free YAML files for multiple purposes, leaving the original YAML untouched and usable as is.

Kustomize targets kubernetes; it understands and can patch kubernetes style API objects. It’s like make, in that what it does is declared in a file, and it’s like sed, in that it emits edited text.

## Usage

### 1) Make a `kustomization` file

In some directory containing your YAML `resource`
files (deployments, services, configmaps, etc.), create a
`kustomization` file.

This file should declare those resources, and any
customization to apply to them, e.g. _add a common
label_.

File structure:

 ```
 ~/someApp
 ├── deployment.yaml
 ├── kustomization.yaml
 └── service.yaml
 ```

The resources in this directory could be a fork of
someone else's configuration.  If so, you can easily
rebase from the source material to capture
improvements, because you don't modify the resources
directly.

Generate customized YAML with:

```
kustomize build ~/someApp
```

The YAML can be directly `applied` to a cluster:

 ```
 kustomize build ~/someApp | kubectl apply -f -
 ```


### 2) Create `variants` using `overlays`

Manage traditional `variants` of a configuration - like
_development_, _staging_ and _production_ - using
`overlays` that modify a common `base`.

File structure:
 ```
 ~/someApp
 ├── base
 │   ├── deployment.yaml
 │   ├── kustomization.yaml
 │   └── service.yaml
 └── overlays
     ├── development
     │   ├── cpu_count.yaml
     │   ├── kustomization.yaml
     │   └── replica_count.yaml
     └── production
         ├── cpu_count.yaml
         ├── kustomization.yaml
         └── replica_count.yaml
 ```

Take the work from step (1) above, move it into a
`someApp` subdirectory called `base`, then
place overlays in a sibling directory.

An overlay is just another kustomization, referring to
the base, and referring to patches to apply to that
base.

This arrangement makes it easy to manage your
configuration with `git`.  The base could have files
from an upstream repository managed by someone else.
The overlays could be in a repository you own.
Arranging the repo clones as siblings on disk avoids
the need for git submodules (though that works fine, if
you are a submodule fan).

Generate YAML with

```sh
kustomize build ~/someApp/overlays/production
```

The YAML can be directly `applied` to a cluster:

 ```sh
 kustomize build ~/someApp/overlays/production | kubectl apply -f -
 ```

## Where should I go next?

Give your users next steps from the Overview. For example:

* [Getting Started](/docs/getting-started/): Get started with $project
* [Examples](/docs/examples/): Check out some example code!

