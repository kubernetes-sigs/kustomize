# Transformer Configurations

Kustomize creates new resources by applying a series of transformations to an original
set of resources. Kustomize provides the following default transformers:

- images
- namespace
- prefix/suffix
- labels
- annotations
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

## Images transformer

The images transformer updates the `image` sub-fields for all resources, based upon the `containers` and `initcontainers` fields.

Example usage in kustomization.yaml:

```yaml
images:
  - name: postgres
    newName: my-registry/my-postgres
    newTag: v1
  - name: nginx
    newTag: 1.8.0
  - name: my-demo-app
    newName: my-app
  - name: alpine
```

## Prefix/suffix transformer

The prefix/suffix transformer adds a prefix/suffix to the `metadata/name` field for all resources. Here is the default prefix transformer configuration:

```yaml
namePrefix:
- path: metadata/name
```

Example usage of `namePrefix` and `nameSuffix` in kustomization.yaml:

```yaml
namePrefix:
  alices-

nameSuffix:
  -v2
```

All Resources with path, `metadata/name` will be updated with the prefix, `alices-`.

## Labels transformer

The labels transformer adds labels to the `metadata/labels` field for all resources. It also adds labels to the `spec/selector` field in all Service resources as well as the `spec/selector/matchLabels` field in all Deployment resources.

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

Example usage in kustomization.yaml:

```yaml
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

## Annotations transformer

The annotations transformer adds annotations to the `metadata/annotations` field for all resources.
Annotations are also added to `spec/template/metadata/annotations` for Deployment,
ReplicaSet, DaemonSet, StatefulSet, Job, and CronJob resources, and `spec/jobTemplate/spec/template/metadata/annotations`
for CronJob resources.

Example usage in kustomization.yaml

```yaml
commonAnnotations:
  oncallPager: 800-555-1212
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

In addition to the default transformers, you can create custom transformer configurations. Save the default transformer configurations to a local directory by calling `kustomize config save -d`, and modify and use these configurations. This tutorial shows how to create custom transformer configurations:

- [support a CRD type](crd/README.md)
- add extra fields for variable substitution
- add extra fields for name reference
- modify image fields in a Custom Resource Definition (CRD)