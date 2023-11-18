---
title: "Namespaces and Names"
linkTitle: "Namespaces and Names"
weight: 4
date: 2023-10-14
description: >
  Working with Namespaces and Names
---

The Namespace can be set for all Resources in a project by adding the [`namespace`] entry to the `kustomization.yaml` file. Consistent naming conventions can be applied to Resource Names in a project with the [`namePrefix`] and [`nameSuffix`] fields.

## Working with Namespaces
[`namespace`] sets the Namespace for all namespaced Resources in a project. This sets the Namespace for both generated Resources (e.g. ConfigMaps and Secrets) and non-generated Resources. This will override Namespace values that already exist.

### Add Namespace
Here is an example of how to set the Namespace of a Deployment and a generated ConfigMap. The ConfigMap is generated with [`configMapGenerator`].

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-namespace

configMapGenerator:
- name: my-config
  literals:
  - FOO=BAR

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

The output shows that the `namespace` field is used to set the Namespace of the Deployment and the generated ConfigMap.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  namespace: my-namespace
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config-m2mg5mb749
  namespace: my-namespace
data:
  FOO: BAR
```

## Working with Names
A prefix or suffix can be set for all Resources in a project with the [`namePrefix`] and [`nameSuffix`] fields. This sets a name prefix and suffix for both generated Resources (e.g. ConfigMaps and Secrets) and non-generated Resources.

Resources such as Deployments and StatefulSets may reference other Resources such as ConfigMaps and Secrets in the Pod Spec. The name prefix and suffix will also propagate to Resource references in a project. Typical uses cases include Service reference from a StatefulSet, ConfigMap reference from a Pod Spec, and Secret reference from a Pod Spec.

### Add Name Prefix
[`namePrefix`] can be used to add a prefix to the name of all Resources in a project.

Here is an example of how to add a prefix to a Deployment.

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

The output shows that the `namePrefix` field is used to add a prefix to the name of the Deployment.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-example
```

### Add Name Suffix
[`nameSuffix`] can be used to add a suffix to the name of all Resources in a project.

Here is an example of how to add a suffix to the name of a Deployment and a generated ConfigMap. The ConfigMap is generated with [`configMapGenerator`].

1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameSuffix: -bar

configMapGenerator:
- name: my-config
  literals:
  - FOO=BAR

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

The output shows that the `nameSuffix` field is used to add a suffix to the name of the Deployment and the generated ConfigMap.
```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: my-config-bar-m2mg5mb749
data:
  FOO: BAR
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-bar
```

### Propagate Name Prefix to Resource Reference
[`namePrefix`] and [`nameSuffix`] propagate Resources name changes to Resource references in a project.

Here is an example of how the name prefix of a generated ConfigMap is propagated to the Pod Spec of a Deployment that references the ConfigMap to set a container environment variable.
1. Create a Kustomization file.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: foo-

configMapGenerator:
- name: special-config
  literals:
  - special.how=very

resources:
- deploy.yaml
```

2. Create a Deployment manifest. This Deployment is configured to set an environment variable in the `busybox` container using data from the generated ConfigMap.
```yaml
# deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example
  name: example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - image: registry.k8s.io/busybox
        name: busybox
        command: [ "/bin/sh", "-c", "env" ]
        env:
          - name: SPECIAL_LEVEL_KEY
            valueFrom:
              configMapKeyRef:
                name: special-config
                key: special.how
```

3. Add name prefix with `kustomize build`.
```bash
kustomize build .
```

The output shows that the name prefix is propagated to the ConfigMap name reference in the Deployment Pod Spec.
```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: foo-special-config-9k6fhm8659
data:
  special.how: very
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example
  name: foo-example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - env
        env:
        - name: SPECIAL_LEVEL_KEY
          valueFrom:
            configMapKeyRef:
              key: special.how
              name: foo-special-config-9k6fhm8659
        image: registry.k8s.io/busybox
        name: busybox
```

[`namespace`]: /docs/reference/api/kustomization-file/namespace/
[`namePrefix`]: /docs/reference/api/kustomization-file/nameprefix/
[`nameSuffix`]: /docs/reference/api/kustomization-file/namesuffix/
[`configMapGenerator`]: /docs/reference/api/kustomization-file/configmapgenerator/
