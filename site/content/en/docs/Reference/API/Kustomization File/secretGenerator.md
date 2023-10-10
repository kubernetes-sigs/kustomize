---
title: "secretGenerator"
linkTitle: "secretGenerator"
type: docs
weight: 21
description: >
    Generate Secret resources.
---

Each entry in the argument list results in the creation of one Secret resource (it's a generator of N secrets).

This works like the [configMapGenerator](/docs/reference/api/kustomization-file/configmapgenerator).

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

secretGenerator:
- name: app-tls
  files:
  - secret/tls.crt
  - secret/tls.key
  type: "kubernetes.io/tls"
- name: app-tls-namespaced
  # you can define a namespace to generate
  # a secret in, defaults to: "default"
  namespace: apps
  files:
  - tls.crt=catsecret/tls.crt
  - tls.key=secret/tls.key
  type: "kubernetes.io/tls"
- name: env_file_secret
  envs:
  - env.txt
  type: Opaque
- name: secret-with-annotation
  files:
  - app-config.yaml
  type: Opaque
  options:
    annotations:
      app_config: "true"
    labels:
      app.kubernetes.io/name: "app2"
```

Secret Resources may be generated much like ConfigMaps can. This includes generating them
from literals, files or environment files.

{{< alert color="success" title="Secret Syntax" >}}
Secret type is set using the `type` field.
{{< /alert >}}
