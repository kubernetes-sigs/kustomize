---
title: "images"
linkTitle: "images"
type: docs
description: >
    修改镜像的名称、tag 或 image digest。
---

修改镜像的名称、tag 或 image digest ，而无需使用 patches 。例如，对于这种 kubernetes Deployment 片段：

```yaml
kind: Deployment
...
spec:
  template:
    spec:
      containers:
      - name: mypostgresdb
        image: postgres:8
      - name: nginxapp
        image: nginx:1.7.9
      - name: myapp
        image: my-demo-app:latest
      - name: alpine-app
        image: alpine:3.7
```

想要将 `image` 做如下更改：

- 将 `postgres:8` 改为 `my-registry/my-postgres:v1`
- 将 nginx tag 从 `1.7.9` 改为 `1.8.0`
- 将镜像名称 `my-demo-app` 改为 `my-app`
- 将 alpine 的 tag `3.7` 改为 digest 值

只需在 *kustomization* 中添加以下内容：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

images:
- name: postgres
  newName: my-registry/my-postgres
  newTag: v1
- name: nginx
  newTag: 1.8.0
- name: my-demo-app
  newName: my-app
- name: alpine
  digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
```
