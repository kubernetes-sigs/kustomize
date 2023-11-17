---
title: "SecretGenerator"
linkTitle: "SecretGenerator"
weight: 2
date: 2023-11-16
description: >
  Generate Secret objects.
---

## SecretGenerator
SeretGenerator generates [Secret] objects.

---

* **apiVersion**: builtin
* **kind**: SecretGenerator
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **generatorArgs** ([GeneratorArgs]({{< relref "../Common%20Definitions/GeneratorArgs.md" >}}))

    GeneratorArgs contains arguments common to generators.

* **type** (string), optional

    Type of the secret. Must be `Opaque` or `kubernetes.io/tls`. Defaults to `Opaque`.

[Secret]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/secret-v1/
