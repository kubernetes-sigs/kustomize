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

* **emitStringData** (bool), optional

	  emitStringData if true generates a v1/Secret with plain-text stringData fields
	  instead of base64-encoded data fields. If a generating field does not have a
	  UTF-8 value, it falls back to being stored as a base64-encoded data field. This
	  is similar to the default binaryData fallback of a configMapGenerator.