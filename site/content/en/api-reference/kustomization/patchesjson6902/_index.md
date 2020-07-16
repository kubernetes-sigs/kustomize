---
title: "patchesJson6902"
linkTitle: "patchesJson6902"
type: docs
description: >
    Patch resources using the [json 6902 standard](https://tools.ietf.org/html/rfc6902)
---

Each entry in this list should resolve to a kubernetes object and a JSON patch that will be applied
to the object.
The JSON patch is documented at <https://tools.ietf.org/html/rfc6902>

target field points to a kubernetes object within the same kustomization
by the object's group, version, kind, name and namespace.
path field is a relative file path of a JSON patch file.
The content in this patch file can be either in JSON format as

```json
 [
   {"op": "add", "path": "/some/new/path", "value": "value"},
   {"op": "replace", "path": "/some/existing/path", "value": "new value"}
 ]
 ```

or in YAML format as

```yaml
- op: add
  path: /some/new/path
  value: value
- op: replace
  path: /some/existing/path
  value: new value
```

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  path: add_init_container.yaml
- target:
    version: v1
    kind: Service
    name: my-service
  path: add_service_annotation.yaml
```

The patch content can be an inline string as well:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  patch: |-
    - op: add
      path: /some/new/path
      value: value
    - op: replace
      path: /some/existing/path
      value: "new value"
```
