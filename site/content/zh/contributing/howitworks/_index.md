---
title: "Writing Code"
linkTitle: "Writing Code"
type: docs
weight: 40
description: >
    How to modify Kustomize
---

{{% pageinfo color="info" %}}
To build kustomize using the locally modified modules, `replace` statements must be added to
the `kustomize/go.mod`.

e.g. if code in `api` was modified, a `replace` statement would need to be added for the
`kustomize/api` module.
{{% /pageinfo %}}

Call stack when running `kustomize build`, with links to code.

## Run build

* [RunBuild](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/kustomize/internal/commands/build/build.go#L121)
  * [MakeKustomizer](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/krusty/kustomizer.go#L32)
  * [Run](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/krusty/kustomizer.go#L47): performs a kustomization. It uses its internal filesystem reference to read the file at the given path argument, interpret it as a kustomization.yaml file, perform the kustomization it represents, and return the resulting resources.
    * Create factories
      * [tranformer.NewFactoryImpl](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/k8sdeps/transformer/factory.go#L17)
      * [resmap.NewFactory](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/resmap/factory.go#L21)
        * [resource.NewFactory](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/resource/factory.go#L23)
          * [kustruct.NewKunstructuredFactoryImpl](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/k8sdeps/kunstruct/factory.go#L28)
      * [loader.NewLoader](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/loader/loader.go#L19)
      * [validator.NewKustValidator](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/k8sdeps/validator/validators.go#L23)
    * [NewKustTarget](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L38)
    * [Load](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L54)
    * [MakeCustomizeResMap](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L109): details in next section
  * [emitResources](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/kustomize/internal/commands/build/build.go#L143)

## Make resource map

* [makeCustomizeResMap](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L117)
  * [AccumulateTarget](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L196): returns a new ResAccumulator, holding customized resources and the data/rules used to do so. The name back references and vars are not yet fixed.
    * [accummulateResources](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L302): fills the given resourceAccumulator with resources read from the given list of paths.
    * Merge config from builtin and CRDs
    * [runGenerators](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L239)
      * [configureBuiltinGenerators](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget_configplugin.go#L28)
        * ConfigMapGenerator
        * SecretGenerator
      * [configureExternalGenerators]()
      * Iterate all generators
    * [runTransfomers](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L274)
      * [configureBuiltinTransformers](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget_configplugin.go#L44)
        * PatchStrategicMergeTransformer
        * PatchTransformer
        * NamespaceTransformer
        * PrefixSuffixTransformer
        * LabelTransformer
        * AnnotationsTransformer
        * PatchJson6902Transformer
        * ReplicaCountTransformer
        * ImageTagTransformer
      * [configureExternalTransformers](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L291)
    * [MergeVars](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/accumulator/resaccumulator.go#L64)
  * The following steps must be done last, not as part of the recursion implicit in AccumulateTarget.
    * [addHashesToNames](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/target/kusttarget.go#L153)
    * [FixBackReferences](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/accumulator/resaccumulator.go#L160): Given that names have changed (prefixs/suffixes added), fix all the back references to those names.
    * [ResolveVars](https://github.com/kubernetes-sigs/kustomize/blob/c7d78970fb86782dbdded3a93944b774f826071f/api/internal/accumulator/resaccumulator.go#L141)
