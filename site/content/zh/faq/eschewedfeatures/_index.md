---
title: "Eschewed Features"
linkTitle: "Eschewed Features"
type: docs
weight: 99
description: >
    Eschewed Features
---

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
[JSON patches] or [SMP patches].
Those can be verbose, so common patches,
like adding labels or annotatations, get dedicated
transformer plugins - `LabelTransformer`,
`AnnotationsTransformer`, etc.
These accept relatively simple YAML configuration
allowing easy targeting of any number of resources.

Another class of edits take data from one specific
object's field and use it in another (e.g. a service
object's name found and copied into a container's
command line).  These reflection-style edits
are called _replacements_.

The above edits create valid output given valid input,
and can provide syntactically and semantically
informed error messages if inputs are invalid.

_Unstructured edits_, edits that don't limit
themselves to a syntax or object structure,
come in many forms.  A common one in the
configuration domain is the template or
parameterization approach.

In this technique, the source
material is sprinkled with strings of the
form `${VAR}`.  A scanner replaces them
with a value taken from a map using `VAR`
as the map key. It's trivial to implement.

kustomize eschews parameterization, because

- The source yaml gets polluted with `$VARs`
  and can no longed be applied as is
  to the cluster (it _must_ be processed).
- The source material is no longer structured,
  making it unusable with any YAML processor.
  It's no longer _data_, it's now logic that
  must be compiled.
- Errors in the output are disconnected from
  the edit that caused it.
- The input becomes [unintelligible] as the project
  scales in any number of dimensions (resource
  count, cluster count, environment count, etc.)

Kustomizations are meant to be sharable and stackable.
Imagine tracing down a problem rooted in a
clever set of stacked regexp replacements
performed by various overlays on some remote base.
We've used such systems, and never want to again.

Other tools (sed, jinja, erb, envsubst, kafka, helm, ksonnet,
etc.) provide varying degrees of unstructured editting
and/or embedded languages, and can be used instead
of, or in a pipe with, kustomize.  If you want to
go all-in on _configuration as a language_, consider [cue].

kustomize is going to stick to YAML in / YAML out.

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

[base]: /kustomize/api-reference/glossary#base
[DAM]: /kustomize/api-reference/glossary#declarative-application-management
[java import]: https://www.codebyamir.com/blog/pitfalls-java-import-wildcards
[JSON patches]: /kustomize/api-reference/glossary#patchjson6902
[kustomization]: /kustomize/api-reference/glossary#kustomization
[OTS workflow]: /kustomize/api-reference/glossary#off-the-shelf-configuration
[SMP patches]: /kustomize/api-reference/glossary#patchstrategicmerge
[parameterization pitfall discussion]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/declarative-application-management.md#parameterization-pitfalls
[unintelligible]: https://github.com/helm/charts/blob/e002378c13e91bef4a3b0ba718c191ec791ce3f9/stable/artifactory/templates/artifactory-deployment.yaml
[cue]: https://cuelang.org/
