---
title: "Labels and Annotations"
linkTitle: "Labels and Annotations"
weight: 4
date: 2017-01-05
description: >
  Working with Labels and Annotations
---

{{< alert color="success" title="TL;DR" >}}
- Set Labels for all Resources declared within a project with `commonLabels`
- Set Annotations for all Resources declared within a project with `commonAnnotations`
{{< /alert >}}

## Motivation
Users may want to define a common set of labels or annotations for all the Resources in a project.

Typical use cases include:
- Identify the Resources within a project by querying their labels.
- Set metadata for all Resources within a project (e.g. environment=test).
- Copy or fork an existing project and add or change labels and annotations.


## Setting Labels
### Propagating Labels to Selectors
In addition to updating the labels for each Resource, any selectors will also be updated to target the labels. For instance, the selectors for Services in the project will be updated to include the `commonLabels` *in addition* to the other labels.

**Note:** Once set, `commonLabels` should not be changed so as not to change the Selectors for Services or Workloads.

### Common Labels
The kubernetes.io documentation defines a set of [Common Labeling Conventions](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/) that may be applied to applications.

**Note:** `commonLabels` should only be set for immutable labels because they will be applied to Selectors.

Labeling Workload Resources simplifies queries for Pods and Pod logs.

{{< alert color="success" title="Command and Examples" >}}
See the [Reference](/docs/reference/) section for the [commonLabels](/docs/reference/api/kustomization-file/commonlabels/) command and examples.
{{< /alert >}}

## Setting Annotations
Setting Annotations is very similar to setting labels as describe above. In addition to updating the annotations for each Resource, any fields that contain ObjectMetadata, such as `PodTemplate`, will also have the annotations added.

{{< alert color="success" title="Command and Examples" >}}
See the [Reference](/docs/reference/) section for the [commonAnnotations](/docs/reference/api/kustomization-file/commonannotations/) command and examples.
{{< /alert >}}
