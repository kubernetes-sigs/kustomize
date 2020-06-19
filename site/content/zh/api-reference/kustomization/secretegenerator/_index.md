---
title: "secretGenerator"
linkTitle: "secretGenerator"
type: docs
description: >
    生成 Secret 资源。
---

列表中的每个条目都将生成一个 Secret（合计可以生成 n 个 Secrets）。

功能与之前描述的 [configMapGenerator](/kustomize/zh/api-reference/kustomization/configmapgenerator) 字段类似。

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

secretGenerator:
- name: app-tls
  files:
  - secret/tls.cert
  - secret/tls.key
  type: "kubernetes.io/tls"
- name: app-tls-namespaced
  # you can define a namespace to generate
  # a secret in, defaults to: "default"
  namespace: apps
  files:
  - tls.crt=catsecret/tls.cert
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
