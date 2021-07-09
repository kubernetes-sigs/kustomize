# Manifest Annotations

This document lists the annotations that can be declared in resource manifests.

### `config.kubernetes.io/local-config`

A value of `"true"` for this annotation declares that the resource is only consumed by
client-side tooling and should not be applied to the API server.

A value of `"false"` can be used to declare that a resource should be applied to
the API server even when it is assumed to be local.
