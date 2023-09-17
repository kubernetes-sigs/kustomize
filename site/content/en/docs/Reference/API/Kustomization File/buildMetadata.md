---
title: "buildMetadata"
linkTitle: "buildMetadata"
type: docs
weight: 2
description: >
    Specify options for including information about the build in annotations or labels. 
---

The `buildMetadata` field is a list of strings. The strings can be one of three builtin
options that add some metadata to each resource about how the resource was built. 

These options are:

- `managedByLabel`
- `originAnnotations`
- `transformerAnnotations`

It is possible to set one or all of these options in the kustomization file:

```yaml
buildMetadata: [managedByLabel, originAnnotations, transformerAnnotations]
```

### Managed By Label
To mark the resource as having been managed by kustomize, you can specify the `managedByLabel`
option in the `buildMetadata` field of the kustomization:

```yaml
buildMetadata: [managedByLabel]
```

This will add the label `app.kubernetes.io/managed-by` to each resource with the version of kustomize 
that has built it. For example, given the following input:

kustomization.yaml
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
buildMetadata: [managedByLabel]
```

service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
```

`kustomize build` will produce a resource with an output like the following:

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/managed-by: kustomize-v4.4.1
  name: myService
spec:
  ports:
  - port: 7002
```


### Origin Annotation
To annotate resources with information about their origin, you can specify the `originAnnotations`: 

```yaml
buildMetadata: [originAnnotations]
```

When this option is set, generated resources will receive an annotation with key `config.kubernetes.io/origin`,
containing data about the generator that produced it. If the resource is from the `resources` field, this annotation
contains data about the file it originated from. 

The possible fields of these annotations are:

- `path`: The path to a resource file itself
- `ref`: If from a remote file or generator, the ref of the repo URL.
- `repo`: If from a remote file or generator, the repo source
- `configuredIn`: If a generated resource, the path to the generator config. If a generator is invoked via a field 
in the kustomization file, this would point to the kustomization file itself. 
- `configuredBy`: If a generated resource, the ObjectReference of the generator config.  


All local file paths are relative to the top-level kustomization, i.e. the kustomization file in the directory upon 
which `kustomize build` was invoked. For example, if someone were to run `kustomize build foo`, all file paths 
in the annotation output would be relative to`foo/kustomization.yaml`. All remote file paths are relative to the root of the 
remote repository. Any fields that are not applicable would be omitted from the final output.  

Here is an example of what this would look like:
```yaml
config.kubernetes.io/origin: |
  path: path.yaml 
  ref: v0.0.0 
  repo: http://github.com/examplerepo 
  configuredIn: path/to/generatorconfig 
  configuredBy: 
    kind: Generator
    apiVersion: builtin 
    name: foo
    namespace: default
```

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

### Transformer Annotations [Alpha]

To annotate resources with information about the transformers that have acted on them, you can add the
`transformerAnnotations` option to the `buildMetadata` field of the kustomization:

```yaml
buildMetadata: [transformerAnnotations]
```

When the `transformerAnnotations` option is set, kustomize will add annotations with information about what transformers 
have acted on each resource. Transformers can be invoked either through various fields in the kustomization file 
(e.g. the `replacements` field will invoke the ReplacementTransformer), or through the `transformers` field.

The annotation key for transformer annotations will be `config.kubernetes.io/transformations`, which will contain a list of
transformer data. The possible fields in each item in this list is identical to the possible fields in `config.kubernetes.io/origin`,
except that the transformer annotation does not have a `path` field: 

The possible fields of these annotations are:

- `path`: The path to a resource file itself
- `ref`: If from a remote file or generator, the ref of the repo URL.
- `repo`: If from a remote file or generator, the repo source
- `configuredIn`: The path to the transformer config. If a transformer is invoked via a field 
in the kustomization file, this would point to the kustomization file itself. 
- `configuredBy`: The ObjectReference of the transformer config.  

All local file paths are relative to the top-level kustomization, i.e. the kustomization file in the directory upon 
which `kustomize build` was invoked. For example, if someone were to run `kustomize build foo`, all file paths 
in the annotation output would be relative to`foo/kustomization.yaml`. All remote file paths are relative to the root of the 
remote repository. Any fields that are not applicable would be omitted from the final output.  

Here is an example of what this would look like:
```yaml
config.kubernetes.io/transformations: |
  - ref: v0.0.0 
    repo: http://github.com/examplerepo 
    configuredIn: path/to/transformerconfig 
    configuredBy: 
      kind: Transformer
      apiVersion: builtin 
      name: foo
      namespace: default
```


While this field is in alpha, it will receive the `alpha` prefix, so you will see the annotation key 
`alpha.config.kubernetes.io/transformations` instead. We are not guaranteeing that the annotation content will be stable during
alpha, and reserve the right to make changes as we evolve the feature. 


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
