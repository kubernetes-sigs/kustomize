---
title: "generatorOptions"
linkTitle: "generatorOptions"
type: docs
description: >
    控制生成 [ConfigMap](/kustomize/zh/api-reference/kustomization/configmapgenerator) 和 [Secret](/kustomize/zh/api-reference/kustomization/secretgenerator) 的行为。
---

此外，在每个生成器中，还可以按每个资源级别设置 generatorOptions，具体使用方法请参见[configMapGenerator](/kustomize/zh/api-reference/kustomization/configmapgenerator)和[secretGenerator](/kustomize/zh/api-reference/kustomization/secretgenerator)。

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
