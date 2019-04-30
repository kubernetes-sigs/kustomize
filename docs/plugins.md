# kustomize plugins

Kustomize offers a plugin framework for people to
write their own resource generators (e.g. a helm
chart processor, a generator that automatically
attaches a Service and Ingress object to each
Deployment) and their own resource transformers
(e.g. a transformer that does some highly
customized processing of the container command
line).

## Specification in `kustomization.yaml`

A kustomization file has two new fields in v2.1:
_generators_ and _transformers_.

Each accepts a list of strings as its arguments:

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

This is exactly like the syntax of the `resources` field.

The value of each entry in a `resources`, `generators`
or `transformers` list must be a relative path to a
YAML file, or a path or URL to a [kustomization].

[kustomization]: glossary.md#kustomization

In the former case the YAML is read from disk directly,
and in the latter case a kustomization is performed,
and its YAML output is merged with the YAML read
directly from files.  The net result in all three cases
is an array of YAML objects.

_Each_ object resulting from a `generators` or
`transformers` field is now further interpreted by
kustomize as a _plugin configuration_ object.

## Configuration and execution

A kustomization file could have the following lines:

```
generators:
- chartInflator.yaml
```

Given this, the kustomization process would expect to
find a file called `chartInflator.yaml` in the
kustomization [root](glossary.md#root).

The file `chartInflator.yaml` could contain:

```
apiVersion: someteam.example.com/v1
kind: ChartInflatorExec
metadata:
  name: notImportantHere
chartName: minecraft
```

The `apiVersion` and `kind` fields of the configuration
objects are used to _locate_ the plugin.

The rest of the file (actually the entire file) is
sent to the plugin as configuration - i.e. as the
plugin's construction arguments.

A kustomization file could include multiple
instantiations of the same plugin, with different
arguments (e.g. to inflate two different helm
charts or two instances of the same chart but with
different values files).

The value order in the `generators` field doesn't
matter, because generated objects are just added
to a sea of objects that kustomize transforms and
emits.

The specified order of transformers in the
`transformers` field is, however, respected, as
transformers aren't expected to be commutative.

Generator plugins are run after processing the
`resources` field (which _reads_ resources), to
_create_ additional resources.

The full set of resources is then passed into the
transformation pipeline, where native (legacy)
transformations like `namePrefix` and
`commonLabel` are applied, followed by all the
transformers run in the order specified.


## Placement

[k8s object]: glossary.md#kubernetes-style-object

Given a configuration object (whick looks like any
other [k8s object]), kustomize will first look for an
_executable_ file called

```
$XDG_CONFIG_HOME/kustomize/plugin/${apiVersion}/${kind}
```

The default value of `XDG_CONFIG_HOME` is `$HOME/.config`.

If this file is not found or is not executable,
kustomize will look for a file called `${kind}.so`
in the same directory and attempt to load it as a
[Go plugin](#go-plugins).

If both checks fails, the plugin load fails the overall
kustomize build.

A `kustomize build` attempt with plugins that
omits the flag

TODO: Change flag

> `--enable_alpha_goplugins_accept_panic_risk`

will fail with a warning about plugin use.

Flag use is an opt-in acknowledging the absence of
plugin provenance.  Its meant to give pause to
someone who blindly downloads a kustomization from
the internet and attempts to run it, without
realizing that it might attempt to run 3rd party
code.


## Writing plugins

### Exec plugins

TODO: Add ptr to example.

A exec plugin is any executable that accepts a
single argument on it's command line - the name of
a YAML file containing its configuration (which it
presumably reads if it needs additional
configuration).

A generator plugin accepts nothing on `stdin`, but emits
generated resources to `stdout`.

A transformer plugin accepts resource YAML on `stdin`,
and emits those resources, possibly transformed, to
`stdout`.


### Go plugins

TODO: Add ptr to example.

[Go plugin]: https://golang.org/pkg/plugin/

A [Go plugin] for kustomize looks like this:

> ```
> +build plugin
> 
> package main
>
> import ...
>
> // go:generate go run sigs.k8s.io/kustomize/cmd/pluginator
> type plugin struct{...}
>
> var KustomizePlugin plugin
>
> func (p *plugin) Config(
>    ldr ifc.Loader, rf *resmap.Factory,
>    k ifc.Kunstructured) error {...}
>
> func (p *plugin) Generate() (resmap.ResMap, error) {...}
>
> func (p *plugin) Transform(m resmap.ResMap) error {...}
> ```

The use of the identifiers `plugin`,
`KustomizePlugin` and the three method signatures
`Configurable`, `Generator`, `Transformer` as
shown is _required_.

The plugin author should of course change the
contents of the `plugin` struct, and the three
method bodies, and the import statements, as
desired.

Here's a build command, which assumes the plugin
source code is sitting right next to where the
shared object (`.so`) files are expected to be:

```
dir=$XDG_CONFIG_HOME/kustomize/plugin/${apiVersion}
go build -buildmode plugin -tags=plugin \
    -o $dir/${kind}.so \
    $dir/${kind}.go
```

For the person willing to compile not just a
plugin but all of kustomze as well, a code
generator will be provided that will convert a Go
plugin to statically linked code in your own
compiled version of kustomize.

#### Caveats

Go plugins allow kustomize extensions that

 * can be tested with the same framework kustomize
   uses to test its _builtin_ generators and
   transformers,

 * run without the performance hit of firing up a
   subprocess and marshalling/unmarshalling data
   for each plugin run.

Go plugins work as [defined][Go plugin], but
fall short of what many people think of when they
hear the word _plugin_.

[ELF]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format

Go plugin compilation creates an [ELF] formatted
`.so` file, which by definition has no information
about the _provenance_ of the file.

One cannot know which version of Go was used,
which packages were imported (and their version),
what value of `GOOS` and `GOARCH` were used,
etc. Skew between the compilation conditions of
the main program ELF and the plugin ELF will cause
a failure at load time.

Exec plugins also lack provenance - but they don't
suffer from the skew problem.

