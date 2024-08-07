---
title: "buildMetadata"
linkTitle: "buildMetadata"
type: docs
weight: 3
description: >
    Specify options for including information about the build in annotations or labels.
---

`apiVersion: kustomize.config.k8s.io/v1beta1`

See the [Tasks section] for examples of how to use the `buildMetadata` field.

### buildMetadata
BuildMetadata specifies options for adding kustomize build information to resource labels or annotations.

---

* **buildMetadata** ([]string)

    List of strings used to toggle different build options. The strings can be one of three builtin
    options that add metadata to each resource about how the resource was built. It is possible to set one or all of these options in the kustomization file.

    These options are:
    - `managedByLabel`
    - `originAnnotations`
    - `transformerAnnotations`



[Tasks section]: /docs/tasks/build_metadata/
