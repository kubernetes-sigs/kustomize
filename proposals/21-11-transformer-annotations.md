<!--
**Note:** When your proposal is complete, all of these comment blocks should be removed.

To get started with this template:

- [ ] **Make a copy of this file.**
  Name it `YY-MM-short-descriptive-title.md` (where `YY-MM` is the current year and month).
- [ ] **Fill out this file as best you can.**
  At minimum, you should fill in the "Summary" and "Motivation" sections.
- [ ] **Create a PR.**
  Ping `@kubernetes-sigs/kustomize-admins` and `@kubernetes-sigs/kustomize-maintainers`.
-->

# Option for origin and transformer annotations

**Authors**:
- natasha41575

**Reviewers**: 
- monopole
- KnVerey
- yuwenma

**Status**: implementable

## Summary

This proposal is to extend the `buildMetadata.originAnnotations` option in the kustomization to annotate
generated resources with information about the source of their generators, and to add a new option 
`buildMetadata.transformerAnnotations` that will add annotations to each resource with information about
all the transformers that have processed it. These options will help users understand how their output was
created and debug it accordingly. 

## Motivation

When a user is managing a large number of resources with kustomize, the output of `kustomize build` will be a 
long stream of yaml resources, and it can be extremely tedious for the user to understand how each resource 
was created and processed. However, we can provide the user with an option to annotate each resource with 
origin and transformation data. This information will help kustomize users
debug their large kustomize configurations. Because these annotations can be used for debugging purposes, we 
would like them to be verbose and communicate as much information as possible to the user.

Kustomize previously had a [feature request](https://github.com/kubernetes-sigs/kustomize/issues/3979) to add an 
option to kustomize build to retain the origin of a resource in an annotation, and a final design was settled 
on after some discussion. It was implemented by [this pull request](https://github.com/kubernetes-sigs/kustomize/pull/4065) 
and released in kustomize v4.3.0. This feature added a new field to the kustomization file, `buildMetadata`, an 
array of boolean options such as `originAnnotations`. If `originAnnotations` is set, `kustomize build` will add annotations 
to each resource describing the resource's origin. 

While the above feature gives us valuable information about where many resources originated from, there are two 
pieces missing:

- Origin annotations for generated resources. Resources can either be generated using a builtin generator such as 
`ConfigMapGenerator` or `SecretGenerator`, or can be generated from a custom generator invoked from either the `generators`
or `transformers` field. 

- Everything in kustomize is a transformer, so we would like to have annotations about which transformers, builtin or 
custom, touched each resource. 

For a user attempting to debug a large stream of output from `kustomize build`, knowing how each resource has been 
transformed through various layers will make it much easier to understand how the output was rendered, easing debugging 
and making changes to kustomization files.

### Builtin Transformers

Builtin transformers include those that are invoked through various kustomization 
fields. Some of these transformers are: 
  - AnnotationsTransformer
  - HashTransformer
  - ImageTagTransformer
  - LabelTransformer
  - LegacyOrderTransformer
  - NamespaceTransformer
  - PatchJson6902Transformer
  - PatchStrategicMergeTransformer
  - PatchTransformer
  - PrefixSuffixTransformer
  - ReplacementTransformer
  - ReplicaCountTransformer 
  - ValueAddTransformer
  
Custom transformers are invoked via the `transformers` field in the kustomization file. Although KRM-style plugins can be 
invoked via the `generators` and `validators` fields, it is impossible for plugins in these fields to transform existing resources,
as `generators` do not take input and `validators` are not permitted to change input. Therefore, all custom transformers will 
appear in the `transformers` field. 

It is, however, possible to put a generator or validator in the `transformers` field, so we should try to differentiate 
the plugins in the `transformers` field where we can, and only store data about actual transformers. 

**Goals:**
1. When `originAnnotations` is set, add annotations to generated resources that describe the source of the generator that 
created them.
2. Add a new option, `transformerAnnotations`, to `buildMetadata` that, when set, will annotate each resource with 
information about which transformers, builtin or custom, touched each resource.

**Non-goals:**
1. Change the syntax of the existing `originAnnotations` option.
2. Add these annotations to the output of kustomize by default.

## Proposal

### Origin Annotation
To retain the information about the origin of a resource, the user can specify the `originAnnotations` 
option in the `buildMetadata` field of the kustomization: 

```yaml
buildMetadata: [originAnnotations]
```

When this option is set, generated resources should receive an annotation with key `config.kubernetes.io/origin`,
containing data about the generator that produced it. If the resource is from the `resources` field, this annotation
contains data about the file it originated from. 

The possible fields of these annotations are:

```yaml
config.kubernetes.io/origin: |
  path: path.yaml # The path to the resource file itself
  ref: v0.0.0 # If from a remote file or generator, the ref of the repo URL
  repo: http://github.com/examplerepo # If from a remote file or generator, the repo source
  configuredIn: path/to/generatorconfig # If a generated resource, the path to the generator config
  configuredBy: # If a generated resource, the ObjectReference of the generator config
    kind: Generator
    apiVersion: builtin 
    name: foo
    namespace: default
```

All local file paths are relative to the top-level kustomization, i.e. the kustomization file in the directory upon 
which `kustomize build` was invoked. For example, if someone were to run `kustomize build foo`, all file paths 
in the annotation output would be relative to`foo/kustomization.yaml`. 

All remote file paths are relative to the root of the remote repository. 

Any fields that are not applicable would be omitted from the final output. If a generator is invoked via 
a field in the kustomization file, `configuredIn` would point to the kustomization file itself. 

#### Examples

##### Resource declared from `resources`
A kustomization such as the following:

```yaml
resources:
- deployment.yaml

buildMetadata: [originAnnotations]
```

would produce a resource with an annotation like the following:

```yaml
config.kubernetes.io/origin: |
  path: deployment.yaml
```

##### Local custom generator
A kustomization such as the following:

```yaml
generators:
- generator.yaml

buildMetadata: [originAnnotations]
```

would produce a resource with an annotation like the following:

```yaml
config.kubernetes.io/origin: |
  configuredIn: generator.yaml
  configuredBy: 
    kind: MyGenerator
    apiVersion: v1
    name: generator
```

##### Remote builtin generator
We have a top-level kustomization such as the following:

```yaml
resources:
- github.com/examplerepo/?ref=v1.0.6

buildMetadata: [originAnnotations]
```

which uses `github.com/examplerepo/?ref=v1.0.6` as a remote base. This remote base has the following kustomization
defined in `github.com/examplerepo/kustomization.yaml`:

```yaml
configMapGenerator:
- name: my-java-server-env-vars
  literals:
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
```

Running `kustomize build` on the top-level kustomization would produce the following output:

```yaml
apiVersion: v1
data:
  JAVA_HOME: /opt/java/jdk
  JAVA_TOOL_OPTIONS: -agentlib:hprof
kind: ConfigMap
metadata:
  name: my-java-server-env-vars-44k658k8gk
  annotations:
    config.kubernetes.io/origin: |
      ref: v1.0.6 
      repo: github.com/examplerepo 
      configuredIn: kustomization.yaml
      configuredBy: 
        kind: ConfigMapGenerator
        apiVersion: builtin 
```

### Transformer Annotations

To retain the information about what transformers have acted on each resource, we can propose a new option 
`transformerAnnotations` in the `buildMetadata` field of the kustomization:

```yaml
buildMetadata: [transformerAnnotations]
```

It is possible for the user to set both `transformerAnnotations` and `originAnnotations`:

```yaml
buildMetadata: [originAnnotations, transformerAnnotations]
```

When the `transformerAnnotations` option is set, kustomize will add annotations with information about what transformers 
have acted on each resource. Transformers can be invoked either through various fields in the kustomization file 
(e.g. the `replacements` field will invoke the ReplacementTransformer), or through the `transformers` field.

The annotation key for transformer annotatinos will be `config.kubernetes.io/transformations`, which will contain a list of
transformer data:

```yaml
config.kubernetes.io/transformations: |
- ref: v0.0.0 # If from a remote transformer, the ref of the repo URL
  repo: http://github.com/examplerepo # If from a remote transformer, the repo source
  configuredIn: path/to/transformerconfig # The path to the transformer config
  configuredBy: # The ObjectReference of the transformer config
    kind: Transformer
    apiVersion: builtin 
    name: foo
    namespace: default
```

The possible fields for each item in the list is identical to the possible fields in `config.kubernetes.io/origin`, except that
the transformer annotation does not have a `path` field for the path to the resource file itself. 

All local file paths are relative to the top-level kustomization, i.e. the kustomization file in the directory upon 
which `kustomize build` was invoked. For example, if someone were to run `kustomize build foo`, all file paths 
in the annotation output would be relative to`foo/kustomization.yaml`. 

All remote file paths are relative to the root of the remote repository. 

Any fields that are not applicable would be omitted from the final output. If a transformer is invoked via 
a field in the kustomization file, `configuredIn` would point to the kustomization file itself. 


#### Examples

#### Local custom transformer
A kustomization such as the following:

```yaml
transformers:
- transformer.yaml

buildMetadata: [transformerAnnotations]

```

would produce a resource with an annotation like the following:

```yaml
config.kubernetes.io/transformations: |
- configuredIn: transformer.yaml
  configuredBy: 
    kind: MyTransformer
    apiVersion: v1
    name: transformer
```

##### Remote builtin transformer + local builtin transformer
We have a top-level kustomization such as the following:

```yaml
resources:
- github.com/examplerepo/?ref=v1.0.6

buildMetadata: [transformerAnnotations]

namespace: my-ns
```

which uses `github.com/examplerepo/?ref=v1.0.6` as a remote base. This remote base has the following kustomization
defined in `github.com/examplerepo/kustomization.yaml`:

```yaml
resources:
- deployment.yaml

namePrefix: pre-
```

`deployment.yaml` contains the following:

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: deploy
```

Running `kustomize build` on the top-level kustomization would produce the following output:

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: pre-deploy
  annotations:
    config.kubernetes.io/transformations: |
    - ref: v1.0.6 
      repo: github.com/examplerepo 
      configuredIn: kustomization.yaml
      configuredBy: 
        kind: PrefixSuffixTransformer
        apiVersion: builtin 
    - configuredIn: kustomization.yaml
      configuredBy:
        kind: NamespaceTransformer
        apiVersion: builtin
```

### Kustomize edit

We will want to provide convenient commands for users to edit the `buildMetadata` field in their kustomization.
We should support the following commands:

`kustomize edit add buildMetadata originAnnotations`

`kustomize edit add buildMetadata transformerAnnotations`

`kustomize edit add buildMetadata *`

`kustomize edit remove buildMetadata originAnnotaitons`

`kustomize edit remove buildMetadata transformerAnnotations`

`kustomize edit remove buildMetadata *`

The wildcard match `*` should add/remove all possible options for `buildMetadata`.

### Implementation

If there is a large tree of kustomization files, the data stored for origin and transformer data could be potentially
very large, so we should be careful about implementation. We should:

1. Ensure that we are only accumulating the data if the option is set.
2. Account for `generators` that may appear in the `transformers` field. When we are processing each plugin invoked
in the `transformers` field, we should check to see if the resources have an `origin` annotation. If not, we can assume that
it was generated by the transformer and we can add the corresponding `origin` annotation. If it already has an `origin`
annotation, we can add the corresponding `transformation` annotation. 
3. A `ConfigMapGenerator` with `merge` set to true actually transforms instead of generates, so we should take care to 
account for this special case. In this case, we should add a transformer annotation. 
 

### User Stories

#### Story 1

I am running kustomize build on some configuration that has already been created. This configuration has a large number
of resources and patches through multiple overlays in different directories. After running `kustomize build`, I get a 
stream of 100+ resources, but I notice that a few of them have a typo, likely originating from the same base or patch 
file. However, because my directory structure is large and complex, it is difficult to narrow down which base these
resources originated from and which patches have been applied to it. In the top-level kustomization file, I can add a
new `buildMetadata` field:

```yaml
buildMetadata: [originAnnotations, transformerAnnotations]
```

The output of `kustomize build` will contain patched resources with annotations similar to the following:
```
annotations:
  config.kubernetes.io/origin: |
    path: deployment.yaml
  config.kubernetes.io/transformations:
  - configuredIn: ../base/kustomization.yaml
    configuredBy: 
      kind: PatchTransformer
      apiVersion: builtin 
  - configuredIn: ../dev/kustomization.yaml
    configuredBy: 
      kind: PatchStrategicMergeTransformer
      apiVersion: builtin 
  - configuredIn: ../dev/kustomization.yaml
    configuredBy: 
      kind: PatchesJson6902Transformer
      apiVersion: builtin 
```

From these annotations, I can narrow down the set of files that contributed to the resource's final output and
debug it accordingly. 

#### Story 2

I am using a tool that automatically runs `kustomize build` and deploys the output to the cluster, such as `skaffold deploy`,
`kubectl apply -k`, or Google's Config Sync. I am unfamiliar with kustomize CLI and do not wish to install it to my
local machine or learn how to use it beyond using tools that run `kustomize build`. I add the `originAnnotations` and 
`transformationAnnotations` options to my top-level kustomization file. 

With one of these tools, I apply my kustomize configuration to my cluster, but later discover that something in my cluster 
is not configured correctly. Because the annotations have persisted to the cluster, I can take a look at the resources in
my cluster and understand how each of them was created, helping me figure out what went wrong. 


### Risks and Mitigations
N/A

### Dependencies
N/A

### Scalability
With a large input, the data collected here could be very large, so we should take care to only accumulate data when necessary.

## Drawbacks
N/A

## Alternatives
1. Putting this data in the comments of the rendered resources, but this would be a large project because 
kustomize doesn't currently support comments. 

2. Making this a flag to `kustomize build` rather than additional field, but we would like to align 
with the kustomize way of avoiding build-time side effects and have everything declared explicitly in the kustomization.

3. Having a separate `kustomize` command that prints out only the origin and/or transformer data. While this could be very
useful as a debugging feature, there are use cases for tools that automatically run kustomize and deploy, where it would be more
useful to the user to have the annotations persist to the cluster. See User Story 2. 

## Rollout Plan
This is a fairly lightweight, minor and optional feature that can be implemented and
released without needing a rollout plan, much like the `originAnnotations` option that
we are extending.  

The `origin` annotation is simply being extended, so there is no need to mark this feature as alpha. The rollout plan
for the `transformation` annotation is outlined below:

### Alpha
- Will the feature be gated by an "alpha" flag? Which one?

The feature will not be gated by an alpha flag, but we will start the annotation with the prefix "alpha", so the
annotation will key be `alpha.config.kubernetes.io/transformations`

- Will the feature be available in `kubectl kustomize` during alpha? Why or why not?

Yes, the feature will be part of `kubectl kustomize` as an option in the `buildMetadata` field, and will likewise
start with the annotation key `alpha.config.kubernetes.io/transformations`

### GA
We will wait for two `kubectl` releases before promoting the feature. We will iterate the feature based on feedback from 
users of Config Sync, Skaffold, kustomize CLI, and kustomize-in-kubectl prior to promotion. 
