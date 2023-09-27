---
title: "Container Images"
linkTitle: "Container Images"
weight: 2
date: 2017-01-05
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

See the [Reference](/docs/reference/) section for [examples](/docs/reference/api/kustomization-file/images/) of working with container images.
