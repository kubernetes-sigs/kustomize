---
title: "replicas"
linkTitle: "replicas"
type: docs
description: >
    Change the number of replicas for a resource.
---

Given this kubernetes Deployment fragment:

```
# deployment.yaml
kind: Deployment
metadata:
  name: deployment-name
spec:
  replicas: 3
```

one can change the number of replicas to 5
by adding the following to your kustomization:

```
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
