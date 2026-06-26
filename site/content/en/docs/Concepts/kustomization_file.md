---
title: "Kustomizaton File"
linkTitle: "Kustomizaton File"
weight: 10
description: >
  What is the Kustomizaton file? 
---

[KRM]: https://github.com/kubernetes/design-proposals-archive/blob/main/architecture/resource-management.md

The kustomization file is a YAML specification of a Kubernetes
Resource Model ([KRM]) object called a _Kustomization_.
A kustomization describes how to generate or transform
other KRM objects.

Although most practical kustomization files don't actually look this
way, a `kustomization.yaml` file is basically four lists:

```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- {pathOrUrl}
- ...

generators:
- {pathOrUrl}
- ...

transformers:
- {pathOrUrl}
- ...

validators:
- {pathOrUrl}
- ...
```

The order in each of these lists is relevant
and respected.

> There are other fields too, e.g. `commonLabels`, `namePrefixes`,
> `patches`, etc.  These fields are _convenience_ fields, shorthand for
> longer transformer configuration stanzas, and are discussed later.
> They're what's used most often, but it's useful to first cover
> the fundamentals before discussing the conveniences.

In all cases the `{pathOrUrl}` list entry can specify

 - a file system path to a YAML _file_ containing one or
   more KRM objects, or
 - a _directory_ (local or in a remote git repo)
   that contains a `kustomization.yaml` file.

In the latter case, the kustomization is recursively built (aka
_hydrated_) into a flat list of KRM objects that's effectively
injected into the encapsulating list in order.  When this happens, the
encapsulating kustomization can be called an _overlay_, and what it
refers to can be called a _base_.

A typical layout:

```
app1/
  kustomization.yaml
    | resources:
    | - ../base
    | patches:
    | - patch1.yaml
  patch1.yaml

app2/
  kustomization.yaml
    | resources:
    | - ../base
    | patches:
    | - patch2.yaml
  patch2.yaml

base/
  kustomization.yaml
    | resources:
    | - deployment.yaml
    | - configMap.yaml
  deployment.yaml
  configMap.yaml
```

[mainly useful]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/krusty/inlinetransformer_test.go#L26

Under `resources`, the result of reading KRM yaml files or executing
recursive kustomizations becomes the list of _input objects_ to the
current build stage.

Under `generators`, `transformers` and `validators`, the result of
reading/hydrating is a list of KRM objects that _configure operations_
that kustomize is expected to perform.

> Some of these fields allow YAML inlining, allowing a KRM object to be
> declared directly in the `kustomization.yaml` file (in practice this
> is [mainly useful] in the `transformers` field).

These configurations specify some executable (e.g. a plugin) along
with that executable's _configuration_.  For example, a replica count
transformer's configuration must specify both an executable capable
of parsing and modifying a _Deployment_, and the actual numerical
value (or increment) to use in the Deployment's `replicas` field.

### Ordering

A build stage first processes `resources`, then it processes `generators`,
adding to the resource list under consideration, then it processes
`transformers` to modify the list, and finally runs `validators` to check the
list for whatever error.

### Conveniences

The `resources` field is a convenience.  One can omit `resources`
field and instead use a generator that accepts a file path list,
expanding it as needed.  Such a generator would read the file system,
doing the job that kustomize does when processing the `resources`
field.

All the other fields in a kustomization file (`configMapGenerator`,
`namePrefix`, `patches`, etc.) are conveniences as well, as they are
shorthand for: run a particular generator or transformer with a
particular configuration.

Likewise, a `validator` is just a transformer that doesn't transform,
but can (just like a transformer) _fail the build_ with an error
message.  Coding up a validator is identical to coding up a
transformer. The only difference is in how it's used by kustomize;
kustomize attempts to disallow validators from making changes.

The next section explains why the `generators` field is also just a
convenience.

### Generators and Transformers

In the code, the interfaces distinguishing a generator from a
transformer are:

```
// Generator creates an instance of ResMap.
type Generator interface {
  Generate() (ResMap, error)
}

// Transformer can modify an instance of ResMap.
type Transformer interface {
  Transform(m ResMap) error
}
```

In these interfaces, a `ResMap` is a list of kubernetes `Resource`s
with ancillary map-like lookup and modification methods.

A generator cannot be a transformer, because it doesn't accept an
input other than its own configuration.  Configuration for both generators
and transformers are done via a distinct (and common) interface.

A transformer doesn't implement `Generator`, but it's capable of
behaving like one.

This is because `ResMap` has the methods

```
Append(*Resource)
Replace(*Resource)
Remove(ResId)
Clear()
  ...etc.
```
i.e. the ResMap interface allows for growing and shrinking
the Resource list, as well as mutating each Resource on it.

A transformer (specifically the author of a transformer)
can call these methods - creating, sorting, destroying, etc.

[kyaml]: https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/doc.go

> At the time of writing, the ResMap is being converted to a mutable
> list RNodes, objects that integrate KRM with a new
> kubernetes-specific YAML library called [kyaml].  As more programs
> speak kyaml, kustomize's role will evolve too.

Transformers have a general generative power.

A kustomization overlay, could, say, fix common oversights made in
cluster configuration.

For example, a transformer could scan all resources, looking for the
_need_ for a `PodDisruptionBudget`, and _conditionally_ add it
and hook it up as a guard rail for the user.

### Everything is a transformer

Every field in a kustomization file could be expressed as a
transformer, so any kustomization file can be converted to a
kustomization file with one `transformers:` field.

So why keep all these fields?

The fields in kustomization file are useful for ordering and
signalling, e.g. _these particular things are transformers, and should
run after the generators, but before the validators_.

Also, they make common use cases easier to express.

E.g. the following two YAML stanzas do the exactly the same thing if
added to a kustomization file:

```
namePrefix: bob-
```

```
transformers:
- |-
  apiVersion: builtin
  kind: PrefixSuffixTransformer
  metadata:
    name: myFancyNamePrefixer
  prefix: bob-
  fieldSpecs:
  - path: metadata/name
```


### Transformed transformers

The arguments to `resources` are usually files containing instances of
_Deployment_, _Service_, _PodDisruptionBudget_, etc., but they could
also be transformer configurations.

In this case the transformer configurations are just grist for the
kustomization mill, and can be modifed and passed up an overlay stack,
and later be used to as input in a `transformers` field, whereupon
they'll be applied to any resources at that kustomization stage.

For example, the following file layout has two apps using a common
pair of bases.

One base contains a deployment and a configMap.  The other contains
transformer configurations.  This is a means to specify a set of
reusable, custom transformer configs.

In between the apps and these bases are intermediate overlays
that transform the base transformer configurations before they are
used in the top level apps.

> ```
> app1/
>   kustomization.yaml
>     | resources:
>     | - ../base/resources
>     | transformers:
>     | - ../transformers1
>     | patches:
>     | - patch1.yaml
>   patch1.yaml
>     | {a patch for resources}
>
> app2/
>   kustomization.yaml
>     | resources:
>     | - ../base/resources
>     | transformers:
>     | - ../transformers2
>     | patches:
>     | - patch2.yaml
>   patch2.yaml
>     | {some other patch for the resources}
>
> transformers1/
>   kustomization.yaml
>     | resources:
>     | - ../base/transformers
>   transformerPatch1.yaml
>     | {a patch for the base transformer configs}
>
> transformers2/
>   kustomization.yaml
>     | resources:
>     | - ../base/transformers
>   transformerPatch1.yaml
>     | {some other patch for the base transformer configs}
>
> base/
>   transformers/
>     kustomization.yaml
>        | resources:
>        | - transformerConfig1.yaml
>        | - transformerConfig2.yaml
>     transformerConfig1.yaml
>     transformerConfig2.yaml
>   resources/
>     kustomization.yaml
>        | resources:
>        | - deployment.yaml
>        | - configMap.yaml
>     deployment.yaml
>     configMap.yaml
> ```

This isn't a recommended or disallowed practice, but something that's
allowed by how kustomization fields are processed.

[References]: /docs/reference/api/kustomization-file/

For a detailed explanation of all available fields in *Kustomization*, check out [References].
