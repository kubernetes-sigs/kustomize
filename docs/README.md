English | [简体中文](zh/README.md)

# Documentation

 * [Installation](INSTALL.md)

 * [Examples](../examples) - detailed walkthroughs of various
    workflows and concepts.

 * [Glossary](glossary.md) - the word of the day is [_root_](glossary.md#kustomization-root).

 * [Kustomize Fields](fields.md) - explanations of the fields
   in a  [kustomization](glossary.md#kustomization) file.

 * [Plugins](plugins) - extending kustomize with
   custom generators and transformers.

 * [Workflows](workflows.md) - steps one might take in
   using bespoke and off-the-shelf configurations.

 * [FAQ](FAQ.md)


## Release notes

 * [kustomize/3.2.2](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv3.2.2) - kustomize CLI
   moved to depend on kustomize Go API [3.3.0](v3.3.0.md).

 * [API 3.3.0](v3.3.0.md) - First release of the kustomize Go API
   in a module excluding the `kustomize` CLI.  From here on,
   the CLI and API will release independently.
 
 * [kustomize/3.2.1](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv3.2.1) - Patch release
   of `kustomize` CLI in its own module,
   depending on Go API release [3.2.0](v3.2.0.md).

 * [3.2.0](v3.2.0.md) - TODO(jingfang)

 * [3.1.1](v3.1.0.md) - TODO(jingfang)

 * [3.1](v3.1.0.md) - Late July 2019. Extended patches and improved resource matching.

 * [3.0](v3.0.0.md) - Late June 2019. Plugin developer release.

 * [2.1](v2.1.0.md) - 18 June 2019.  Plugins, ordered resources, etc.

 * [2.0](v2.0.0.md) - Mar 2019.
   kustomize [v2.0.3] is available in [kubectl v1.14][kubectl].

 * [1.0](v1.0.1.md) - May 2018.  Initial release after development
   in the [kubectl repository].


## Policies

 * [Versioning](versioningPolicy.md) - how the code and
   the kustomization file evolve in time.

 * [Eschewed features](eschewedFeatures.md) - why certain features
   are (currently) not supported in kustomize.

 * [Contributing guidelines](../CONTRIBUTING.md) - please read
   before sending a PR.

 * [Code of conduct](../code-of-conduct.md)

[v2.0.3]: https://github.com/kubernetes-sigs/kustomize/releases/tag/v2.0.3
[kubectl]: https://kubernetes.io/blog/2019/03/25/kubernetes-1-14-release-announcement
[kubectl repository]: https://github.com/kubernetes/kubectl
