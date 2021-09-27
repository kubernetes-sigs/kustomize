# Kustomize roadmap 2021 Q3/4

Presented at the June 2, 2021 SIG-CLI meeting: [recording](https://youtu.be/YUNSIxWlOPA)

kustomize maintainers: @knverey, @monopole

[Objective: Improve contributor community](#objective-improve-contributor-community)

[Objective: Improve end-user experience](#objective-improve-end-user-experience)

[Objective: Improve extension experience](#objective-improve-extension-experience)



## Objective: Improve contributor community

**WHO: End user who also contributes source code.**

- Publish (this) roadmap.
- Release **sigs.k8s.io/kustomize/api** **v1.0.0** ([PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/5))
  - [Fewer public packages](https://github.com/kubernetes-sigs/kustomize/issues/3942):
    - builtins (generated code from legacy plugins): Internal
    - hasher (for secret and configmap generation): Internal, or maybe kyaml
    - ifc (single file with interfaces): Internal except maybe KustHasher and Validator
    - Image (utils for parsing image field): Move to image transformer
    - konfig: internalkrusty: keep external but rename it to e.g. “build” or “runner”
    - kv: internal
    - loader: internal
    - provenance: probably needs to stay external for the linker, but confirm
    - provider: internal
    - resmap: needs to stay external as long as the go plugins exist
    - testutils: try to make internal, but not a big deal
  - [Vendor all transitive deps](https://github.com/kubernetes-sigs/kustomize/issues/3706).
    Since kustomize is in kubectl, we must do as kubectl does to manage deps, exposing new transitive deps in code review.
  - [Remove starlark dependencies](https://github.com/kubernetes-sigs/kustomize/issues/3943) (but make it possible to re-enable via injection)
- **Kustomization v1 (also end-user impact)** ([PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/12))
  - [Deprecate `vars` field](https://github.com/kubernetes-sigs/kustomize/issues/2052). Allowed but no longer honored. This includes:
    - merging and documenting the [ReplacementTransformer](https://github.com/kubernetes-sigs/kustomize/issues/3492)
    - [supporting vars→replacements migration in `kustomize edit fix`](https://github.com/kubernetes-sigs/kustomize/issues/3849), with documentation
    - emitting warnings on stderr pointing users to the migration plan
  - [Deprecate `crds` field](https://github.com/kubernetes-sigs/kustomize/issues/3944) with migration instructions to “openapi"
  - [Add `reorder` field](https://github.com/kubernetes-sigs/kustomize/issues/3913). Default should be FIFO and legacy should also be supported (could add alphabetic and custom sort support eventually). Replaces -reorder flag.
  - [Consider deprecating `configurations` field](https://github.com/kubernetes-sigs/kustomize/issues/3945) (old, pre-plugin, pre-openapi global configuration)
- Improve contributor documentation
  - Clarify [contributor docs](https://kubectl.docs.kubernetes.io/contributing/kustomize/features/) regarding when discussions should happen on issues vs. brought to a SIG meeting
  - [Add a top-level ARCHITECTURE.md document](https://github.com/kubernetes-sigs/kustomize/pull/3924).
  - [Improve docs for kyaml libraries](https://github.com/kubernetes-sigs/kustomize/issues/3950), especially by adding examples.
  - [Instructions to upgrade kustomize-in-kubectl](https://github.com/kubernetes-sigs/kustomize/issues/3951) (how to make a PR like [this one](https://github.com/kubernetes/kubernetes/pull/101120)). The description of that PR contains the instructions that worked on that particular date. Let this happen frequently
- [Improve the release process](https://github.com/kubernetes-sigs/kustomize/issues/3952) to support regular biweekly releases [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/7)
  - More automation - it has toil because of the multiple modules.
  - More maintainers with authorization to do releases (e.g. create branches, delete mistake branches, move tags, trigger builds, etc.).
  - Include an upgrade PR to kubectl in the release process.
  - Send kustomize CLI version number into kubectl ([kubectl issue](https://github.com/kubernetes/kubectl/issues/797) / [kustomize issue](https://github.com/kubernetes-sigs/kustomize/issues/1424))
- Project administration
  - DONE Clean up Kustomize-related [KEPs](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli) to reflect current status
  - DONE Onboard Kustomize repo to SIG-CLI’s Triage Party for bug scrubs
  - [Rename master branch to main](https://github.com/kubernetes-sigs/kustomize/issues/3954)



## Objective: Improve end-user experience

**WHO: End user that wants kustomize build artifacts (binaries, containers).**

- Improve end-user documentation [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/9)
  - Easily discovered [tutorials](https://github.com/kubernetes-sigs/kustomize/issues/3973) for folks who just want to get the job done.
    Implicit best practice recommendation.
  - Usage [tutorial](https://github.com/kubernetes-sigs/kustomize/issues/3973) videos
  - Clear explanation of differing capabilities of kustomize standalone vs. kustomize in kubectl. [Document kubectl kustomize integration #3951](https://github.com/kubernetes-sigs/kustomize/issues/3951)
  - [The kustomize doc site should pull examples from the kustomize repo](https://github.com/kubernetes-sigs/kustomize/issues/3974)
- [Telemetry to guide investment decisions](https://github.com/kubernetes-sigs/kustomize/issues/3941) ([PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/13))
  -  Idea: Resurrect "crawler" to gather stats on popularity of features so we know where to focus.
  -  Idea: Allow opt-in to build data (Kustomize file sent to some server at build time)
  -  A battery of [Go benchmark tests](https://github.com/kubernetes-sigs/kustomize/issues/3248) to monitor performance. We've had reports of performance regression, but no measures *in-repo*.
- **kustomize cli v5** ([PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/14))
  - [Indentation changes from go-yaml](https://github.com/kubernetes-sigs/kustomize/issues/3946), if we cannot prevent them in kubectl kustomize. (also related to a [kyaml feature request](https://github.com/kubernetes-sigs/kustomize/issues/3559))
  - [Drop the `--reorder` flag](https://github.com/kubernetes-sigs/kustomize/issues/3947)
  - [Decide what to do with `kustomize cfg`](https://github.com/kubernetes-sigs/kustomize/issues/3953)
    - Drop 'cfg create setter' and 'cfg set' and delete setter code from Kustomize (it lives in cmd/config).
    - Keep the cfg read-only commands: grep, cat, tree.
  - The --enable-**alpha**-plugins flag will remain in this version: this only changes when plugins graduate.
- Features
  - [Kustomize Components KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/1802-kustomize-components)
    - Needs end-user docs
  - [OpenAPI KEP](https://github.com/kubernetes-sigs/kustomize/issues/3723)
     -  Should deprecate[ the existing `crds` field](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/crds/)
  - [Binary releases for M1 (darwin_arm64)](https://github.com/kubernetes-sigs/kustomize/issues/3736)
  - [confusion around namespace replacement](https://github.com/kubernetes-sigs/kustomize/issues/880) - this is the most +1’d feature request and is two years old
  - Release version 1.0 of sigs.k8s.io/kustomize/api for incorporating kustomize into other tools, as well as v1 of Kustomization. [see details under Contributor Community]
  - git caching scheme (behind a flag) https://github.com/kubernetes-sigs/kustomize/pull/3655
  - hashicorp go-getter support via dependency injection or [external plugin](https://github.com/kubernetes-sigs/kustomize/issues/3585)



## Objective: Improve extension experience

**WHO: Plugin developers: end users who extend kustomize, but don’t think about internals.**

- Replace `Resource` fields with annotations so that the data is visible to and survives across KRM function boundaries:
  - [Replace Resource.options with annotations #3975](https://github.com/kubernetes-sigs/kustomize/issues/3975)
  - [Replace Resource.refBy with annotations #3976](https://github.com/kubernetes-sigs/kustomize/issues/3976)
- Create KEP for plugin [graduation](https://github.com/kubernetes-sigs/kustomize/issues/2721) out of alpha. [PROJECT](https://github.com/kubernetes-sigs/kustomize/projects/15)
  Proposals for this KEP:
  - Implement the [Composition KEP](https://github.com/kubernetes/enhancements/pull/2300) .
  - Deprecation story
    - Deprecate shared Go libs. (if you use a Kustomization file that mentioned a plugin that must be loaded as a .so file, it fails). Although these have advantages--notably in terms of execution speed and ease of testing--they aren't portable.
    - Deprecate the legacy path lookup exec plugin style in favor of the KRM style that also already exists.
    - Add deprecation warnings if someone is using legacy plugins. See if we get complaints. Upon removal, add a fix command if feasible or at minimum a pointer to migration docs.
  - Create a roadmap for plugin inclusion in kustomize-in-kubectl, or explicitly document that plugins will never be included (and why).
  - Determine whether plugins need to be permanently gated with a flag. Remove the existing flag, and replace it with a non-alpha flag if necessary.
  - Deprecate KRM plugin configuration options that promote violations of Kustomize’s policy that everything required for a build should be committed (no side-effects from env, cli flags, etc). All plugin config should be in the KRM config object for the plugin.
    - Starlark plugins: Deprecate generic URL download for starlark plugins, replacing it with git-specific functionality in line with Kustomization’s own git URL support. Subject the relative path to loader restrictions.
    - Container plugins: Deprecate network access, storage mount and env options.
    - Exec plugins: Subject the exec path to loader restrictions.
  - Move KRM plugin provider specification from an annotation to a reserved field.
  - Add support for content-addressable OCI artifact storage for starlark and exec plugin providers.
  - Implement a Catalog KRM resource that uses [OCI Artifacts](https://github.com/opencontainers/artifacts) ([maybe](https://jzelinskie.com/posts/oci-artifacts/)) to improve plugin distribution and enable end users to bypass specifying explicit plugin provider versions in their KRM plugin configs. This likely deserves a KEP on its own.
  - Develop a streamlined contribution process for new “builtin” functionality, i.e. built-in transformers that essentially wrap kyaml filters.
- Add extensive docs, tests and examples for all plugin mechanisms that will remain supported after graduation.
