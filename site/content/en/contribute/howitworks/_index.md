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

## Call stack
Call stack when running `kustomize build`, with links to code:

#### Run build

* [Build Command](https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/commands/build/build.go#L65)
  * [Validate](https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/commands/build/build.go#L114)
  * [MakeKustomizer](https://github.com/kubernetes-sigs/kustomize/blob/master/api/krusty/kustomizer.go#L36)
  * [Run](https://github.com/kubernetes-sigs/kustomize/blob/master/api/krusty/kustomizer.go#L53): performs a kustomization. It uses its internal filesystem reference to read the file at the given path argument, interpret it as a `kustomization.yaml` file, perform the kustomization it represents, and returns the resulting resources.
    * Create factories
      * [depprovider.GetFactory](https://github.com/kubernetes-sigs/kustomize/blob/master/api/provider/depprovider.go#L36)
      * [resmap.NewFactory](https://github.com/kubernetes-sigs/kustomize/blob/master/api/resmap/factory.go#L22)
        * [resource.Factory](https://github.com/kubernetes-sigs/kustomize/blob/master/api/resource/factory.go#L23)
      * [loader.NewLoader](https://github.com/kubernetes-sigs/kustomize/blob/master/api/loader/loader.go#L20)
    * [NewKustTarget](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L42)
      * [Load](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L58)
      * [Kustomization](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L85)
    * [MakeCustomizeResMap](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L117): details in [next section](#make-resource-map)
  * [write output to yaml file](https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/commands/build/build.go#L89-#L97)

### Make resource map

* [makeCustomizeResMap](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L121)
  * [AccumulateTarget](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L179): returns a new ResAccumulator, holding customized resources and the data/rules used to do so. The name back references and vars are not yet fixed.
    * [accummulateResources](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L404): fills the given resourceAccumulator with resources read from the given list of paths.
    * [accumulateComponents](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L445)
    * Merge config from builtin and CRDs
    * [runGenerators](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L253)
      * [configureBuiltinGenerators](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget_configplugin.go#L34)
        * ConfigMapGenerator
        * SecretGenerator
        * HelmChartInflationGenerator
      * [configureExternalGenerators](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L286)
        * Iterate all generators
    * [runTransfomers](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L317)
      * [configureBuiltinTransformers](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget_configplugin.go#L69)
        * PatchStrategicMergeTransformer
        * PatchTransformer
        * NamespaceTransformer
        * PrefixSuffixTransformer
        * SuffixTransformer
        * LabelTransformer
        * AnnotationsTransformer
        * PatchJson6902Transformer
        * ReplicaCountTransformer
        * ImageTagTransformer
        * ReplacementTransformer
      * [configureExternalTransformers](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L333)
      * [runValidators](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L364)
    * [MergeVars](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/accumulator/resaccumulator.go#L71)
    * [IgnoreLocals](https://github.com/kubernetes-sigs/kustomize/blob/master/api/internal/target/kusttarget.go#L241)
  * The following steps must be done last, not as part of the recursion implicit in AccumulateTarget.
    * [addHashesToNames](https://github.com/kubernetes-sigs/kustomize/blob/2f2ba40876b3b6c5b33281e8dee503010a1bc537/api/internal/target/kusttarget.go#L156)
    * [FixBackReferences](https://github.com/kubernetes-sigs/kustomize/blob/2f2ba40876b3b6c5b33281e8dee503010a1bc537/api/internal/accumulator/resaccumulator.go#L163): Given that names have changed (prefixs/suffixes added), fix all the back references to those names.
    * [ResolveVars](https://github.com/kubernetes-sigs/kustomize/blob/2f2ba40876b3b6c5b33281e8dee503010a1bc537/api/internal/accumulator/resaccumulator.go#L144)
