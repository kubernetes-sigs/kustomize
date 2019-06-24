# Glossary
[CRD spec]: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/
[CRD]: #custom-resource-definition
[DAM]: #declarative-application-management
[Declarative Application Management]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/declarative-application-management.md
[JSON]: https://www.json.org/
[JSONPatch]: https://tools.ietf.org/html/rfc6902
[JSONMergePatch]: https://tools.ietf.org/html/rfc7386
[Resource]: #resource
[YAML]: http://www.yaml.org/start.html
[application]: #application
[apply]: #apply
[apt]: https://en.wikipedia.org/wiki/APT_(Debian)
[base]: #base
[bases]: #base
[bespoke]: #bespoke-configuration
[gitops]: #gitops
[k8s]: #kubernetes
[kubernetes]: #kubernetes
[kustomize]: #kustomize
[kustomization]: #kustomization
[kustomizations]: #kustomization
[off-the-shelf]: #off-the-shelf-configuration
[overlay]: #overlay
[overlays]: #overlay
[patch]: #patch
[patches]: #patch
[patchJson6902]: #patchjson6902
[patchExampleJson6902]: https://github.com/kubernetes-sigs/kustomize/blob/master/examples/jsonpatch.md
[patchesJson6902]: #patchjson6902
[proposal]: https://github.com/kubernetes/community/pull/1629
[rebase]: https://git-scm.com/docs/git-rebase
[resource]: #resource
[resources]: #resource
[root]: #kustomization-root
[rpm]: https://en.wikipedia.org/wiki/Rpm_(software)
[strategic-merge]: https://git.k8s.io/community/contributors/devel/sig-api-machinery/strategic-merge-patch.md
[target]: #target
[variant]: #variant
[variants]: #variant
[workflow]: workflows.md

## application

An _application_ is a group of k8s resources related by
some common purpose, e.g.  a load balancer in front of a
webserver backed by a database.
[Resource] labelling, naming and metadata schemes have
historically served to group resources together for
collective operations like _list_ and _remove_.

This [proposal] describes a new k8s resource called
_application_ to more formally describe this idea and
provide support for application-level operations and
dashboards.

[kustomize] configures k8s resources, and the proposed
application resource is just another resource.


## apply

The verb _apply_ in the context of k8s refers to a
kubectl command and an in-progress [API
endpoint](https://goo.gl/UbCRuf) for mutating a
cluster.

One _applies_ a statement of what one wants to a
cluster in the form of a complete resource list.

The cluster merges this with the previously applied
state and the actual state to arrive at a new desired
state, which the cluster's reconcilation loop attempts
to create.  This is the foundation of level-based state
management in k8s.

## base

A _base_ is a [kustomization] referred to
by some other [kustomization].

Any kustomization, including an [overlay], can be a base to
another kustomization.

A base has no knowledge of the overlays that refer to it.

For simple [gitops] management, a base configuration
could be the _sole content of a git repository
dedicated to that purpose_.  Same with [overlays].
Changes in a repo could generate a build, test and
deploy cycle.


## bespoke configuration

A _bespoke_ configuration is a [kustomization] and some
[resources] created and maintained internally by some
organization for their own purposes.

The [workflow] associated with a _bespoke_ config is
simpler than the workflow associated with an
[off-the-shelf] config, because there's no notion of
periodically capturing someone else's upgrades to the
[off-the-shelf] config.

## custom resource definition

One can extend the k8s API by making a
Custom Resource Definition ([CRD spec]).

This defines a custom [resource] (CD), an entirely
new resource that can be used alongside _native_
resources like ConfigMaps, Deployments, etc.

Kustomize can customize a CD, but to do so
kustomize must also be given the corresponding CRD
so that it can interpret the structure correctly.

## declarative application management

Kustomize aspires to support [Declarative Application Management],
a set of best practices around managing k8s clusters.

In brief, kustomize should

 * Work with any configuration, be it bespoke,
   off-the-shelf, stateless, stateful, etc.
 * Support common customizations, and creation of
   [variants] (e.g. _development_ vs.
   _staging_ vs. _production_).
 * Expose and teach native k8s APIs, rather than
   hide them.
 * Add no friction to version control integration to
   support reviews and audit trails.
 * Compose with other tools in a unix sense.
 * Eschew crossing the line into templating, domain
   specific languages, etc., frustrating the other
   goals.

## generator

A generator makes resources that can be used as is,
or fed into a [transformer].


## gitops

Devops or CICD workflows that use a git repository as a
single source of truth and take action (e.g., build,
test or deploy) when that truth changes.

## kustomization

The term _kustomization_ refers to a
`kustomization.yaml` file, or more generally to a
directory (the [root]) containing the
`kustomization.yaml` file and all the relative file
paths that it immediately references (all the local
data that doesn't require a URL specification).

I.e. if someone gives you a _kustomization_ for use
with [kustomize], it could be in the form of

 * one file called `kustomization.yaml`,
 * a tarball (containing that YAML file plus what it references),
 * a git archive (ditto),
 * a URL to a git repo (ditto), etc.

A kustomization file contains [fields](fields.md)
falling into four categories:

 * _resources_ - what existing [resources] are to be customized.
   Example fields: _resources_, _crds_.

 * _generators_ - what _new_ resources should be created.
   Example fields: _configMapGenerator_ (legacy),
   _secretGenerator_ (legacy), _generators_ (v2.1).

 * _transformers_ - what to _do_ to the aforementioned resources.
   Example fields: _namePrefix_, _nameSuffix_, _images_,
   _commonLabels_, _patchesJson6902_, etc. and the more
   general _transformers_ (v2.1) field.

 * _meta_ - fields which may influence all or some of
   the above.  Example fields: _vars_, _namespace_,
   _apiVersion_, _kind_, etc.


## kustomization root

The directory that immediately contains a
`kustomization.yaml` file.

When a kustomization file is processed, it may or may
not be able to access files outside its root.

Data files like resource YAML files, or text files
containing _name=value_ pairs intended for a ConfigMap
or Secret, or files representing a patch to be used in
a patch transformation, must live _within or below_ the
root, and as such are specified via _relative
paths_ exclusively.

A special flag (in v2.1), `--load_restrictions none`,
is provided to relax this security feature, to, say,
allow a patch file to be shared by more than one
kustomization.

Other kustomizations (other directories containing a
`kustomization.yaml` file) may be referred to by URL, by
absolute path, or by relative path.

If kustomization __A__ depends on kustomization __B__, then

 * __B__ may not _contain_ __A__.
 * __B__ may not _depend on_ __A__, even transitively.

__A__ may contain __B__, but in this case it might be
simplest to have __A__ directly depend on __B__'s
resources and eliminate __B__'s kustomization.yaml file
(i.e. absorb __B__ into __A__).

Conventionally, __B__ is in a directory that's sibling
to __A__, or __B__ is off in a completely independent
git repository, referencable from any kustomization.


A common layout is

> ```
> ├── base
> │   ├── deployment.yaml
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── dev
>     │   ├── kustomization.yaml
>     │   └── patch.yaml
>     ├── prod
>     │   ├── kustomization.yaml
>     │   └── patch.yaml
>     └── staging
>         ├── kustomization.yaml
>         └── patch.yaml
> ```

The three roots `dev`, `prod` and `staging`
(presumably) all refer to the `base` root.  One would
have to inspect the `kustomization.yaml` files to be
sure.

## kubernetes

[Kubernetes](https://kubernetes.io) is an open-source
system for automating deployment, scaling, and
management of containerized applications.

It's often abbreviated as _k8s_.

## kubernetes-style object

[fields required]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

An object, expressed in a YAML or JSON file, with the
[fields required] by kubernetes.  Basically just a
_kind_ field to identify the type, a _metadata/name_
field to identify the particular instance, and an
_apiVersion_ field to identify the version (if there's
more than one version).

## kustomize

_kustomize_ is a command line tool supporting
template-free, structured customization of declarative
configuration targetted to k8s-style objects.

_Targetted to k8s means_ that kustomize has some
understanding of API resources, k8s concepts like
names, labels, namespaces, etc. and the semantics of
resource patching.

kustomize is an implementation of [DAM].


## off-the-shelf configuration

An _off-the-shelf_ configuration is a kustomization and
resources intentionally published somewhere for others
to use.

E.g. one might create a github repository like this:

> ```
> github.com/username/someapp/
>   kustomization.yaml
>   deployment.yaml
>   configmap.yaml
>   README.md
> ```

Someone could then _fork_ this repo (on github) and
_clone_ their fork to their local disk for
customization.

This clone could act as a [base] for the user's
own [overlays] to do further customization.

## overlay

An _overlay_ is a kustomization that depends on
another kustomization.

The [kustomizations] an overlay refers to (via file
path, URI or other method) are called [bases].

An overlay is unusable without its bases.

An overlay may act as a base to another overlay.

Overlays make the most sense when there is _more than
one_, because they create different [variants] of a
common base - e.g.  _development_, _QA_, _staging_ and
_production_ environment variants.

These variants use the same overall resources, and vary
in relatively simple ways, e.g. the number of replicas
in a deployment, the CPU to a particular pod, the data
source used in a ConfigMap, etc.

One configures a cluster like this:

> ```
>  kustomize build someapp/overlays/staging |\
>      kubectl apply -f -
>
>  kustomize build someapp/overlays/production |\
>      kubectl apply -f -
> ```

Usage of the base is implicit - the overlay's
kustomization points to the base.

See also [root].

## package

The word _package_ has no meaning in kustomize, as
kustomize is not to be confused with a package
management tool in the tradition of, say, [apt] or
[rpm].

## patch

General instructions to modify a resource.

There are two alternative techniques with similar
power but different notation - the
[strategic merge patch](#patchstrategicmerge)
and the [JSON patch](#patchjson6902).

## patchStrategicMerge

A _patchStrategicMerge_ is [strategic-merge]-style patch (SMP).

An SMP looks like an incomplete YAML specification of
a k8s resource.  The SMP includes `TypeMeta`
fields to establish the group/version/kind/name of the
[resource] to patch, then just enough remaining fields
to step into a nested structure to specify a new field
value, e.g. an image tag.

By default, an SMP _replaces_ values.  This
usually desired when the target value is a simple
string, but may not be desired when the target
value is a list.

To change this
default behavior, add a _directive_.  Recognized
directives include _replace_ (the default), _merge_
(avoid replacing a list), _delete_ and a few more
(see [these notes][strategic-merge]).

Note that for custom resources, SMPs are treated as
[json merge patches][JSONMergePatch].

Fun fact - any resource file can be used as
an SMP, overwriting matching fields in another
resource with the same group/version/kind/name,
but leaving all other fields as they were.

TODO(monopole): add ptr to example.

## patchJson6902

A _patchJson6902_ refers to a kubernetes [resource] and
a [JSONPatch] specifying how to change the resource.

A _patchJson6902_ can do almost everything a
_patchStrategicMerge_ can do, but with a briefer
syntax.  See this [example][patchExampleJson6902].

## plugin

A chunk of code used by kustomize, but not necessarily
compiled into kustomize, to generate and/or transform a
kubernetes resource as part of a kustomization.

Details [here](plugins).

## resource

A _resource_ in the context of a REST-ful API is the
target object of an HTTP operation like _GET_, _PUT_ or
_POST_.  k8s offers a REST-ful API surface to interact
with clients.

A _resource_, in the context of a kustomization, is a
[root] relative path to a [YAML] or [JSON] file
describing a k8s API object, like a Deployment or a
ConfigMap, or it's a path to a kustomization, or a URL
that resolves to a kustomization.

More generally, a resource can be any correct YAML file
that [defines an object](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields)
with a _kind_ and a _metadata/name_ field.

## root

See [kustomization root][root].

## sub-target / sub-application / sub-package

A _sub-whatever_ is not a thing. There are only
[bases] and [overlays].

## target

The _target_ is the argument to `kustomize build`, e.g.:

> ```
>  kustomize build $target
> ```

`$target` must be a path or a url to a [kustomization].

The target contains, or refers to, all the information
needed to create customized resources to send to the
[apply] operation.

A target can be a [base] or an [overlay].

## transformer

A transformer can modify a resource, or merely
visit it and collect information about it in the
course of a `kustomize build`.

## variant

A _variant_ is the outcome, in a cluster, of applying
an [overlay] to a [base].

E.g., a _staging_ and _production_ overlay both modify
some common base to create distinct variants.

The _staging_ variant is the set of resources exposed
to quality assurance testing, or to some external users
who'd like to see what the next version of production
will look like.

The _production_ variant is the set of resources
exposed to production traffic, and thus may employ
deployments with a large number of replicas and higher
cpu and memory requests.
