# Eschewed Features

## Removal Directives

Kustomize supports configurations that can be reasoned about as
_compositions_ or _mixins_ - concepts that are widely accepted as
a best practice in various programming languages.

To this end, Kustomize offers various _addition_ directives.  One
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

## Environment Variable Substitution

Kustomize wants to support the best practice of storing one's
configuration in a version control system.

Dynamically mixing in data from the environment (at `kustomize
build` time) would violate that goal.

If one wants to, say, set the tag used on an image to match an
environment variable, the best practice would be to make
the command

```
kustomize edit set imagetag nginx:$MY_NGINX_VERSION
```

part of some encapsulating work flow executed before `kustomize
build`.


[base]: glossary.md#base
[kustomization]: glossary.md#kustomization
[OTS workflow]: workflows.md#off-the-shelf-configuration
