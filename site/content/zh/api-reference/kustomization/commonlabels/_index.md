---
title: "commonLabels"
linkTitle: "commonLabels"
type: docs
description: >
    为所有资源和 selectors 增加标签。
---

为所有资源和 selectors 增加标签。如果资源上已经存在注解键，该值将被覆盖。

{{% pageinfo color="warning" %}}
一旦将资源应用于集群，就不应更改诸如 Deployments 和 Services 之类的资源选择器。

将 commonLabels 更改为可变资源可能会导致部署失败。
{{% /pageinfo %}}

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

## Example

### 文件输入

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  someName: someValue
  owner: alice
  app: bingo

resources:
- deploy.yaml
- service.yaml
```

```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
```

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: example
```

### 构建输出

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bingo
    owner: alice
    someName: someValue
  name: example
spec:
  selector:
    app: bingo
    owner: alice
    someName: someValue
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: bingo
    owner: alice
    someName: someValue
  name: example
spec:
  selector:
    matchLabels:
      app: bingo
      owner: alice
      someName: someValue
  template:
    metadata:
      labels:
        app: bingo
        owner: alice
        someName: someValue
```
