---
title: "Labels and Annotations"
linkTitle: "Labels and Annotations"
weight: 3
date: 2023-10-14
description: >
  Working with Labels and Annotations
---

A common set of labels can be applied to all Resources in a project by adding a [`labels`] or [`commonLabels`] entry to the `kustomization.yaml` file. Similarly, a common set of annotations can be applied to Resources with the [`commonAnnotations`] field.

## Working with Labels
### Add Labels
[`labels`] can be used to add labels to the `metadata` field of all Resources in a project. This will override values for label keys that already exist.

Here is an example of how to add labels to the `metadata` field.
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

The output shows that the `labels` field is used to add labels to the `metadata` field of the Service and Deployment Resources.
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
### Add Template Labels
[`labels.includeTemplates`] can be used to add labels to the template field of all applicable Resources in a project.

Here is an example of how to add labels to the template field of a Deployment.
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

The output shows that labels are added to the `metadata` field and the `labels.includeTemplates` field is used to add labels to the template field of the Deployment. However, the [Service] Resource does not have a template field, and Kustomize does not add this field.

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
### Add Selector Labels
[`labels.includeSelectors`] can be used to add labels to the selector field of applicable Resources in a project. Note that this also adds labels to the template field for applicable Resources.

Labels added to the selector field should not be changed after Workload and Service Resources have been created in a cluster.

Here is an example of how to add labels to the selector field.
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

The output shows that labels are added to the `metadata` field and the `labels.includeSelectors` field is used to add labels to the selector and template fields for applicable Resources. However, the [Service] Resource does not have a template field, and Kustomize does not add this field.
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

The following example produces the same result. The [`commonLabels`] field is equivalent to using [`labels.includeSelectors`].
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

## Working with Annotations
### Add Annotations
[`commonAnnotations`] can be used to add annotations to all Resources in a project. This will override values for annotations keys that already exist. Annotations are propagated to the Deployment Pod template.

Here is an example of how to add annotations to a Deployment.
1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  oncallPager: 800-867-5309

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
The output shows that the `commonAnnotations` field is used to add annotations to a Deployment.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  annotations:
    oncallPager: 800-867-5309
spec:
  template:
    metadata:
      annotations:
        oncallPager: 800-867-5309
```

[`labels`]: /docs/reference/api/kustomization-file/labels/
[`labels.includeTemplates`]: /docs/reference/api/kustomization-file/labels/
[`labels.includeSelectors`]: /docs/reference/api/kustomization-file/labels/
[`commonLabels`]: /docs/reference/api/kustomization-file/commonlabels/
[`commonAnnotations`]: /docs/reference/api/kustomization-file/commonannotations/
[Service]: https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/
