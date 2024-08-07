---
title: "commonLabels"
linkTitle: "commonLabels"
type: docs
weight: 4
description: >
    Add Labels and Selectors to all resources.
---
`apiVersion: kustomize.config.k8s.io/v1beta1`

See the [Tasks section] for examples of how to use `commonLabels`.

### commonLabels
Adds [Labels and Selectors] to resources.

* **commonLabels** (map[string]string)

    Map of labels to add to all resources. Labels will be added to resource selector and template fields where applicable.


[Tasks section]: /docs/tasks/labels_and_annotations/
[Labels and Selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
