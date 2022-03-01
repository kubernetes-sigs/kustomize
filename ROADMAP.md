# Kustomize roadmap 2022

Presented at the [January 26, 2022, SIG-CLI meeting](https://youtu.be/l2plzJ9MRlk?t=1321)

kustomize maintainers: @knverey, @natasha41575

[Objective: Improve contributor community](#objective-improve-contributor-community)

[Objective: Improve end-user experience](#objective-improve-end-user-experience)

[Objective: Improve extension experience](#objective-improve-extension-experience)

## Objective: Improve contributor community

**_WHO: End user who also contributes source code._**

Top priority:

- Kustomization v1 (also end-user impact) ([PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/12))
  - Remove the following fields:
    - [vars](https://github.com/kubernetes-sigs/kustomize/issues/2052)
    - [patchesJson6902, patchesStrategicMerge (consolidate on \`patches)](https://github.com/kubernetes-sigs/kustomize/issues/4376)
    - [helmChartInflationGenerator, helmCharts, helmGlobals](https://github.com/kubernetes-sigs/kustomize/issues/4401)
    - all long-deprecated fields in Kustomization v1 such as \`bases\` and those being accommodate by kustomize edit \[[see code snippet](https://github.com/kubernetes-sigs/kustomize/blob/ee4b7847f0beb6c0d2070673b10f23f7b3e92e82/api/types/fix.go#L15)\]
  - Ensure that \`kustomize edit fix\` handles migrations for all those, and that anything it changes is not still present in v1.
  - [Add reorder field](https://github.com/kubernetes-sigs/kustomize/issues/3913). Default should be FIFO and legacy should also be supported (could add alphabetic and custom sort support eventually). Replaces -reorder flag.
  - [Reconcile openapi and crds field](https://github.com/kubernetes-sigs/kustomize/issues/3944)
  - [Consider deprecating configurations field](https://github.com/kubernetes-sigs/kustomize/issues/3945) (old, pre-plugin, pre-openapi global configuration)
  - [Add a field to enable the managedby label](https://github.com/kubernetes-sigs/kustomize/issues/4047)

Second priority:

- Improve contributor documentation
  - [Instructions to upgrade kustomize-in-kubectl](https://github.com/kubernetes-sigs/kustomize/issues/3951)

Also very valuable to the project:

- [Improve the release process](https://github.com/kubernetes-sigs/kustomize/issues/3952) to support regular biweekly releases [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/7)
- Release sigs.k8s.io/kustomize/api v1.0.0 [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/5)
  - [Reduce the public surface of the API module](https://github.com/kubernetes-sigs/kustomize/issues/3942)
  - [Vendor all transitive deps](https://github.com/kubernetes-sigs/kustomize/issues/3706). Since kustomize is in kubectl, we must do as kubectl does to manage deps, exposing new transitive deps in code review.
- Project administration
  - [Rename master branch to main](https://github.com/kubernetes-sigs/kustomize/issues/3954)



## Objective: Improve end-user experience

**_WHO: End user that wants kustomize build artifacts (binaries, containers)._**

Top priorities:

- Bug fixes:
  - Fix bugs in basic anchor support: [issue query](https://github.com/kubernetes-sigs/kustomize/issues?q=is%3Aopen+is%3Aissue+label%3Aarea%2Fanchors)
  - integer keys support: [#3446](https://github.com/kubernetes-sigs/kustomize/issues/3446)
  - kyaml not respecting \`$patch replace|retainKeys\`: [#2037](https://github.com/kubernetes-sigs/kustomize/issues/2037)
  - kustomize removing quotes from namespace field values: [#4146](https://github.com/kubernetes-sigs/kustomize/issues/4146)
  - Kustomize doesn’t support metadata.generateName: [#641](https://github.com/kubernetes-sigs/kustomize/issues/641)
- Send kustomize CLI version number into kubectl ([kubectl issue](https://github.com/kubernetes/kubectl/issues/797) / [kustomize issue](https://github.com/kubernetes-sigs/kustomize/issues/1424))
- Kustomize performance investigations/improvements [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/13)
- [Support generic resource references in name reference tracking](https://github.com/kubernetes-sigs/kustomize/issues/3418)
- [KEP 4267: retain the resource origin and transformer data in annotations](https://github.com/kubernetes-sigs/kustomize/pull/4267)

Secondary priorities:

- kustomize cli v5 ([PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/14))
  - [Drop the --reorder flag](https://github.com/kubernetes-sigs/kustomize/issues/3947)
  - [Graduate cfg read-only commands out of alpha](https://github.com/kubernetes-sigs/kustomize/issues/4090).
  - [Drop the –enable-managedby-label](https://github.com/kubernetes-sigs/kustomize/issues/4047)
  - Drop old plugin-related fields in favor of [the Catalog-style fields](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/2906-kustomize-function-catalog).
  - [Drop the helm flags](https://github.com/kubernetes-sigs/kustomize/issues/4401)
- [Confusion around namespace replacement](https://github.com/kubernetes-sigs/kustomize/issues/880).

Also very valuable to the project:

- [Overinclusion of root directory error in error messages](https://github.com/kubernetes-sigs/kustomize/issues/4348)
- [Add kustomize localize command](https://github.com/kubernetes-sigs/kustomize/issues/3980)
- [Fix Windows support in test suite](https://github.com/kubernetes-sigs/kustomize/issues/4001)
- Improve end-user documentation [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/9)


## Objective: Improve extension experience

**_WHO: Plugin developers: end users who extend kustomize, but don’t think about internals._**

This objective is described in detail in the [Kustomize Plugin Graduation KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/2953-kustomize-plugin-graduation) / [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/15) .

Top priorities:

- Fix core usability issues with KRM Function extensions:
  - [Better errors for function config failures](https://github.com/kubernetes-sigs/kustomize/issues/4398)
  - [Container KRM Mounts are not mounting via function parameters](https://github.com/kubernetes-sigs/kustomize/issues/4290)
  - [Resolution of local file references in extensions transformer configuration](https://github.com/kubernetes-sigs/kustomize/issues/4154)
  - [Do not silently ignore plugins when config has typo](https://github.com/kubernetes-sigs/kustomize/issues/4399)
  - [KRM Exec Function can't locate executable when referencing a base](https://github.com/kubernetes-sigs/kustomize/issues/4347)
- Once core usability issues are fixed, [deprecate legacy exec and Go plugin support](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/2953-kustomize-plugin-graduation)
- [Catalog KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/2906-kustomize-function-catalog)

Secondary priorities:

- [Remove Starlark support](https://github.com/kubernetes-sigs/kustomize/issues/4349)
- [Composition KEP](https://github.com/kubernetes/enhancements/pull/2300). The implementation is complete in [#4223](https://github.com/kubernetes-sigs/kustomize/pull/4323), but depends on:
  - [Convert resources and components to be backed by a reusable generator](https://github.com/kubernetes-sigs/kustomize/issues/4402)
  - [Enable explicitly invoked transformers to use default fieldSpecs](https://github.com/kubernetes-sigs/kustomize/issues/4404)
  - [Enable built-in generators to be used in the transformers field ](https://github.com/kubernetes-sigs/kustomize/issues/4403)


Also very valuable to the project:

- [Improve docs for kyaml libraries](https://github.com/kubernetes-sigs/kustomize/issues/3950), especially by adding examples.
- [Create a reserved field for plugin runtime information](https://github.com/kubernetes-sigs/kustomize/issues/4405)
- [Develop new standard process for implementing builtin transformers](https://github.com/kubernetes-sigs/kustomize/issues/4400)
