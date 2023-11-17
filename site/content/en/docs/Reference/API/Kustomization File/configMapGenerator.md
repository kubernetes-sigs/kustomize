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

* **configMapGenerator** ([][ConfigMapArgs](#configmapargs))

    List of metadata to generate ConfigMaps.


### ConfigMapArgs
ConfigMapArgs is metadata used to generate a ConfigMap.

---

* **GeneratorArgs** ([GeneratorArgs](/docs/reference/api/common-definitions/generatorargs/))

    GeneratorArgs contains arguments common to generators.


[Tasks section]: /docs/tasks/configmap_generator/
[ConfigMap]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/config-map-v1/
