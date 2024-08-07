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

* **secretGenerator** ([]SecretArgs)

    List of metadata to generate Secrets.

     _SecretArgs represents metadata and options for Secret generation._

    {{< include "../included/secretargs.md" >}}

[Tasks section]: /docs/tasks/secret_generator/
[Secret]: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/secret-v1/
