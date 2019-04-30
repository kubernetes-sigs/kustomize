# kustomize plugins

Kustomize offers a plugin framework for people to
write their own resource _generators_ (e.g. a helm
chart processor, a generator that automatically
attaches a Service and Ingress object to each
Deployment) and their own resource _transformers_
(e.g. a transformer that does some highly
customized processing of the container command
line).

## Specification in `kustomization.yaml`

Start by adding a `generators:` and/or `transformers:`
field to your kustomization.

Each field is a string array:

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
or `transformers` array must be a relative path to a
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

## Configuration

A kustomization file could have the following lines:

```
generators:
- chartInflator.yaml
```

Given this, the kustomization process would expect to
find a file called `chartInflator.yaml` in the
kustomization [root](glossary.md#kustomization-root).

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

## Execution

Plugins are only used during a run of the
`kustomize build` command.

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

Given a plugin configuration object (it looks like any
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

> `--enable_alpha_goplugins_accept_panic_risk`

will fail with a warning about plugin use.

_TODO: Change flag_

Flag use is an opt-in acknowledging the absence of
plugin provenance.  It's meant to give pause to
someone who blindly downloads a kustomization from
the internet and attempts to run it, without
realizing that it might attempt to run 3rd party
code in plugin form.  The plugin would have to be
installed already, but nevertheless the flag is a
reminder.

## Writing plugins

### Exec plugins

[chartinflator]: ../plugin/someteam.example.com/v1/ChartInflatorExec

See this example [helm chart inflator][chartInflator].

A exec plugin is any executable that accepts a
single argument on its command line - the name of
a YAML file containing its configuration.

A generator plugin accepts nothing on `stdin`, but emits
generated resources to `stdout`.

A transformer plugin accepts resource YAML on `stdin`,
and emits those resources, possibly transformed, to
`stdout`.

kustomize uses an exec plugin adapter to provide
marshalled resources on `stdin` and capture
`stdout` for further processing.

### Go plugins

[Go plugin]: https://golang.org/pkg/plugin/
[secretgenerator]: ../plugin/builtin/SecretGenerator.go

See this example [secret generator][secretGenerator].

A [Go plugin] for kustomize looks like this:

> ```
> +build plugin
>
> package main
>
> import (
>	"sigs.k8s.io/kustomize/pkg/ifc"
>	"sigs.k8s.io/kustomize/pkg/resmap"
>   ...
> )
>
> type plugin struct {...}
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
method bodies, and add imports as desired.

Here's a build command, which assumes the plugin
source code is sitting right next to where the
shared object (`.so`) files are expected to be:

```
dir=$XDG_CONFIG_HOME/kustomize/plugin/${apiVersion}
go build -buildmode plugin -tags=plugin \
    -o $dir/${kind}.so \
    $dir/${kind}.go
```


#### Caveats

Go plugins allow kustomize extensions that

 * can be tested with the same framework kustomize
   uses to test its _builtin_ generators and
   transformers,

 * run without the performance hit of firing up a
   subprocess and marshalling/unmarshalling data
   for each plugin run.

[ELF]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format

Go plugins work as [defined][Go plugin], but fall
short of what many people think of when they hear
the word _plugin_.  Go plugin compilation creates
an [ELF] formatted `.so` file, which by definition
has no information about the _provenance_ of the
file.

One cannot know which version of Go was used,
which packages were imported (and their version),
what value of `GOOS` and `GOARCH` were used,
etc. Skew between the compilation conditions of
the main program ELF and the plugin ELF will cause
a failure at load time.

Exec plugins also lack provenance, but don't
suffer from the skew problem.
