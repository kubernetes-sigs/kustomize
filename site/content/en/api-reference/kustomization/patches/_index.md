---
title: "patches"
linkTitle: "patches"
type: docs
description: >
    Patch resources
---

[strategic merge]: /kustomize/api-reference/glossary#patchstrategicmerge
[JSON]: /kustomize/api-reference/glossary#patchjson6902

Patches (also call overlays) add or override fields on resources.  They are provided using the
`patches` Kustomization field.

The `patches` field contains a list of patches to be applied in the order they are specified.

Each patch may:

- be either a [strategic merge] patch, or a [JSON] patch
- be either a file, or an inline string
- target a single resource or multiple resources

The patch target selects resources by group, version, kind, name, namespace, labelSelector and
annotationSelector.  Any resource which matches all the **specified** fields has the patch applied
to it (regular expressions).

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patches:
- path: patch.yaml
  target:
    group: apps
    version: v1
    kind: Deployment
    name: deploy.*
    labelSelector: "env=dev"
    annotationSelector: "zone=west"
- patch: |-
    - op: replace
      path: /some/existing/path
      value: new value
  target:
    kind: MyKind
    labelSelector: "env=dev"
```

The `name` and `namespace` fields of the patch target selector are
automatically anchored regular expressions. This means that the value `myapp`
is equivalent to `^myapp$`. 

Consider the following `deployment.yaml` common for both the examples:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
      - name: the-container
        image: registry/conatiner:latest
```

## Example I

### Intent

To Make the container image point to a specific version and not to the latest container in the
registry.

### File Input

```yaml
# kustomization.yaml
resources:
- deployment.yaml

patches:
- path: patch.yaml
```

```yaml
# patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  template:
    containers:
      - name: the-container
        image: registry/conatiner:1.0.0
```

### Build Output

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
    - image: registry/conatiner:1.0.0
      name: the-container
```

## Example II

### Intent

To Make the container image point to a specific version and not to the latest container in the
registry.

### File Input

```yaml
# kustomization.yaml
resources:
- deployment.yaml

patches:
- target:
    kind: Deployment
    name: the-deployment
  path: patch.json
```

```yaml
# patch.json
[
   {"op": "replace", "path": "/spec/template/containers/0/image", "value": "registry/conatiner:1.0.0"}
]

```

### Build Output

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 5
  template:
    containers:
    - image: registry/conatiner:1.0.0
      name: the-container
```
