---
title: "ConfigMapGenerator"
linkTitle: "ConfigMapGenerator"
weight: 1
date: 2023-11-16
description: >
  Generate ConfigMap objects.
---

## ConfigMapGenerator
ConfigMapGenerator generates [ConfigMap] objects.

---

* **apiVersion**: builtin
* **kind**: ConfigMapGenerator
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **generatorArgs** ([GeneratorArgs]({{< relref "../Common%20Definitions/GeneratorArgs.md" >}}))

    GeneratorArgs contains arguments common to generators.

[ConfigMap]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/config-map-v1/
