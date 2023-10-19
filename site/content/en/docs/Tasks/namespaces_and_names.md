---
title: "Namespaces and Names"
linkTitle: "Namespaces and Names"
weight: 4
date: 2023-10-14
description: >
  Working with Namespaces and Names
---

The Namespace can be set for all Resources in a project by adding the `namespace` entry to the `kustomization.yaml` file. Consistent naming conventions can be applied to Resource Names in a project with the `namePrefix` and `nameSuffix` fields.

## Add Namespace
The Namespace for all namespaced Resources in a project can be set with the `namespace` field. This will override Namespace values that already exist.

The following example sets the Namespace of a Deployment.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-namespace
resources:
- deploy.yaml
```

2. Create a Deployment manifest.
```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
```

3. Add Namespace with `kustomize build`.
```bash
kustomize build .
```

The output shows that namespace is added to the metadata field.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  namespace: my-namespace
```

## Update object name prefix

The following example adds a prefix to the name of a Deployment.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: foo-
resources:
- deploy.yaml
```

2. Create a Deployment manifest.
```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
```

3. Add name prefix with `kustomize build`.
```bash
kustomize build .
```

The output shows that a prefix is added to the name.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-example
```

## Update object name suffix
The following example adds a suffix to the name of a Deployment.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameSuffix: -bar
resources:
- deploy.yaml
```

2. Create a Deployment manifest.
```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
```

3. Add name suffix with `kustomize build`.
```bash
kustomize build .
```

The output shows that a suffix is added to the name.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-bar
```
