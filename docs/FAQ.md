# FAQ

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
kustomization file, and refering to this directory as a
[`base`](glossary.md#base) from any kustomization that
wants to use it.  This encourages modularity and
relocatability.

To disable this, use v3, and the `load_restrictor` flag:

```
kustomize build --load_restrictor none $target
```
