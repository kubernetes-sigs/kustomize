---
title: "secretGenerator"
linkTitle: "secretGenerator"
type: docs
weight: 21
description: >
    Generate Secret objects.
---

`apiVersion: kustomize.config.k8s.io/v1beta1`

See the [Tasks section] for examples of how to use `secretGenerator`.

### secretGenerator
SecretGenerator generates [Secret] objects.

---

* **secretGenerator** ([][SecretArgs](#secretargs))

    List of metadata to generate Secrets.


### SecretArgs
SecretArgs is metadata used to generate a Secret.

---

* **GeneratorArgs** ([GeneratorArgs](/docs/reference/api/common-definitions/generatorargs/))

    GeneratorArgs contains arguments common to generators.

* **type** (string), optional

    Type of the secret. Must be `Opaque` or `kubernetes.io/tls`. Defaults to `Opaque`.

[Tasks section]: /docs/tasks/secret_generator/
[Secret]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/secret-v1/
