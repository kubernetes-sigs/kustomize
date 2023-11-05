---
title: "Labels and Annotations"
linkTitle: "Labels and Annotations"
weight: 3
date: 2023-10-14
description: >
  Working with Labels and Annotations
---

A common set of labels can be applied to all Resources in a project by adding a `labels` or `commonLabels` entry to to the `kustomization.yaml` file. Similarly, a common set of annotations can be applied to Resources with the `commonAnnotations` field.

# Labels
## Add Labels
Add labels to all Resources in a project. This will override values for label keys that already exist. The `includeSelectors` and `includeTemplates` flags enable label propagation to Selectors and Templates respectively. These flags are disabled by default.

The following example adds labels to a Deployment and Service without changing the selector and Template labels.

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

3. Add labels with `kustomize build`.
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
## Add Template Labels
Add labels to Resource templates with the `labels.includeTemplate` field.

The following example adds labels and template labels to a Deployment and Service without changing the selector labels.

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

3. Add labels with `kustomize build`.
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
## Add Selector Labels
Add labels to Resource selectors and templates with the `labels.includeSelectors` field. Selector labels should not be changed after Workload and Service Resources have been created in a cluster. This is equivalent to using the `commonLabels` field.

The following example adds labels and selector labels to a Deployment and Service.

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

3. Add labels with `kustomize build`.
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

The following example produces the same result. The `commonLabels` field is equivalent to using `labels.includeSelectors`.
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

# Annotations
## Add Annotations
Add annotations to all Resources in a project with the `commonAnnotations` field. This will override values for annotations keys that already exist. annotations are propagated to Pod Templates.

The following example adds common annotations to a Deployment.

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

3. Add annotations with `kustomize build`.
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
