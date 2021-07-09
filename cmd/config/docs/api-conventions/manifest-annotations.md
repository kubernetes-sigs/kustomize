# Manifest Annotations

This document lists the annotations that can be declared in resource manifests.

### `config.kubernetes.io/local-config`

A value of `"true"` for this annotation declares that the configuration is to
local tools rather than a remote Resource. e.g. The `Kustomization` config in a
`kustomization.yaml` **SHOULD** contain this annotation so that tools know it is
not intended to be sent to the Kubernetes API server.

Example:

```yaml
metadata:
  annotations:
    config.kubernetes.io/local-config: "true"
```

A value of `"false"` can be used to declare that a resource should be applied to
the cluster in situations where the tool assumes the resource is local.
