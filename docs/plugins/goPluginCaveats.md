[plugin package]: https://golang.org/pkg/plugin
[Go modules]: https://github.com/golang/go/wiki/Modules
[ELF]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format
[tensorflow plugin]: https://www.tensorflow.org/guide/extend/op

# Go plugin Caveats

A _Go plugin_ is a compilation artifact described
by the Go [plugin package].  It is built with
special flags and cannot run on its own.
It must be loaded into a running Go program.

> A normal program written in Go might be usable
> as _exec plugin_, but is not a _Go plugin_.

Go plugins allow kustomize extensions that run
without the cost marshalling/unmarshalling all
resource data to/from a subprocess for each plugin
run.  The Go plugin API assures a certain level of
consistency to avoid confusing downstream
transformers.

Go plugins work as described in the [plugin
package], but fall short of common notions
associated with the word _plugin_.

## The skew problem

Go plugin compilation creates an [ELF] formatted
`.so` file, which by definition has no information
about the provenance of the object code.

Skew between the compilation conditions (versions
of package dependencies, `GOOS`, `GOARCH`) of the
main program ELF and the plugin ELF will cause
plugin load failure, with non-helpful error
messages.

Exec plugins also lack provenance, but won't fail
due to compilation skew.

In either case, the only sensible way to share a
plugin is as some kind of _bundle_ (a git repo
URL, a git archive file, a tar file, etc.)
containing source code, tests and associated data,
unpackable under
`$XDG_CONFIG_HOME/kustomize/plugin`.

In the case of a Go plugin, an _end user_
accepting a shared plugin _must compile both
kustomize and the plugin_.

This means a one-time run of
```
GOPATH=${whatever} go get \
    sigs.k8s.io/kustomize/cmd/kustomize@${releaseVersion}
```

and then a normal development cycle using

```
go build -buildmode plugin \
    -o ${wherever}/${kind}.so ${wherever}/${kind}.go
```
with paths and the release version tag (e.g. `v3.0.0`)
adjusted as needed.

For comparison, consider what one
must do to write a [tensorflow plugin].

## Why support Go plugins?

### Safety
 
The Go plugin developer sees the same API offered
to native kustomize operations, assuring certain
semantics, invariants, checks, etc. An exec
plugin sub-process dealing with this via
stdin/stdout will have an easier time screwing
things up for downstream transformers and
consumers.

Minor point: if the plugin reads files via
the kustomize-provided file `Loader` interface, it
will be constrained by kustomize file loading
restrictions.  Of course, nothing but a code audit
prevents a Go plugin from importing the `io` package
and doing whatever it wants.

### Debugging

A Go plugin developer can debug the plugin _in
situ_, setting breakpoints inside the plugin and
elsewhere while running a plugin in feature tests.

To get the best of both worlds (shareability and safety),
a developer can write an `.go` program that functions
as an _exec plugin_, but can be processed by `go generate`
to emit a _Go plugin_ (or vice versa).

### Unit of contribution 

All the builtin generators and transformers
are themselves Go plugins.  This means that
the kustomize maintainers can promote a contributed
plugin to a builtin without needing code changes
(beyond those mandated by normal code review).

### Ecosystems grow through use

Tooling could ease Go plugin _sharing_, but this
requires some critical mass of Go plugin
_authoring_, which in turn is hampered by
confusion around sharing.  [Go modules], once they
are more widely adopted, will solve the
biggest plugin sharing difficulty: ambiguous
plugin vs host dependencies.
