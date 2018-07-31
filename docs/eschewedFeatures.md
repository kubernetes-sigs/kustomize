# Eschewed Features

## Removal directives

`kustomize` supports configurations that can be reasoned about as
_compositions_ or _mixins_ - concepts that are widely accepted as
a best practice in various programming languages.

To this end, `kustomize` offers various _addition_ directives.  One
can add labels, annotations, patches, resources and bases.
Corresponding _removal_ directives are not offered.

Removal semantics would introduce many possibilities for
inconsistency, and the need to add code to detect, report and
reject it.  It would also allow, and possibly encourage,
unnecessarily complex configuration layouts.

When faced with a situation where removal is desirable, it's
always possible to remove things from a base like labels and
annotations, and/or split multi-resource manifests into individual
resource files - then add things back as desired via the
[kustomization].

If the underlying base is outside of one's control, an [OTS
workflow] is the recommended best practice.  Fork the base, remove
what you don't want and commit it to your private fork, then use
kustomize on your fork.  As often as desired, use _git rebase_ to
capture improvements from the upstream base.

## Build-time side effects from CLI args or env variables

`kustomize` supports the best practice of storing one's
entire configuration in a version control system.

Changing `kustomize build` configuration output as a result
of additional arguments or flags to `build`, or by
consulting shell environment variable values in `build`
code, would violate that goal.

`kustomize` insteads offers [kustomization] file `edit`
commands.  Like any shell command, they can accept
environment variable arguments.

For example, to set the tag used on an image to match an
environment variable, run

```
kustomize edit set imagetag nginx:$MY_NGINX_VERSION
```

as part of some encapsulating work flow executed before
`kustomize build`.


## Globs in kustomization files

`kustomize` supports the best practice of storing one's
entire configuration in a version control system.

Globbing the local file system for files not explicitly
declared in the [kustomization] file at `kustomize build` time
would violate that goal.

Allowing globbing in a kustomization file would also introduce
the same problems as allowing globbing in [java import]
declarations or BUILD/Makefile dependency rules.

`kustomize` will instead provide kustomization file editting
commands that accept globbed arguments, expand them at _edit
time_ relative to the local file system, and store the resulting
explicit names into the kustomization file.

In this way the resources, patches and bases used at _build time_
remain explicitly declared in version control.


[base]: glossary.md#base
[kustomization]: glossary.md#kustomization
[OTS workflow]: workflows.md#off-the-shelf-configuration
[java import]: https://www.codebyamir.com/blog/pitfalls-java-import-wildcards
