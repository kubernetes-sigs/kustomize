# Architecture

## Repository layout

[human-edited docs]: https://github.com/kubernetes-sigs/cli-experimental/tree/master/site
[generated docs]: https://github.com/kubernetes-sigs/cli-experimental/tree/master/docs
[rendered docs]: https://kubectl.docs.kubernetes.io
[openapi]: https://kubernetes.io/blog/2016/12/kubernetes-supports-openapi

[`api` module]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/go.mod
[`api`]: #the-api-module
[`cmd/config` module]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/go.mod
[`cmd/config`]: #the-cmdconfig-module
[`kustomize` module]: https://github.com/kubernetes-sigs/kustomize/blob/master/kustomize/go.mod
[`kustomize`]: #the-kustomize-module
[`kyaml` module]: https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/go.mod
[`kyaml`]: #the-kyaml-module

[`kyaml/kio.Filter`]: https://github.com/Kubernetes-sigs/kustomize/blob/master/kyaml/kio/kio.go

[3922]: https://github.com/kubernetes-sigs/kustomize/issues/3922


 |  directory  |             |
 | ---------:  | :---------- |
 |       `api` | The [`api`] module, holding high level kustomize code, suitable for import by other programs.  |
 |       `cmd` | Various Go programs aiding repo management (see also `hack`). As an outlier, includes the special [`cmd/config`] module. |
 |      `docs` | Old home of documentation.  New homes: [human-edited docs], [generated docs] and [rendered docs]. |
 |  `examples` | Full kustomization examples that run as pre-commit-merge tests.  |
 | `functions` | Examples of plugins in KRM function form. TODO([3922]): Move under `plugin`. |
 |      `hack` | Various shell scripts to help with code mgmt. |
 | `kustomize` | The [`kustomize`] module, holding the `main.go` for kustomize. |
 |     `kyaml` | The [`kyaml`] module, holding Kubernetes-specific YAML editing packages used by the [`api`] module. |
 |    `plugin` | Examples of Kustomize plugins.  |
 | `releasing` | Instructions for releasing the various modules.  |
 |      `site` | Old generated documentation, kept to provide redirection links to the new docs. |


## Modules

[semantically versioned]: https://semver.org
[Go modules]: https://github.com/golang/go/wiki/Modules

The [Go modules] in the kustomize repository are [semantically versioned].


### The `kustomize` module

Intra-repo module dependence: [`api`], [`cmd/config`], [`kyaml`].

The [`kustomize` module] contains the `main.go` for `kustomize`, buildable with

```
(cd kustomize; go install .)
```

[appears in kubectl]: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/kubectl/pkg/cmd/kustomize/kustomize.go

Below this are packages containing
[cobra](http://github.com/spf13/cobra) commands implementing `build`,
`edit`, `fix`, etc., packages linked together by `main.go`.

These command packages are intentionally public, semantically
versioned, and can be used in other programs.  Specifically, the
`kustomize build` command [appears in kubectl] as `kubectl kustomize`.

The code in the `build` package is dominated by flag validation,
with minimal business logic.  The critical lines are something
like

```
# Make a kustomizer.
k := krusty.MakeKustomizer(
  HonorKustomizeFlags(krusty.MakeDefaultOptions()),
)

# Run the kustomizer, sending location of kustomization.yaml
m := k.Run(fSys, "/path/to/dir")

# Write the result as YAML.
writer.Write(m.AsYaml())
```

The `krusty` package is in the [`api`] module.

### The `api` module

Intra-repo module dependence: [`kyaml`] and code generated from builtin plugin modules.

The [`api` module] is used by CLI programs like `kustomize` and `kubectl`
to read and honor `kustomization.yaml` files and all that implies.

The main public packages in the [`api` module] are

 | package   |             |
 | --------: | :---------- |
 | `filesys` | A kustomize-specific file system abstraction, to ease writing tests. |
 | `filters` | Implementations of [`kyaml/kio.Filter`] used by kustomize to transform Kubernetes objects. |
 |  `konfig` | Configuration methods and constants in the kustomize API. |
 |  `krusty` | Primary API entry point.  Holds the kustomizer and hundreds of tests for it. |
 |  `loader` | Loads kustomization files and the files they refer to, enforcing security rules. |
 |  `resmap` | The primary internal data structure over which the kustomizer and filters work. |
 |   `types` | The `Kustomization` object and ancillary structs. |

### The `cmd/config` module

Intra-repo module dependence: [`kyaml`]

This module contains cobra commands and kyaml-based functionality to
provide unix-like file manipulation commands to kustomize like `grep`
and `tree`.  These commands may be included in any program that
manipulates k8s YAML (e.g. kustomize).

### The `kyaml` module

[gopkg.in/yaml.v3]: https://pkg.go.dev/gopkg.in/yaml.v3

The [`kyaml` module] is a kubernetes-focussed enhancement of [gopkg.in/yaml.v3].

The YAML manipulation performed by a kustomize is based on these libraries.

These libraries evolve independently of kustomize, and other programs depend on them.


## How kustomize build works

The command `kustomize build` accepts a single string argument, which
must resolve to a directory (possibly in a remote git repository)
called the _kustomization root_.

This directory must contain a file called `kustomization.yaml`, with
YAML that marshals into a single instance of a `Kustomization` object.

For the remainder of this document, the word _kustomization_ refers to
either of these things.

This kustomization is the access node to a directed, acyclic graph of
Kubernetes objects, including other kustomizations, to include in a
build.

Execution of `build` starts and ends in the [`api`] module,
frequently dipping into the [`kyaml`] module for lower level
YAML manipulation.

### The `build` flow

  - Validate command lines arguments and flags.

  - Make a `Kustomizer` as a function of those arguments.

  - Call `Run` on the kustomizer, passing the path to the kustomization.

    `Run` returns an instance of `ResMap`, which is the `api`
    package's representation of a list of kubernetes objects.
    These objects are fully hydrated, per the instructions in the
    kustomization.

  - Emit the YAML version of those objects, to files or `stdout`.


### The `Run` function

  - Create various objects

    - A `ResMap` factory.

      Makes  `ResMaps` from byte streams, other `ResMaps`, etc.

    - A file `loader.Loader`.

      It's fed an appropriate set of restrictions, and the path to the kustomization.

    - A plugin loader.

      It finds plugins (transformers, generators or validators)
      and prepare them for running.

    - A `KustTarget` encapsulating all of the above.

      A KustTarget contains one `Kustomization` and represents
	  everything that kustomization can reach.  This will include
	  other `KustTarget` instances, each having a smaller purvue than
	  the one referencing it.

  - Call `KustTarget.Load` to load its kustomization.

    This step deals with deprecations and field changes.

  - Load [openapi] data specified by the kustomization.

    This is needed to recognize k8s kinds and their special
    properties, e.g. which kinds are cluster-scoped, which kinds refer to others, etc.

  - Call `KustTarget.makeCustomizedResmap` to create the `ResMap` result.

    This visits everything referenced by the kustomization,
    performing all generation, transformation and validation.

  - Finish the Run with

    - Optional  `legacy` reordering of objects, overriding
      a FIFO rule.
    - Optional addition of _Managed by kustomize_ annotations.

### The `makeCustomizedResmap` function

  This function starts the process of object transformation,
  as well as accumulation of recursively referenced data.

  - Call `ra := KustTarget.AccumulateTarget`.

    The result, `ra`, is a resource accumulator that contains
    everything referred to by the current kustomization, now full
    hydrated.

  - Uniquify names of generated objects by appending content hashes.

    This cannot be done until the objects are complete.

  - Fix all name references (given that names may have changed).

    E.g. if a ConfigMaps was given a generated name, all objects that
    refer to that ConfigMap must be given it's name.

  - Resolve vars, replacing them with whatever they refer to (a legacy feature).

### The `AccumulateTarget` function

  - Call `AccumulateResources` over the `resources` field (this can recurse).
  - Call `AccumulateComponents` over the `components` field (this can recurse),
  - Load legacy (pre-plugin) global kustomize configuration,
  - Load legacy (pre-openapi) _Custom Resource Definition_ data.
  - In the context of the data loaded above, run the kustomization's
    - generators,
    - transformers,
    - and validators.
  - Accumulate `vars` (make note of them for later replacement).

### `AccumulateResources` and component accumulation

  - If the path leads to a file:

    Accumulate the objects in the file (treating them
    as opaque kubernetes objects).

  - If it's a directory:
    - Create a new `KustTarget` referring to that directory's kustomization.
	- Call `subRa := KustTarget.AccumulateTarget`.
    - Call `ra.MergeAccumulator(subRa)`
	This completes a recursion.

  - If it's a git URL with a path, clone the git repository, and process
    the path as directory as described above.

That completes the top down flow outline of the build process.
