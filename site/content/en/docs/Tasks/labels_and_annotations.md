---
title: "Labels and Annotations"
linkTitle: "Labels and Annotations"
weight: 3
date: 2023-10-14
description: >
  Working with Labels and Annotations
---

A common set of Labels and Annotations can be applied to all Resources in a project by adding a `commonLabels` and a `commonAnnotations` entry to the `kustomization.yaml` file.

## Add common Labels
Add Labels to all Resources in a project with the `commonLabels` field. This will override values for Label keys that already exist. The `commonLabels` field also adds Labels to Selectors and Templates. Selector Labels should not be changed after Workload and Service Resources have been created in a cluster.

The following example adds common Labels to a Deployment, the Pod Selectors, and the Pod Template spec.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: foo
  environment: test
resources:
- deploy.yaml
```

2. Create a Deployment manifest.
```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
    bar: baz
spec:
  selector:
    matchLabels:
      app: nginx
      bar: baz
  template:
    metadata:
      labels:
        app: nginx
        bar: baz
    spec:
      containers:
      - name: nginx
        image: nginx
```

3. Add Labels with `kustomize build`.
```bash
kustomize build .
```

The output shows that labels are added and updated in the metadata, selector, and template fields.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: foo # Update
    environment: test # Add
    bar: baz # Existing labels are ignored
spec:
  selector:
    matchLabels:
      app: foo # Update
      environment: test # Add
      bar: baz # Existing labels are ignored
  template:
    metadata:
      labels:
        app: foo # Update
        environment: test # Add
        bar: baz # Existing labels are ignored
    spec:
      containers:
      - image: nginx
        name: nginx
```

## Disable Selector and Template updates
Add Labels to Resources in a project without propagating the Labels to the Selectors with the `labels` field. The `includeSelectors` and `includeTemplates` flags enable Label propagation to Selectors and Templates respectivley. These flags are disabled by default.

The following example adds Labels to a Deployment and Service without changing the Selector and Template Labels.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

labels:
  - pairs:
      someName: someValue
      owner: alice
      app: bingo

resources:
- deploy.yaml
- service.yaml
```

2. Create Deployment and Service manifests.
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

3. Add Labels with `kustomize build`.
```bash
kustomize build .
```
The output shows that labels are added to the metadata field while the template and selector fields remain unchanged.
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bingo
    owner: alice
    someName: someValue
  name: example
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: bingo
    owner: alice
    someName: someValue
  name: example
```

## Update Selectors with Labels

The following example adds Labels and Selector Labels to a Deployment and Service. This is equivalent to using the `commonLabels` field.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

labels:
  - pairs:
      someName: someValue
      owner: alice
      app: bingo
    includeSelectors: true

resources:
- deploy.yaml
- service.yaml
```

2. Create Deployment and Service manifests.
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

3. Add Labels with `kustomize build`.
```bash
kustomize build .
```
The output shows that labels are added to the metadata, selector, and template fields.
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

## Update Templates with Labels

The following example adds Labels and Template Labels to a Deployment and Service without changing the Selector Labels.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

labels:
  - pairs:
      someName: someValue
      owner: alice
      app: bingo
    includeTemplates: true

resources:
- deploy.yaml
- service.yaml
```

2. Create Deployment and Service manifests.
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

3. Add Labels with `kustomize build`.
```bash
kustomize build .
```
The output shows that labels are added to the metadata and template fields while the selector field remains unchanged.
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bingo
    owner: alice
    someName: someValue
  name: example
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
  template:
    metadata:
      labels:
        app: bingo
        owner: alice
        someName: someValue
```

## Add common Annotations
Add Annotations to all Resources in a project with the `commonAnnotations` field. This will override values for Annotations keys that already exist. Annotations are propagated to Pod Templates.

The following example adds common Annotations to a Deployment.

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-555-1212

commonLabels:
  app: bingo
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

3. Add Annotations with `kustomize build`.
```bash
kustomize build .
```
The output shows that annotations are added to the metadata field.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    oncallPager: 800-555-1212
  labels:
    app: bingo
  name: example
spec:
  selector:
    matchLabels:
      app: bingo
  template:
    metadata:
      annotations:
        oncallPager: 800-555-1212
      labels:
        app: bingo
```
