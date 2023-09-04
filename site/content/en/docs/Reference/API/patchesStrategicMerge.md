---
title: "patchesStrategicMerge"
linkTitle: "patchesStrategicMerge"
type: docs
weight: 17
description: >
    Patch resources using the strategic merge patch standard.
---

{{% pageinfo color="warning" %}}
The `patchesStrategicMerge` field was deprecated in v5.0.0. This field will never be removed from the
kustomize.config.k8s.io/v1beta1 Kustomization API, but it will not be included
in the kustomize.config.k8s.io/v1 Kustomization API. When Kustomization v1 is available,
we will announce the deprecation of the v1beta1 version. There will be at least
two releases between deprecation and removal of Kustomization v1beta1 support from the
kustomize CLI, and removal itself will happen in a future major version bump.

Please move your `patchesStrategicMerge` into
the [patches](/references/kustomize/kustomization/patches) field. This field supports patchesStrategicMerge,
but with slightly different syntax. You can run `kustomize edit fix` to automatically convert
`patchesStrategicMerge` to `patches`.
{{% /pageinfo %}}

Each entry in this list should be either a relative
file path or an inline content
resolving to a partial or complete resource
definition.

The names in these (possibly partial) resource
files must match names already loaded via the
`resources` field.  These entries are used to
_patch_ (modify) the known resources.

Small patches that do one thing are best, e.g. modify
a memory request/limit, change an env var in a
ConfigMap, etc.  Small patches are easy to review and
easy to mix together in overlays.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesStrategicMerge:
- service_port_8888.yaml
- deployment_increase_replicas.yaml
- deployment_increase_memory.yaml
```

The patch content can be a inline string as well.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: nignx:latest
```

Note that kustomize does not support more than one patch
for the same object that contain a _delete_ directive. To remove
several fields / slice elements from an object create a single
patch that performs all the needed deletions.

A patch can refer to a resource by any of its previous names or kinds.
For example, if a resource has gone through name-prefix transformations, it can refer to the
resource by its current name, original name, or any intermediate name that it had.

## Patching custom resources

Strategic merge patches may require additional configuration via [openapi](../openapi) field to work as expected with custom resources. For example, if a resource uses a merge key other than `name` or needs a list to be merged rather than replaced, Kustomize needs openapi information informing it about this.
