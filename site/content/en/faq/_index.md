---
title: "FAQ"
linkTitle: "FAQ"
type: docs
menu:
  main:
    weight: 70
---

## kubectl doesn't have the latest kustomize, when will it be updated?

TLDR: This is blocked on either moving kubectl into its own repo, or changing its dependencies.  ETA k8s ~1.20.

The adoption of go modules in the kubernetes/kubernetes repo broke the update process for kustomize.
This is due to the kustomize libraries depending on the kubernetes apimachinery libraries, which are
published out of the kubernetes staging directory.

2 pieces of work are underway which will allow kustomize to be updated in kubectl:

- migrating kubectl out of kubernetes/kubernetes (expected Kubernetes ~1.20)
- migrating kustomize off of the apimachinery libraries (expected Kuberntes ~1.20)
  - [2506](https://github.com/kubernetes-sigs/kustomize/issues/2506)

Once either of these issues is resolved we will then update kubectl with the latest kustomize version.

## security: file 'foo' is not in or below 'bar'

v2.0 added a security check that prevents
kustomizations from reading files outside their own
directory root.

This was meant to help protect the person inclined to
download kustomization directories from the web and use
them without inspection to control their production
cluster
(see [#693](https://github.com/kubernetes-sigs/kustomize/issues/693),
[#700](https://github.com/kubernetes-sigs/kustomize/pull/700),
[#995](https://github.com/kubernetes-sigs/kustomize/pull/995) and
[#998](https://github.com/kubernetes-sigs/kustomize/pull/998))

Resources (including configmap and secret generators)
can _still be shared_ via the recommended best practice
of placing them in a directory with their own
kustomization file, and referring to this directory as a
[`base`](/kustomize/api-reference/glossary#base) from any kustomization that
wants to use it.  This encourages modularity and
relocatability.

To disable this, use v3, and the `load_restrictor` flag:

```
kustomize build --load_restrictor none $target
```

## Some field is not transformed by kustomize

Example: [#1319](https://github.com/kubernetes-sigs/kustomize/issues/1319), [#1322](https://github.com/kubernetes-sigs/kustomize/issues/1322), [#1347](https://github.com/kubernetes-sigs/kustomize/issues/1347) and etc.

The fields transformed by kustomize is configured explicitly in [defaultconfig](https://github.com/kubernetes-sigs/kustomize/tree/master/api/konfig/builtinpluginconsts/defaultconfig.go). The configuration itself can be customized by including `configurations` in `kustomization.yaml`, e.g.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configurations:
- kustomizeconfig.yaml
```

The configuration directive allows customization of the following transformers:

```yaml
commonAnnotations: []
commonLabels: []
nameprefix: []
namespace: []
varreference: []
namereference: []
images: []
replicas: []
```

To persist the changes to default configuration, submit a PR like [#1338](https://github.com/kubernetes-sigs/kustomize/pull/1338), [#1348](https://github.com/kubernetes-sigs/kustomize/pull/1348) and etc.
