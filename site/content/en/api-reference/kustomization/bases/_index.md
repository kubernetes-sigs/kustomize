---
title: "bases"
linkTitle: "bases"
type: docs
description: >
    Add resources from a kustomization dir.
---

{{% pageinfo color="warning" %}}
The `bases` field was deprecated in v2.1.0
{{% /pageinfo %}}

Move entries into the [resources](/kustomize/api-reference/kustomization/resources)
field.  This allows bases - which are still a
[central concept](/kustomize/api-reference/glossary#base) - to be
ordered relative to other input resources.
