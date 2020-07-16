---
title: "patchesStrategicMerge"
linkTitle: "patchesStrategicMerge"
type: docs
description: >
    Patch resources using the strategic merge patch standard.
---

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
