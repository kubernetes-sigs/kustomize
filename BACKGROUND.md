[Kubernetes objects]: https://kubernetes.io/docs/reference/using-api/api-concepts
[Kubernetes object]: https://kubernetes.io/docs/reference/using-api/api-concepts
[`Kustomization` object]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization

# Architecture

## Background Context

### Kubernetes

A Kubernetes cluster has an _API server_ that accepts [Kubernetes
objects] from the outside world then stores them in a
database.  At any given moment these objects, in toto, are a
specification of the desired state of the cluster.  Controllers in the
cluster read this specification, and try to push the cluster to a
state matching the spec.

A Kubernetes user prepares Kubernetes objects as YAML text files, then
posts them to the API server using some tool. Typically that tool is
the `kubectl` program; specifically the `kubectl apply` command.

The `apply` command uses a local configuration file (usually
`$HOME/.kube/config`) to facilitate finding and authenticating to the
API server, then composes and posts the proper series of requests to
API endpoints to send the resources.


### The use of kustomize with Kubernetes

`kustomize` fills a functional slot between those kubernetes objects
in YAML text files and the API server in the cloud.

kustomize provides a means to define and use a particular Kubernetes
object - a `Kustomization` - that in turn refers to some set of
Kubernetes object files, and _changes_ to make to that set.

The basic usage is:

```
kustomize build ${dir} | kubectl apply -f -
```

where `${dir}` is a directory containing a `kustomization.yaml` file.
This file contains one instance of a [`Kustomization` object], which
in turn specifies the full configuration via various fields.

In practice, the pipe (`|`) is usually replaced with a review stage.
More on this below.

#### kustomize is a stream emitter

kustomize is similar to `sed` in that it _emits_ a stream of objects
that are either edited from some source material, or generated on the
fly.

Unlike `sed`, kustomize doesn't read `stdin`.

> kustomize is meant to be used as a step in a production deployment
> system where all data moves on a verifiable path from a version control
> commit to production.

Because of this, kustomize output should not change based on
ephemeral command line flags, environment variables, or input streams.
kustomize output should only change in response to changes in a file
system.

To write the stream to a file just use shell redirection.  To split
the output into different files, use:

```
kustomize build --output ${outDir} ${dir}
```

The `${outDir}` must be a directory outside the input `${dir}`.  Each
object in the stream will be written to its own file in `${outDir}`
directory, all at the same level, with generated file names.

#### kustomize is a variant builder

A typical use case is to have a large set of Kubernetes objects in
various files that make up a _base_ cluster configuration, then use
kustomize to create _variants_ of that base for application to a
cluster.

Such variants could be a frequently changing _development_ variant, a
production-candidate _staging_ variant, and a relatively stable and
expensive (resource intensive) _production_ variant.

The difference between a variant and its base (or bases) are described
in the `kustomization.yaml` file.


#### kustomize edits the `kustomization.yaml` file

An important kustomize use case is working with objects that live in
unedited forks of upstream git repositories.  If the material is
unedited, one can perform conflict-free rebases at will.

So, with one exception, kustomize doesn't want to edit Kubernetes
objects on disk. The exception is that kustomize has commands that
create and edit a `kustomization.yaml` file in-situ, much as the `go`
tool has commands to edit `go.mod` files.

##### The `Kustomization` object

When reading `kustomization.yaml`, kustomize expects to find one
instance of a [`Kustomization` object].

kustomize doesn't require this object to have the [Kubernetes object]
fields `apiVersion`, `kind` and `metadata`, but it will validate and
honor them if they are present, and if one uses kustomize to edit the
`kustomization.yaml` file, kustomize will add them if they are
missing.  These special Kubernetes object fields are the only fields
in a kustomization file that can be said to have _default values_ if
missing.

The remaining `Kustomization` fields cannot have default values,
because by definition they are user-specific customization
instructions applied to other files.

#### DRY vs. hydrated

Cluster operators want a reliable record of the configuration that the
cluster is supposed to track, and a trivial means to _roll back_ (in
a version control sense) to a previous configuration.

This implies the existence of a repository of fully _hydrated_ cluster
configuration.

> The adjective _hydrated_ in this context is used as an antonym to
> _dry_.  The word _dry_ comes from the phrase _Don't Repeat
> Yourself_.  A _dry configuration_ is in some normalized state that
> eliminates repeated information for brevity and duplication-error
> avoidance, the cost being that the configuration is typically no
> longer directly usable without processing.

In the kustomize approach, the dry material lives in one or more bases
that can be used as-is without involving kustomize.

A `Kustomization` object refers to the dry config, and is used to
create a stream of hydrated configuration.  The Kustomization object
is absent from this stream.

The stream can be channeled into a proposed change to a repository
storing hydrated configuration.  If the change is approved and
committed, the commit can trigger some process that pulls objects and
applies them to a cluster.
