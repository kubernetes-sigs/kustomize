# Transformer Configurations

Kustomize creates new resources by applying a series of transformations to an original
set of resources. Kustomize provides the following default transformers:

- namespace
- prefix/suffix
- label
- annotation
- name reference
- variable reference

A `fieldSpec` list, in a transformer's configuration, determines which resource types and which fields
within those types the transformer can modify.

## FieldSpec

FieldSpec is a type that represents a path to a field in one kind of resource.

```yaml
group: some-group
version: some-version
kind: some-kind
path: path/to/the/field
create: false
```

If `create` is set to `true`, the transformer creates the path to the field in the resource if the path is not already found. This is most useful for label and annotation transformers, where the path for labels or annotations may not be set before the transformation.

## Prefix/suffix transformer

The prefix/suffix transformer adds a prefix/suffix to the `metadata/name` field for all resources. Here is the default prefix transformer configuration:

```yaml
namePrefix:
- path: metadata/name
```

## Label transformer

The label transformer adds labels to the `metadata/labels` field for all resources. It also adds labels to the `spec/selector` field in all Service resources as well as the `spec/selector/matchLabels` field in all Deployment resources.

```yaml
commonLabels:
- path: metadata/labels
  create: true

- path: spec/selector
  create: true
  version: v1
  kind: Service

- path: spec/selector/matchLabels
  create: true
  kind: Deployment
```

## Name reference transformer

Name reference transformer's configuration is different from all other transformers. It contains a list of `nameReferences`, which represent all of the possible fields that a type could be used as a reference in other types of resources. A `nameReference` contains a type such as ConfigMap as well as a list of `fieldSpecs` where ConfigMap is referenced in other resources. Here is an example:

```yaml
kind: ConfigMap
version: v1
fieldSpecs:
- kind: Pod
  version: v1
  path: spec/volumes/configMap/name
- kind: Deployment
  path: spec/template/spec/volumes/configMap/name
- kind: Job
  path: spec/template/spec/volumes/configMap/name
```

Name reference transformer's configuration contains a list of `nameReferences` for resources such as ConfigMap, Secret, Service, Role, and ServiceAccount. Here is an example configuration:

```yaml
nameReference:
- kind: ConfigMap
  version: v1
  fieldSpecs:
  - path: spec/volumes/configMap/name
    version: v1
    kind: Pod
  - path: spec/containers/env/valueFrom/configMapKeyRef/name
    version: v1
    kind: Pod
  # ...
- kind: Secret
  version: v1
  fieldSpecs:
  - path: spec/volumes/secret/secretName
    version: v1
    kind: Pod
  - path: spec/containers/env/valueFrom/secretKeyRef/name
    version: v1
    kind: Pod
```

## Customizing transformer configurations

Kustomize has a default set of transformer configurations. You can save the default transformer configurations to a local directory by calling `kustomize config save -d`, and modify and use these configurations. Kustomize also supports adding new transformer configurations to kustomization.yaml. This tutorial shows how to customize those configurations to:

- [support a CRD type](crd/README.md)
- add extra fields for variable substitution
- add extra fields for name reference