---
title: "patchesJson6902"
linkTitle: "patchesJson6902"
type: docs
weight: 16
description: >
    Patch resources using the [json 6902 standard](https://tools.ietf.org/html/rfc6902)
---

{{% pageinfo color="warning" %}}
The `patchesJson6902` field was deprecated in v5.0.0. This field will never be removed from the
kustomize.config.k8s.io/v1beta1 Kustomization API, but it will not be included
in the kustomize.config.k8s.io/v1 Kustomization API. When Kustomization v1 is available,
we will announce the deprecation of the v1beta1 version. There will be at least
two releases between deprecation and removal of Kustomization v1beta1 support from the
kustomize CLI, and removal itself will happen in a future major version bump.

Please move your `patchesJson6902` into
the [patches](/references/kustomize/kustomization/patches) field. This field supports patchesJson6902,
but with slightly different syntax. You can run `kustomize edit fix` to automatically convert
`patchesJson6902` to `patches`.
{{% /pageinfo %}}

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
   {"op": "replace", "path": "/some/existing/path", "value": "new value"},
   {"op": "copy", "from": "/some/existing/path", "path": "/some/path"},
   {"op": "move", "from": "/some/existing/path", "path": "/some/existing/destination/path"},
   {"op": "remove", "path": "/some/existing/path"},
   {"op": "test", "path": "/some/path", "value": "my-node-value"}
 ]
 ```

or in YAML format as

```yaml
# add: creates a new entry with a given value
- op: add
  path: /some/new/path
  value: value
# replace: replaces the value of the node with the new specified value
- op: replace
  path: /some/existing/path
  value: new value
# copy: copies the value specified in from to the destination path
- op: copy
  from: /some/existing/path
  path: /some/path
# move: moves the node specified in from to the destination path
- op: move
  from: /some/existing/path
  path: /some/existing/destination/path
# remove: delete's the node('s subtree)
- op: remove
  path: /some/path
# test: check if the specified node has the specified value, if the value differs it will throw an error
- op: test
  path: /some/path
  value: "my-node-value"
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

A patch can refer to a resource by any of its previous names or kinds.
For example, if a resource has gone through name-prefix transformations, it can refer to the
resource by its current name, original name, or any intermediate name that it had.