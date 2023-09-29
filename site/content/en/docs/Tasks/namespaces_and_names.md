---
title: "Namespaces and Names"
linkTitle: "Namespaces and Names"
weight: 3
date: 2023-09-29
description: >
  Working with Namespaces and Names
---

{{< alert color="success" title="TL;DR" >}}
- Set the Namespace for all Resources within a Project with `namespace`
- Prefix the Names of all Resources within a Project with `namePrefix`
- Suffix the Names of all Resources within a Project with `nameSuffix`
{{< /alert >}}

## Motivation
It may be useful to enforce consistency across the namespace and names of all Resources within a project.

Typical use cases include:
- Ensure all Resources are in the correct Namespace
- Ensure all Resources share a common naming convention
- Copy or fork an existing project and change the Namespace or Resource name

## Setting Namespace
The Namespace for all namespaced Resources declared in the Resource Config may be set with `namespace`. This sets the Namespace for both generated Resources (e.g. ConfigMaps and Secrets) and non-generated Resources within a project.

{{< alert color="success" title="Command and Examples" >}}
See the [Reference](/docs/reference/) section for the [namespace](/docs/reference/api/kustomization-file/namespace/) command and examples.
{{< /alert >}}

## Setting Resource Name prefix and suffix
The prefix or suffix value may be added to the name of all generated Resources (e.g. ConfigMaps and Secrets) and non-generated Resources within a project.

The name prefix and suffix is propagated to references to the Resource. Common scenarios include:
- Service references from StatefulSets
- ConfigMap references from PodSpecs
- Secret references from PodSpecs

{{< alert color="success" title="Command and Examples" >}}
See the [Reference](/docs/reference/) section for the [namePrefix](/docs/reference/api/kustomization-file/nameprefix/) and [nameSuffix](/docs/reference/api/kustomization-file/namesuffix/) commands and examples.
{{< /alert >}}
