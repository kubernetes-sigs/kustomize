---
title: "commonAnnotations"
linkTitle: "commonAnnotations"
type: docs
description: >
    为所有字段添加注释。
---

为所有资源添加注释，如果资源上已经存在注解键，该值将被覆盖。

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-555-1212
```

## Example

### 输入文件

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-555-1212

resources:
- deploy.yaml
```

```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  ...
```

### 构建输出

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  annotations:
    oncallPager: 800-555-1212
spec:
  ...
```
