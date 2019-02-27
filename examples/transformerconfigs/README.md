# Transformer Configurations

Kustomize computes the resources by applying a series of transformers:
- namespace transformer
- prefix/suffix transformer
- label transformer
- annotation transformer
- name reference transformer
- variable reference transformer

Each transformer takes a list of resources and modifies certain fields. The modification is based on the transformer's rule.
The fields to update is the transformer's configuration, which is a list of filedspec that can be represented in YAML format.

## fieldSpec
FieldSpec is a type to represent a path to a field in one kind of resources. It has following format
```
group: some-group
version: some-version
kind: some-kind
path: path/to/the/field
create: false
```
If `create` is set to true, it indicates the transformer to create the path if it is not found in the resources. This is most useful for label and annotation transformers, where the path for labels or annotations may not be set before the transformation.

## prefix/suffix transformer
Name prefix suffix transformer adds prefix and suffix to the `metadata/name` field for all resources with following configuration:
```
namePrefix:
- path: metadata/name
```

## label transformer
Label transformer adds labels to `metadata/labels` field for all resources. It also adds labels to `spec/selector` field in all Service and to `spec/selector/matchLabels` field in all Deployment.
```
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
  (etc.)  
```

## name reference transformer
Name reference transformer's configuration is different from all other transformers. It contains a list of namebackreferences, which represented all the possible fields that a type could be used as a reference in other types of resources. A namebackreference contains a type such as ConfigMap as well as a list of FieldSpecs where ConfigMap is referenced. Here is an example.
```
kind: ConfigMap
version: v1
FieldSpecs:
- kind: Pod
  version: v1
  path: spec/volumes/configMap/name
- kind: Deployment
  path: spec/template/spec/volumes/configMap/name
- kind: Job
  path: spec/template/spec/volumes/configMap/name
  (etc.)
```
Name reference transformer configuration contains a list of such namebackreferences for ConfigMap, Secret, Service, Role, ServiceAccount and so on.
```
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
    (etc.)
- kind: Secret
  version: v1
  fieldSpecs:
  - path: spec/volumes/secret/secretName
    version: v1
    kind: Pod
  - path: spec/containers/env/valueFrom/secretKeyRef/name
    version: v1
    kind: Pod
    (etc.)    
```

## customizing transformer configurations

Kustomize has a default set of configurations. They can be saved to local directory through `kustomize config save -d`. Kustomize allows modifying those configuration files and using them in kustomization.yaml file. This tutorial shows how to customize those configurations to
- [support a CRD type](crd/README.md)
- add extra fields for variable substitution
- add extra fields for name reference
