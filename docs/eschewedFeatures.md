# Eschewed Features

The maintainers established this list to
place bounds on the kustomize feature
set.  The bounds can be changed with
a consensus on the risks.

For a bigger picture about why kustomize
does some things and not others, see the
glossary entry for [DAM].

## Removal directives

`kustomize` supports configurations that can be reasoned about as
_compositions_ or _mixins_ - concepts that are widely accepted as
a best practice in various programming languages.

To this end, `kustomize` offers various _addition_ directives.
One may add labels, annotations, patches, resources, bases, etc.
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

## Unstructured edits

_Structured edits_ are changes controlled by
knowledge of the k8s API, and YAML or JSON syntax.

Most edits performed by kustomize can be expressed as
[JSON patches] or [SMP patches].  Common edits, like
adding labels or adding a name prefix, get dedicated 
shorthand commands.  Another class of edits take
data from one specific object's field and use it in
another (e.g. a service object's name found and
copied into a container's command line).

These edits are designed to create valid output
given valid input, and can provide syntactically
and semantically informed error messages if inputs
are invalid.

_Unstructured edits_, e.g. a templating approach,
or a command to replace any target string in the
character stream with some other string, aren't
limited by any syntax or object structure.
  
Such powerful techniques are eschewed because
- There would be no way to say that a kustomization
  was correct without running it and checking
  the output.
- Errors in the output would be
  disconnected from the edit that caused it.
- They are toil to maintain by a rotating
  staff of operators.
    
Kustomizations are meant to be sharable and stackable.
Imagine tracing down a problem rooted in a
clever set of stacked regexp replacements
performed by various overlays on some remote base. 

Other tools (sed, jinja, erb, envsubst, helm, ksonnet,
etc.) provide varying degrees of unstructured editting
and/or embedded languages, and can be used instead
of, or in a pipe with, kustomize.

## Build-time side effects from CLI args or env variables

`kustomize` supports the best practice of storing one's
entire configuration in a version control system.

Changing `kustomize build` configuration output as a result
of additional arguments or flags to `build`, or by
consulting shell environment variable values in `build`
code, would frustrate that goal.

`kustomize` insteads offers [kustomization] file `edit`
commands.  Like any shell command, they can accept
environment variable arguments.

For example, to set the tag used on an image to match an
environment variable, run

```
kustomize edit set image nginx:$MY_NGINX_VERSION
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

[base]: glossary.md#base
[DAM]: glossary.md#declarative-application-management
[java import]: https://www.codebyamir.com/blog/pitfalls-java-import-wildcards
[JSON patches]: glossary.md#patchjson6902
[kustomization]: glossary.md#kustomization
[OTS workflow]: workflows.md#off-the-shelf-configuration
[SMP patches]: glossary.md#patchstrategicmerge
