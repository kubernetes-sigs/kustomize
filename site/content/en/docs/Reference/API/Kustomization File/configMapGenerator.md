---
title: "configMapGenerator"
linkTitle: "configMapGenerator"
type: docs
weight: 6
description: >
    Generate ConfigMap objects.
---

### configMapGenerator

ConfigMapGenerator generates ConfigMap objects.

See the [Tasks section] for examples of how to use `configMapGenerator`.

---

* **apiVersion**: kustomize.config.k8s.io/v1beta1
* **kind**: Kustomization
* **configMapGenerator** ([][ConfigMapArgs](#configmapargs))

    List of ConfigMaps to generate.


### ConfigMapArgs

ConfigMapArgs contains the metadata of the generated ConfigMap.

---

* **GeneratorArgs** ([GeneratorArgs](/docs/reference/api/common-definitions/generatorargs/))

    GeneratorArgs contains arguments common to generators.


[Tasks section]: /docs/tasks/configmap_generator/
