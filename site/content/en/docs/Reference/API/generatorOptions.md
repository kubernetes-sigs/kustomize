---
title: "generatorOptions"
linkTitle: "generatorOptions"
type: docs
weight: 8
description: >
    Control behavior of [ConfigMap]() and
    [Secret]() generators.
---



Additionally, generatorOptions can be set on a per resource level within each
generator. For details on per-resource generatorOptions usage see
[field-name-configMapGenerator]() and See [field-name-secretGenerator]().

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

generatorOptions:
  # labels to add to all generated resources
  labels:
    kustomize.generated.resources: somevalue
  # annotations to add to all generated resources
  annotations:
    kustomize.generated.resource: somevalue
  # disableNameSuffixHash is true disables the default behavior of adding a
  # suffix to the names of generated resources that is a hash of
  # the resource contents.
  disableNameSuffixHash: true
  # if set to true, the immutable property is added to generated resources
  immutable: true
```

## Example I

Using ConfigMap

### Input Files

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: my-application-properties
  files:
  - application.properties
generatorOptions:
  labels:
    kustomize.generated.resources: config-label
  annotations:
    kustomize.generated.resource: config-annotation
```

```yaml
# application.properties
FOO=Bar
```

### Output File

```yaml
apiVersion: v1
data:
  application.properties: |-
    # application.properties
    FOO=Bar
kind: ConfigMap
metadata:
  annotations:
    kustomize.generated.resource: config-annotation
  labels:
    kustomize.generated.resources: config-label
  name: my-application-properties-f7mm6mhf59
```

## Example II

Using Secrets

### Input Files

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: app-tls
  files:
    - "tls.cert"
    - "tls.key"
  type: "kubernetes.io/tls"
generatorOptions:
  labels:
    kustomize.generated.resources: secret-label
  annotations:
    kustomize.generated.resource: secret-annotation
  disableNameSuffixHash: true
```

### Output File

```yaml
apiVersion: v1
data:
  tls.cert: TFMwdExTMUNSVWQuLi50Q2c9PQ==
  tls.key: TFMwdExTMUNSVWQuLi4wdExRbz0=
kind: Secret
metadata:
  annotations:
    kustomize.generated.resource: secret-annotation
  labels:
    kustomize.generated.resources: secret-label
  name: app-tls
type: kubernetes.io/tls
```
