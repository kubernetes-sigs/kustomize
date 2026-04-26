---
title: "Local Configuration"
linkTitle: "Local Configuration"
weight: 70
description: >
  What is the config.kubernetes.io/local-config annotation?
---

[well-known annotations]: https://kubernetes.io/docs/reference/labels-annotations-taints/#config-kubernetes-io-local-config
[manifest-annotations.md]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/manifest-annotations.md

The `config.kubernetes.io/local-config` annotation marks a resource as
local to the build: kustomize reads it like any other input, but omits
it from the rendered output. The annotation is part of the [well-known
annotations] shared across the KRM tooling ecosystem.

A resource is local-config when it carries the annotation with the
value `"true"`:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: settings
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  region: us-east-1
```

Setting the annotation to `"false"`, or omitting it entirely, keeps
the resource in the output.

## Why it is useful

Local-config resources let you keep build-time configuration alongside
the resources you intend to apply, without having to manage that
configuration in a separate file tree. They participate in name
references and can be consumed by transformers, generators, and
functions, but they never reach the cluster.

### As a data source for a Replacement

A `Replacement` can read fields from a local-config resource and copy
them into one or more target resources. The source object exists only
to carry the values:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: image-source
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  image: registry.example.com/app:v1.2.3
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
replacements:
  - source:
      kind: ConfigMap
      name: image-source
      fieldPath: data.image
    targets:
      - select:
          kind: Deployment
          name: app
        fieldPaths:
          - spec.template.spec.containers.0.image
```

The rendered output contains the patched `Deployment` only; the
`ConfigMap` is dropped.

### As a name-reference anchor across multiple objects

When several resources share the same name fragment (for example, a
hostname that appears in an `Ingress`, a `Secret`, and a `Deployment`
env var), a local-config resource can serve as the canonical name. A
`namePrefix` applied to the build then propagates to every reference
through kustomize's name-reference machinery, while the anchor itself
is removed from the output.

A complete worked example lives in
[`api/krusty/namereference_test.go`][nameref-test] under
`TestIssue4884_UseLocalConfigAsNameRefSource`.

[nameref-test]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/krusty/namereference_test.go

### As shared configuration for KRM Functions

A function declared in `transformers:` or `generators:` is itself a
KRM resource. Marking it local-config keeps the function definition in
the kustomization, where it can be patched by overlays, without
emitting the function spec to the cluster:

```
apiVersion: example.com/v1
kind: ValidatingTransformer
metadata:
  name: enforce-labels
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/function: |
      container:
        image: registry.example.com/fn-enforce-labels:v1
spec:
  required:
    - app.kubernetes.io/name
```

## How kustomize processes local-config resources

Internally, kustomize uses the `IsLocalConfig` filter from
[`kyaml/kio/filters/local.go`][local-go] (and the related
`DropLocalNodes` helper in `api/resource/factory.go`) to skip
local-config resources when assembling the build output. A resource
is dropped when the annotation is present with any value other than
`"false"`.

[local-go]: https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/kio/filters/local.go

The companion `kustomize cfg cat` command exposes two flags for
inspecting these resources directly:

- `--include-local` includes local-config resources in the output.
- `--exclude-non-local` keeps only local-config resources.

## See also

- [Kubernetes well-known annotations][well-known annotations]
- [Manifest annotations reference][manifest-annotations.md]
