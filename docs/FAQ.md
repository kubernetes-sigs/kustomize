# FAQ

## security: file 'foo' is not in or below 'bar'

v2.0 added a security check that prevents
kustomizations from reading files outside their own
directory root.

This was meant to help protect the person inclined to
download kustomization directories from the web and use
them without inspection to control their production
cluster (see [#693](https://github.com/kubernetes-sigs/kustomize/issues/693)).

Resources (including configmap and secret generators)
can _still be shared_ via the recommended best practice
of placing them in a directory with their own
kustomization file, and refering to this directory as a
[`base`](glossary.md#base) from any kustomization that
wants to use it.  This encourages modularity and
relocatability.

At the moment (in v2.0.3), however, there's no
(released) analogous way to share patch files and other
transformer configuration data between kustomizations.

As a stop-gap until we add base-like behavior for
transformers, we've added a flag to disable the check:


```
kustomize build --load_restrictor none $target
```

This flag is not in v2.0.3, but is available from head
(`go install sigs.k8s.io/kustomize`).
