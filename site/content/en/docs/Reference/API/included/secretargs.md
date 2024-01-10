---
title: "secretargs"
weight: 2
date: 2023-11-17
description: >
  SecretArgs contains arguments to generate Secrets.
headless: true
_build:
  list: never
  render: never
  publishResources: false
---

{{< include "generatorargs.md" >}}

* **type** (string), optional

    Type of the secret. Must be `Opaque` or `kubernetes.io/tls`. Defaults to `Opaque`.
