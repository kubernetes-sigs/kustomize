---
title: "Container Images"
linkTitle: "Container Images"
weight: 2
date: 2023-09-29
description: >
  Working with Container Images
---

{{< alert color="success" title="TL;DR" >}}
- Set the container image name and tag
{{< /alert >}}

## Motivation
It may be useful to define or override the container image, tag, or digest which are used across many Workloads.

Container image tags and digests are used to refer to a specific version or instance of a container image - e.g. for the nginx container image you might use the tag 1.15.9 or 1.14.2.

Typical uses cases include:
- Update the container image name or tag for multiple Workloads at once
- Increase visibility of the versions of container images being used within the project
- Set the image tag from external sources - such as environment variables
- Copy or Fork an existing Project and change the Image Tag for a container
- Change the registry used for an image

{{< alert color="success" title="Command and Examples" >}}
See the [Reference](/docs/reference/) section for the [images](/docs/reference/api/kustomization-file/images/) command and examples.
{{< /alert >}}
