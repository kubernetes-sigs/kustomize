---
title: "Writing Code"
linkTitle: "Writing Code"
type: docs
weight: 40
description: >
    Background information to help orient newcomers to the Kustomize codebase
---

{{< alert color="success" title="Workspace mode" >}}
Kustomize supports go workspace mode, so if adding locally modified modules, remember to also add it to the [go.work](https://github.com/kubernetes-sigs/kustomize/blob/master/go.work) file.

{{< /alert >}}

#### Build command

All kustomize commands can be found in [kustomize/commands](https://github.com/kubernetes-sigs/kustomize/tree/master/kustomize/commands)

Running `kustomize build` triggers the [NewCmdBuild](https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/commands/build/build.go#L65), which does the following:

- creates a [Kustomizer](https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/commands/build/build.go#L77) instance

- calls [Run](https://github.com/kubernetes-sigs/kustomize/blob/482e8930fc256672afd4ff5d531ec8fe80d35119/api/krusty/kustomizer.go#L53-L543)  to perform the kustomization

- returns the [kustomization output as yaml](https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/commands/build/build.go#L89-L97)

