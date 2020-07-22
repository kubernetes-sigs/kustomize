---
title: "Kustomize Plugins"
linkTitle: "Kustomize Plugins"
type: docs
weight: 98
description: >
    Kustomize plugins guide
---

Kustomize offers a plugin framework allowing people to write their own resource _generators_
and _transformers_.

[generator options]: https://github.com/kubernetes-sigs/kustomize/tree/master/examples/generatorOptions.md
[transformer configs]: https://github.com/kubernetes-sigs/kustomize/tree/master/examples/transformerconfigs

Write a plugin when changing [generator options]
or [transformer configs] doesn't meet your needs.

[12-factor]: https://12factor.net

* A _generator_ plugin could be a helm chart
   inflator, or a plugin that emits all the
   components (deployment, service, scaler,
   ingress, etc.) needed by someone's [12-factor]
   application, based on a smaller number of free
   variables.

* A _transformer_ plugin might perform special
   container command line edits, or any other
   transformation beyond those provided by the
   builtin (`namePrefix`, `commonLabels`, etc.)
   transformers.

## Specification in `kustomization.yaml`

Start by adding a `generators` and/or `transformers`
field to your kustomization.

Each field accepts a string list:

> ```
> generators:
> - relative/path/to/some/file.yaml
> - relative/path/to/some/kustomization
> - /absolute/path/to/some/kustomization
> - https://github.com/org/repo/some/kustomization
>
> transformers:
> - {as above}
> ```

The value of each entry in a `generators` or
`transformers` list must be a relative path to a
YAML file, or a path or URL to a [kustomization].
This is the same format as demanded by the
`resources` field.

[kustomization]: /kustomize/api-reference/glossary#kustomization

YAML files are read from disk directly.  Paths or
URLs leading to kustomizations trigger an
in-process kustomization run.  Each of the
resulting objects is now further interpreted by
kustomize as a _plugin configuration_ object.

## Configuration

A kustomization file could have the following lines:

```
generators:
- chartInflator.yaml
```

Given this, the kustomization process would expect
to find a file called `chartInflator.yaml` in the
kustomization [root](/kustomize/api-reference/glossary#kustomization-root).

This is the plugin's configuration file;
it contains a YAML configuration object.

The file `chartInflator.yaml` could contain:

```
apiVersion: someteam.example.com/v1
kind: ChartInflator
metadata:
  name: notImportantHere
chartName: minecraft
```

__The `apiVersion` and `kind` fields are
used to locate the plugin.__

[k8s object]: /kustomize/api-reference/glossary#kubernetes-style-object

Thus, these fields are required.  They are also
required because a kustomize plugin configuration
object is also a [k8s object].

To get the plugin ready to generate or transform,
it is given the entire contents of the
configuration file.

[NameTransformer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin/prefixsuffixtransformer/PrefixSuffixTransformer_test.go
[ChartInflator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/chartinflator/ChartInflator_test.go
[plugins]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin

For more examples of plugin configuration YAML,
browse the unit tests below the [plugins] root,
e.g. the tests for [ChartInflator] or
[NameTransformer].

## Placement

Each plugin gets its own dedicated directory named

[`XDG_CONFIG_HOME`]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

```
$XDG_CONFIG_HOME/kustomize/plugin
    /${apiVersion}/LOWERCASE(${kind})
```

The default value of [`XDG_CONFIG_HOME`] is
`$HOME/.config`.

The one-plugin-per-directory requirement eases
creation of a plugin bundle (source, tests, plugin
data files, etc.) for sharing.

In the case of a [Go plugin](#go-plugins), it also
allows one to provide a `go.mod` file for the
single plugin, easing resolution of package
version dependency skew.

When loading, kustomize will first look for an
_executable_ file called

```
$XDG_CONFIG_HOME/kustomize/plugin
    /${apiVersion}/LOWERCASE(${kind})/${kind}
```

If this file is not found or is not executable,
kustomize will look for a file called `${kind}.so`
in the same directory and attempt to load it as a
[Go plugin](#go-plugins).

If both checks fail, the plugin load fails the overall
`kustomize build`.

## Execution

Plugins are only used during a run of the
`kustomize build` command.

Generator plugins are run after processing the
`resources` field (which itself can be viewed as a
generator, simply reading objects from disk).

The full set of resources is then passed into the
transformation pipeline, wherein builtin
transformations like `namePrefix` and
`commonLabel` are applied (if they were specified
in the kustomization file), followed by the
user-specified transformers in the `transformers`
field.

The order specified in the `transformers` field is
respected, as transformers cannot be expected to
be commutative.

#### No Security

Kustomize plugins do not run in any kind of
kustomize-provided sandbox.  There's no notion
of _"plugin security"_.

A `kustomize build` that tries to use plugins but
omits the flag

> `--enable_alpha_plugins`

will not load plugins and will fail with a
warning about plugin use.

The use of this flag is an opt-in acknowledging
the unstable (alpha) plugin API, the absence of
plugin provenance, and the fact that a plugin
is not part of kustomize.

To be clear, some kustomize plugin downloaded
from the internet might wonderfully transform
k8s config in a desired manner, while also
quietly doing anything the user could do to the
system running `kustomize build`.

## Authoring

There are two kinds of plugins, [exec](#exec-plugins) and [Go](#go-plugins).

### Exec plugins

A _exec plugin_ is any executable that accepts a
single argument on its command line - the name of
a YAML file containing its configuration (the file name
provided in the kustomization file).

> TODO: restrictions on plugin to allow the _same exec
> plugin_ to be targeted by both the
> `generators` and `transformers` fields.
>
> - first arg could be the fixed string
>   `generate` or `transform`,
>   (the name of the configuration file moves to
>   the 2nd arg), or
> - or by default an exec plugin behaves as a tranformer
>   unless a flag `-g` is provided, switching the
>   exec plugin to behave as a generator.

[helm chart inflator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/chartinflator
[bashed config map]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/bashedconfigmap
[sed transformer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/sedtransformer
[hashicorp go-getter]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/gogetter

#### Examples

* [helm chart inflator] - A generator that inflates a helm chart.
* [bashed config map] -  Super simple configMap generation from bash.
* [sed transformer] - Define your unstructured edits using a
   plugin like this one.
* [hashicorp go-getter] - Download kustomize layes and build it to generate resources

A generator plugin accepts nothing on `stdin`, but emits
generated resources to `stdout`.

A transformer plugin accepts resource YAML on `stdin`,
and emits those resources, presumably transformed, to
`stdout`.

kustomize uses an exec plugin adapter to provide
marshalled resources on `stdin` and capture
`stdout` for further processing.

#### Generator Options

A generator exec plugin can adjust the generator options for the resources it emits by setting one of the following internal annotations.

> NOTE: These annotations are local to kustomize and will not be included in the final output.

**`kustomize.config.k8s.io/needs-hash`**

Resources can be marked as needing to be processed by the internal hash transformer by including the `needs-hash` annotation. When set valid values for the annotation are `"true"` and `"false"` which respectively enable or disable hash suffixing for the resource. Omitting the annotation is equivalent to setting the value `"false"`.

Hashes are determined as follows:

* For `ConfigMap` resources, hashes are based on the values of the `name`, `data`, and `binaryData` fields.
* For `Secret` resources, hashes are based on the values of the `name`, `type`, `data`, and `stringData` fields.
* For any other object type, hashes are based on the entire object content (i.e. all fields).

Example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
  annotations:
    kustomize.config.k8s.io/needs-hash: "true"
data:
  foo: bar
```

**`kustomize.config.k8s.io/behavior`**

The `behavior` annotation will influence how conflicts are handled for resources emitted by the plugin. Valid values include "create", "merge", and "replace" with "create" being the default.

Example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
  annotations:
    kustomize.config.k8s.io/behavior: "merge"
data:
  foo: bar
```

### Go plugins

Be sure to read [Go plugin caveats](goplugincaveats/).

[Go plugin]: https://golang.org/pkg/plugin/

A `.go` file can be a [Go plugin] if it declares
'main' as it's package, and exports a symbol to
which useful functions are attached.

It can further be used as a _kustomize_ plugin if
the symbol is named 'KustomizePlugin' and the
attached functions implement the `Configurable`,
`Generator` and `Transformer` interfaces.

A Go plugin for kustomize looks like this:

> ```
> package main
>
> import (
> "sigs.k8s.io/kustomize/api/resmap"
>   ...
> )
>
> type plugin struct {...}
>
> var KustomizePlugin plugin
>
> func (p *plugin) Config(
>    h *resmap.PluginHelpers,
>    c []byte) error {...}
>
> func (p *plugin) Generate() (resmap.ResMap, error) {...}
>
> func (p *plugin) Transform(m resmap.ResMap) error {...}
> ```

Use of the identifiers `plugin`, `KustomizePlugin`
and implementation of the method signature
`Config` is required.

Implementing the `Generator` or `Transformer`
method allows (respectively) the plugin's config
file to be added to the `generators` or
`transformers` field in the kustomization file.
Do one or the other or both as desired.

[secret generator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/secretsfromdatabase
[service generator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/someservicegenerator
[string prefixer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/stringprefixer
[date prefixer]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/dateprefixer
[sops encoded secrets]: https://github.com/monopole/sopsencodedsecrets
[SOPSGenerator]: https://github.com/omninonsense/kustomize-sopsgenerator

#### Examples

* [service generator] - generate a service from a name and port argument.
* [string prefixer] - uses the value in `metadata/name` as the prefix.
   This particular example exists to show how a plugin can
   transform the behavior of a plugin.  See the
   `TestTransformedTransformers` test in the `target` package.
* [date prefixer] - prefix the current date to resource names, a simple
   example used to modify the string prefixer plugin just mentioned.
* [secret generator] - generate secrets from a toy database.
* [sops encoded secrets] - a more complex secret generator that converts SOPS files into Kubernetes Secrets
* [SOPSGenerator] - another generator that decrypts SOPS files into Secrets
* [All the builtin plugins](https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin).
   User authored plugins are
   on the same footing as builtin operations.

A Go plugin can be both a generator and a
transformer.  The `Generate` method will run along
with all the other generators before the
`Transform` method runs.

Here's a build command that sensibly assumes the
plugin source code sits in the directory where
kustomize expects to find `.so` files:

```
d=$XDG_CONFIG_HOME/kustomize/plugin\
/${apiVersion}/LOWERCASE(${kind})

go build -buildmode plugin \
   -o $d/${kind}.so $d/${kind}.go
```
