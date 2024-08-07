---
title: "configMapGenerator"
linkTitle: "configMapGenerator"
type: docs
weight: 6
description: >
    Generate ConfigMap objects.
---
`apiVersion: kustomize.config.k8s.io/v1beta1`

See the [Tasks section] for examples of how to use `configMapGenerator`.

### configMapGenerator
ConfigMapGenerator generates [ConfigMap] objects.

---

* **configMapGenerator** ([]ConfigMapArgs)

    List of metadata to generate ConfigMaps.

    _ConfigMapArgs represents metadata and options for ConfigMap generation._

    {{< include "../included/generatorargs.md" >}}


[Tasks section]: /docs/tasks/configmap_generator/
[ConfigMap]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/config-map-v1/
