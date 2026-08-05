# kustomize plugins

This directory holds [kustomize plugins][extending kustomize],
each in its own sub-directory.
 
### Directories
 
 * `builtin`
 
   These are plugins written as [Go plugins].
   
   They are converted to statically linked code in the
   kustomize binary by a code generator ([pluginator]) at
   kustomize build time.
   
   They are maintained as part of kustomize.

 * `someteam.example.com/v1`

   Example plugins, maintained
   and tested by the kustomize maintainers,
   but not built-in to kustomize.
   
   Some of these might get promoted to `builtin` someday,
   as happened with the [Helm Chart Inflator].
      
 * `untested/v1`
   
   Untested, unmaintained plugins.
   
   These might be former examples that have been abandoned and
   may soon be deleted, or they might be WIP plugins that will
   someday become examples or builtins.

### Plugin home

Kustomize looks up Exec and Go plugins relative to a plugin home
directory. When resolving it, kustomize tries the following paths in
order and uses the first one that exists:

1. `$KUSTOMIZE_PLUGIN_HOME` — set this to override the default search
   paths below (useful when plugins live outside `$HOME`).
2. `$XDG_CONFIG_HOME/kustomize/plugin`.
3. `$HOME/.config/kustomize/plugin` — used when `$XDG_CONFIG_HOME` is
   unset.
4. `$HOME/kustomize/plugin`.

If none of the above exist, plugin resolution fails. On Windows,
`$HOME` falls back to `$USERPROFILE`. See
[`api/konfig/plugins.go`](../api/konfig/plugins.go) for the resolution
code.

The `generators` and `transformers` fields in `kustomization.yaml` are lists
of file paths, each pointing to a plugin configuration file. The `apiVersion`
and `kind` are read from that referenced configuration file, not from the list
entry itself.

From those fields kustomize derives the plugin directory
`<plugin-home>/<apiVersion group>/<apiVersion version>/<lowercase kind>/` and
looks for the plugin there in two forms: first an executable file named
`<Kind>` (an Exec plugin), then `<Kind>.so` (a Go plugin).

For example, given a configuration file referenced from `transformers`:

```yaml
apiVersion: someteam.example.com/v1
kind: SedTransformer
```

kustomize looks for the executable
`<plugin-home>/someteam.example.com/v1/sedtransformer/SedTransformer`, and
failing that, the Go plugin
`<plugin-home>/someteam.example.com/v1/sedtransformer/SedTransformer.so`.

#### Testing

Regardless of the [style](#plugin-styles) used to write a plugin,
it should be accompanied by a Go unit test, written using the framework
maintained by the kustomize maintainers for just that purpose.

To see how this works, run any plugin test, e.g.
this plugin written in bash:
```
pushd plugin/someteam.example.com/v1/bashedconfigmap
go test -v .
popd
```

For plugins with many tests, it's possible to target just one test:
```
pushd plugin/builtin/patchstrategicmergetransformer
go test -v -run TestBadPatchStrategicMergeTransformer PatchStrategicMergeTransformer_test.go
popd
```

### Plugin styles

For more discussion, see [extending kustomize].

* a bare executable
 
  This can be anything, e.g. a shell script, a shell
  script that runs java in a JVM, or python in a python
  VM, etc.  They accept a YAML stream of resources on
  stdin, and emit a YAML stream on stdout.  They accept
  configuration data in a file specified as the first
  argument on their command line.

  If the executable is written in Go, it can take advantage
  of the same libraries as the kustomize builtin plugins.
  
* a [KRM function]

  These are containerized executables, that are pickier
  about their input.  Rather than accepting a YAML stream
  of k8s resources, they want one `ResourceList` object
  (with the resources in that list).

* a [Go plugin]

  These are built as shared object libraries.  Like
  a Go program, they're written in an unimportable
  `main` package with its own `go.mod` file.
  Go plugins cannot be reliably distributed (see docs),
  and are meant only as a structured way to write a
  builtin plugin intended for distribution with kustomize.

[pluginator]: ../cmd/pluginator
[Helm Chart Inflator]: ./builtin/helmchartinflationgenerator
[KRM function]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
[Go plugin]: https://golang.org/pkg/plugin
[Go plugins]: https://golang.org/pkg/plugin
[extending kustomize]: https://kubectl.docs.kubernetes.io/guides/extending_kustomize/

