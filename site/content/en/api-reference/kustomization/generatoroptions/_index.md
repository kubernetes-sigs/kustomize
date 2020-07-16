---
title: "generatorOptions"
linkTitle: "generatorOptions"
type: docs
description: >
    Control behavior of [ConfigMap](/kustomize/api-reference/kustomization/configmapgenerator) and
    [Secret](/kustomize/api-reference/kustomization/secretgenerator) generators.
---



Additionally, generatorOptions can be set on a per resource level within each
generator. For details on per-resource generatorOptions usage see
[field-name-configMapGenerator](/kustomize/api-reference/kustomization/configmapgenerator) and See [field-name-secretGenerator](/kustomize/api-reference/kustomization/secretgenerator).

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
```
