---
title: "commonLabels"
linkTitle: "commonLabels"
type: docs
description: >
    Add labels and selectors to add all resources.
---

Add labels and selectors to all resources.  If the label key already is present on the resource,
the value will be overridden.

{{% pageinfo color="warning" %}}
Selectors for resources such as Deployments and Services shouldn't be changed once the
resource has been applied to a cluster.

Changing commonLabels to live resources could result in failures.
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

### File Input

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

### Build Output

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
