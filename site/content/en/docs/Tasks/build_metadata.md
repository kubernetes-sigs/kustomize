---
title: "Build Metadata"
linkTitle: "Build Metadata"
weight: 5
date: 2024-01-10
description: >
  Adding build information to labels or annotations.
---

Kustomize build information can be added to resource labels or annotations with the [`buildMetadata`] field.

## Add Managed By Label

Specify the `managedByLabel` option in the `buildMetadata` field to mark the resource as having been managed by Kustomize.

The following example adds the `app.kubernetes.io/managed-by` label to a resource.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
buildMetadata:
- managedByLabel
```

2. Create a Service manifest.
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
```

3. Add label with `kustomize build`.

The output shows that the `managedByLabel` option adds the `app.kubernetes.io/managed-by` label with Kustomize build information.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  labels:
    app.kubernetes.io/managed-by: kustomize-v5.2.1
spec:
  ports:
  - port: 7002
```

## Add Origin Annotation with Local Resource

Specify the `originAnnotations` option in the `buildMetadata` field to annotate resources with information about their origin.

The possible fields of these annotations are:

- `path`: The path to a resource file itself.
- `ref`: If from a remote file or generator, the git reference of the repository URL.
- `repo`: If from a remote file or generator, the repository source.
- `configuredIn`: The path to the generator configuration for a generated resource. This would point to the Kustomization file itself if a generator is invoked via a field.
- `configuredBy`: The ObjectReference of the generator configuration for a generated resource.

If the resource is from the `resources` field, this annotation contains data about the file it originated from. All local file paths are relative to the top-level Kustomization, i.e. the Kustomization file in the directory upon
which `kustomize build` was invoked. For example, if someone were to run `kustomize build foo`, all file paths
in the annotation output would be relative to `foo/kustomization.yaml`. All remote file paths are relative to the root of the
remote repository. Any fields that are not applicable would be omitted from the final output.

The following example adds the `config.kubernetes.io/origin` annotation to a non-generated resource defined in a local file.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
buildMetadata:
- originAnnotations
```

2. Create a Service manifest.
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
```

3. Add origin annotation with `kustomize build`.

The output shows that the `originAnnotations` option adds the `config.kubernetes.io/origin` annotation with Kustomize build information.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
spec:
  ports:
  - port: 7002
```

## Add Origin Annotation with Local Generator

Generated resources will receive an annotation containing data about the generator that produced it with the `originAnnotations` option.

The following example adds the `config.kubernetes.io/origin` annotation to a generated resource.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: my-java-server-env-vars
  literals:
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
buildMetadata:
- originAnnotations
```

2. Generate a ConfigMap that includes an origin annotation with `kustomize build`.

The output shows that the `originAnnotations` option adds the `config.kubernetes.io/origin` annotation with information about the local ConfigMapGenerator that generated the ConfigMap.

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: my-java-server-env-vars-c68g99m4hf
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        kind: ConfigMapGenerator
        apiVersion: builtin
data:
  JAVA_HOME: /opt/java/jdk
  JAVA_TOOL_OPTIONS: -agentlib:hprof
```


## Add Origin Annotation with Remote Generator
A remote file or generator will receive an annotation containing the repository URL and git reference with the `originAnnotations` option.

The following example adds the `config.kubernetes.io/origin` annotation to a resource generated with a remote generator.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- github.com/examplerepo/?ref=v1.0.6
buildMetadata:
- originAnnotations
```

2. This example uses a remote base with the following Kustomization.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: my-java-server-env-vars
  literals:
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
```

3. Generate a ConfigMap that includes an origin annotation with `kustomize build`.

The output shows that the `originAnnotations` option adds the `config.kubernetes.io/origin` annotation with build information about the remote ConfigMapGenerator that generated the ConfigMap.

```yaml
kind: ConfigMap
apiVersion: v1
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
data:
  JAVA_HOME: /opt/java/jdk
  JAVA_TOOL_OPTIONS: -agentlib:hprof
```


## Add Annotation with Local Transformer
**FEATURE STATE**: [alpha]

While this field is in alpha, it will receive the `alpha` prefix, so you will see the annotation key
`alpha.config.kubernetes.io/transformations` instead. We are not guaranteeing that the annotation content will be stable during alpha, and reserve the right to make changes as we evolve the feature.

Add the `transformerAnnotations` option to the `buildMetadata` field to annotate resources with information about the transformers that have acted on them.

When the `transformerAnnotations` option is set, Kustomize will add annotations with information about what transformers
have acted on each resource. Transformers can be invoked either through various fields in the Kustomization file
(e.g. the `replacements` field will invoke the ReplacementTransformer), or through the `transformers` field.

The annotation key for transformer annotations will be `alpha.config.kubernetes.io/transformations`, which will contain a list of transformer data. The possible fields in each item in this list is identical to the possible fields in `config.kubernetes.io/origin`, except that the transformer annotation does not have a `path` field:

The possible fields of these annotations are:

- `ref`: If from a remote file or generator, the git reference of the repository URL.
- `repo`: If from a remote file or generator, the repository source.
- `configuredIn`: The path to the transformer configuration. This would point to the Kustomization file itself if a transformer is invoked via a field.
- `configuredBy`: The ObjectReference of the transformer configuration.

All local file paths are relative to the top-level Kustomization. This behavior is similar to how the `originAnnotations` option works.

The following example adds the `alpha.config.kubernetes.io/transformations` annotation to a resource updated with the NamespaceTransformer.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: app
resources:
- service.yaml
buildMetadata:
- transformerAnnotations
```

2. Create a Service manifest.
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
```

3. Add transformer annotation with `kustomize build`.

The output shows that the `transformerAnnotations` option adds the `alpha.config.kubernetes.io/transformations` annotation with build information about the transformer that updated the resource.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: app
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: NamespaceTransformer
spec:
  ports:
  - port: 7002
```

## Add Annotation with Local and Remote Transformer

The following example adds the `alpha.config.kubernetes.io/transformations` annotation to a resource updated by a local and remote transformer.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: app
resources:
- github.com/examplerepo/?ref=v1.0.6
buildMetadata:
- transformerAnnotations
```

2. This example uses a remote base with the following Kustomization.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
namePrefix: pre-
```

The `service.yaml` contains the following:
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
```

3. Run `kustomize build`.

The output shows that the `transformerAnnotations` option adds the `alpha.config.kubernetes.io/transformations` annotation with build information about the transformers that updated the resource.

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: pre-deploy
  namespace: app
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

[`buildMetadata`]: /docs/reference/api/kustomization-file/buildmetadata/
