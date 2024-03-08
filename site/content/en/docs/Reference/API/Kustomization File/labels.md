---
title: "labels"
linkTitle: "labels"
type: docs
weight: 10
description: >
    Add labels and optionally selectors to all resources.
---
`apiVersion: kustomize.config.k8s.io/v1beta1`

See the [Tasks section] for examples of how to use `labels`.

### labels
Adds labels and optionally selectors to all resources.

* **labels** ([]Label)

    List of labels and label selector options.

    _Label holds labels to add to resources and options for customizing how those labels are applied, potentially using selectors and template metadata._

    * **pairs** (map[string]string)

        Map of labels that the transformer will add to resources.

    * **includeSelectors** (bool), optional

        IncludeSelectors indicates whether the transformer should include the fieldSpecs for selectors. Custom fieldSpec specified by `fields` will be merged with builtin fieldSpecs if this is true. Defaults to false.

    * **includeTemplates** (bool), optional

        IncludeTemplates indicates whether the transformer should include the `spec/template/metadata` fieldSpec. Custom fieldSpecs specified by `fields` will be merged with the `spec/template/metadata` fieldSpec if this is true. If IncludeSelectors is true, IncludeTemplates is not needed. Defaults to false.

    * **fields** (\[\][FieldSpec]({{< relref "../Common%20Definitions/FieldSpec.md" >}})), optional

        Fields specifies the field on each resource that LabelTransformer should add the label to. It essentially allows the user to re-define the field path of the Kubernetes labels field from `metadata/labels` for different resources.


[Tasks section]: /docs/tasks/labels_and_annotations/
[Labels and Selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
