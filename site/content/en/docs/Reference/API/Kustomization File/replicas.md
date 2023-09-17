---
title: "replicas"
linkTitle: "replicas"
type: docs
weight: 19
description: >
    Change the number of replicas for a resource.
---

Given this kubernetes Deployment fragment:

```yaml
kind: Deployment
metadata:
  name: deployment-name
spec:
  replicas: 3
```

one can change the number of replicas to 5
by adding the following to your kustomization:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

replicas:
- name: deployment-name
  count: 5
```

This field accepts a list, so many resources can
be modified at the same time.

As this declaration does not take in a `kind:` nor a `group:`
it will match any `group` and `kind` that has a matching name and
that is one of:

- `Deployment`
- `ReplicationController`
- `ReplicaSet`
- `StatefulSet`

For more complex use cases, revert to using a patch.

## Example

### Input File

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
      - name: the-container
        image: registry/container:latest
```

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

replicas:
- name: the-deployment
  count: 10

resources:
- deployment.yaml
```

### Output
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 10
  template:
    containers:
      - name: the-container
        image: registry/container:latest
```
