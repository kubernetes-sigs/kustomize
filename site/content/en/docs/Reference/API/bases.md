---
title: "bases"
linkTitle: "bases"
type: docs
weight: 1
description: >
    Add resources from a kustomization dir.
---

{{% pageinfo color="warning" %}}
The `bases` field was deprecated in v2.1.0. This field will never be removed from the
kustomize.config.k8s.io/v1beta1 Kustomization API, but it will not be included
in the kustomize.config.k8s.io/v1 Kustomization API. When Kustomization v1 is available,
we will announce the deprecation of the v1beta1 version. There will be at least
two releases between deprecation and removal of Kustomization v1beta1 support from the
kustomize CLI, and removal itself will happen in a future major version bump.

You can run `kustomize edit fix` to automatically convert `bases` to `resources`.
{{% /pageinfo %}}

Move entries into the [resources](/references/kustomize/kustomization/resource)
field.  This allows bases - which are still a
[central concept](/references/kustomize/glossary#base) - to be
ordered relative to other input resources.
