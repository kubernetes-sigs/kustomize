---
title: "patches"
linkTitle: "patches"
type: docs
weight: 15
description: >
    Patch resources
---

[strategic merge]: /references/kustomize/glossary#patchstrategicmerge
[JSON6902]: /references/kustomize/glossary#patchjson6902

Patches (also called overlays) add or override fields on resources.  They are provided using the
`patches` Kustomization field.

The `patches` field contains a list of patches to be applied in the order they are specified.

Each patch may:

- be either a [strategic merge] patch, or a [JSON6902] patch
- be either a file, or an inline string
- target a single resource or multiple resources

The patch target selects resources by `group`, `version`, `kind`, `name`, `namespace`, `labelSelector` and
`annotationSelector`. Any resource which matches all the **specified** fields has the patch applied
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

## Name and kind changes

With `patches` it is possible to override the kind or name of the resource it is
editing with the options `allowNameChange` and `allowKindChange`. For example:
```yaml
resources:
- deployment.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
  options:
    allowNameChange: true
    allowKindChange: true
```
By default, these fields are false and the patch will leave the kind and name of the resource untouched.

## Name references

A patch can refer to a resource by any of its previous names or kinds.
For example, if a resource has gone through name-prefix transformations, it can refer to the
resource by its current name, original name, or any intermediate name that it had.

## Patching custom resources

[Strategic merge] patches may require additional configuration via [openapi](../openapi) field to work as expected with custom resources. For example, if a resource uses a merge key other than `name` or needs a list to be merged rather than replaced, Kustomize needs openapi information informing it about this.

[JSON6902] patch usage is the same for built-in and custom resources.

## Examples

Consider the following `deployment.yaml` common for all examples:
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dummy-app
  labels:
    app.kubernetes.io/name: nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:stable
          ports:
            - name: http
              containerPort: 80
```

### Intents

- Make the container image point to a specific version and not to the latest container in the registry. 
- Adding a standard label containing the deployed version.

There are multiple possible strategies that all achieve the same results. 

### Patch using Inline Strategic Merge

```yaml
# kustomization.yaml
resources:
- deployment.yaml
patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: dummy-app
        labels:
          app.kubernetes.io/version: 1.21.0
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: not-used
      spec:
        template:
          spec:
            containers:
              - name: nginx
                image: nginx:1.21.0
    target:
      labelSelector: "app.kubernetes.io/name=nginx"
```

If a `target` is specified, the `name` contained in the metadata is required but not used.

### Patch using Inline JSON6902

```yaml
# kustomization.yaml
resources:
- deployment.yaml
patches:
  - patch: |-
      - op: add
        path: /metadata/labels/app.kubernetes.io~1version
        value: 1.21.0
    target:
      group: apps
      version: v1
      kind: Deployment
  - patch: |-
      - op: replace
        path: /spec/template/spec/containers/0/image
        value: nginx:1.21.0
    target:
      labelSelector: "app.kubernetes.io/name=nginx"
```

The `target` field is always required for JSON6902 patches.  
A special replacement character `~1` is used to replace `/` in label name.

### Patch using Path Strategic Merge

```yaml
# kustomization.yaml
resources:
- deployment.yaml
patches:
  - path: add-label.patch.yaml
  - path: fix-version.patch.yaml
    target:
      labelSelector: "app.kubernetes.io/name=nginx"
```

As with the Inline Strategic Merge, the `target` field can be omitted.
In that case, the target resource is matched using 
the `apiVersion`, `kind` and `name` from the patch.

```yaml
# add-label.patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dummy-app
  labels:
    app.kubernetes.io/version: 1.21.0
```

```yaml
# fix-version.patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-used
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.21.0
```

As with the Inline Strategic Merge, the `name` field in the patch is not used when a `target` is specified.

### Patch using Path JSON6902

```yaml
# kustomization.yaml
resources:
- deployment.yaml
patches:
  - path: add-label.patch.json
    target:
      group: apps
      version: v1
      kind: Deployment
  - path: fix-version.patch.yaml
    target:
      labelSelector: "app.kubernetes.io/name=nginx"
```

As with Inline JSON6902, the `target` field is mandatory.

```yaml
# add-label.patch.json
[
  {"op": "add", "path": "/metadata/labels/app.kubernetes.io~1version", "value": "1.21.0"}
]
```

```yaml
# fix-version.patch.yaml
- op: replace
  path: /spec/template/spec/containers/0/image
  value: nginx:1.21.0
```

External patch file can be written both as YAML or JSON.
The content must follow the JSON6902 standard.

### Build Output

All four patches strategies lead to the exact same output:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: nginx
    app.kubernetes.io/version: 1.21.0
  name: dummy-app
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: nginx
    spec:
      containers:
        - image: nginx:1.21.0
          name: nginx
          ports:
            - containerPort: 80
              name: http
```
