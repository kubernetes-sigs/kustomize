---
title: "Versioning Policy"
linkTitle: "Versioning Policy"
weight: 99
type: docs
---

Running `kustomize` means one is running a
particular version of a program (a CLI), using a
particular version of underlying packages (a Go
API), and reading a particular version of a
[kustomization] file.

> If you're having trouble with `go get`, please
> read  [Go API Versioning](#go-api-versioning)
> and be patient.

## CLI Program Versioning

The command `kustomize version` prints a three
field version tag (e.g. `v3.0.0`) that aspires to
[semantic versioning].

This notion of semver applies only to the CLI;
the command names, their arguments and their flags.

The major version changes when some backward
incompatibility appears in how the commands
behave.

### Installation

See the [installation docs](INSTALL.md).

## Go API Versioning

The public methods in the public packages
of module `sigs.k8s.io/kustomize/api` constitute
the _kustomize Go API_.

#### Version sigs.k8s.io/kustomize/v3 and earlier

[import path]: https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher

In `kustomize/v3` (and preceding major versions), the
kustomize program and the API live the same Go
module at `sigs.k8s.io/kustomize`, at [import path]
`sigs.k8s.io/kustomize/v3`.

This has been fine for the CLI, but it presents a
problem for the Go API.

[minimal version selection]: https://research.swtch.com/vgo-mvs

The process around Go modules, in particular the
notion of [minimal version selection], demands
that the module respect semver.

Almost all the code in module
`sigs.k8s.io/kustomize/v3` is exposed (not in a
directory named `internal`).  Even a minor
refactor changing a method name or argument type
in some deeply buried (but still public) method is
a backward incompatible change.  As a result, Go
API semver hasn't been followed.  This was a mistake.

Some options are

- continue to ignore Go API semver and stick to
  CLI semver (eliminating the usefullness of
  minimal version selection),

- obey semver, and increment the module's major
  version number with every release (drastically
  reducing the usefullness of minimal version
  selection - since virtually all releases will
  be major),

- slow down change in the huge API in favor of
  stability, yet somehow continue to deliver
  features,

- drastically reduce the API surface, stabilize on
  semver there, and refactor as needed inside
  `internal`.

The last option seems the most appealing.

#### The first stable API version is coming

The first stable API version will launch
as the Go module

```
sigs.k8s.io/kustomize/api
```

The _kustomize_ program itself (`main.go`
and CLI specific code) will have moved out of
`sigs.k8s.io/kustomize` and into the new module
`sigs.k8s.io/kustomize/kustomize`.  This is a
submodule in the same repo, and it will retain its
current notion of semver (e.g. a backward
incompatible change in command behavior will
trigger a major version bump).  This module will
not export packages; it's just home to a `main`
package.

The `sigs.k8s.io/kustomize/api` module will
obey semver with a sustainable public
surface, informed by current usage.  Clients
should import packages from this module, i.e.
from import paths prefixed by
`sigs.k8s.io/kustomize/api/` at first,
and later by `sigs.k8s.io/kustomize/api/v2/`.
The kustomize binary
itself is an API client requiring this module.

The clients and API will evolve independently.

## Kustomization File Versioning

The kustomization file is a struct that is part of
the kustomize Go API (the `sigs.k8s.io/kustomize`
module), but it also evolves as a k8s API object -
it has an `apiVersion` field containing its
own version number.

### Field Change Policy

- A field's meaning cannot be changed.
- A field may be deprecated, then removed.
- Deprecation means triggering a _minor_ (semver)
  version bump in the kustomize Go API, and
  defining a migration path in a non-fatal error
  message.
- Removal means triggering a _major_ (semver)
  version bump in the kustomize Go API, and fatal
  error if field encountered (as with any unknown
  field).  Likewise a change in `apiVersion`.

### The `edit fix` Command

This `kustomize` command reads a Kustomization
file, converts deprecated fields to new
fields, and writes it out again in the latest
format.

This is a type version upgrade mechanism that
works within _major_ API revisions.  There is no
downgrade capability, as there's no use case for
it (see discussion below).

### Examples

With the 2.0.0 release, there were three field
removals:

- `imageTag` was deprecated when `images` was
   introduced, because the latter offers more
   general features for image data manipulation.
   `imageTag` was removed in v2.0.0.
- `patches` was deprecated and replaced by
   `patchesStrategicMerge` when `patchesJson6902`
   was introduced, to make a clearer
   distinction between patch specification formats.
   `patches` was removed in v2.0.0.
- `secretGenerator/commands` was removed
   due to security concerns in v2.0.0
   with no deprecation period.

The `edit fix` command in a v2.0.x binary
will no longer recognize these fields.

## Relationship to the k8s API

### Review of k8s API versioning

The k8s API has specific [conventions] and a
process for making [changes].

The presence of an `apiVersion` field in a k8s
native type signals:

- its reliability level (alpha vs beta vs
  generally available),
- the existence of code to provide default values
  to fields not present in a serialization,
- the existence of code to provide both forward
  and backward conversion between different
  versions of types.

The k8s API promises a lossless _conversion_
between versions over a specific range.  This
means that a recent client can write an object
bearing the newest possible value for its version,
the server will accept it and store it in
"versionless" JSON form in storage, and can
convert it to a range of older versions should
an older client request data.

For native k8s types, this all requires writing Go
code in the kubernetes core repo, to provide
defaulting and conversions.

For CRDs, there's a [proposal] on how to manage
versioning (e.g. a remote service can offer type
defaulting and conversions).

### Differences

- A k8s API server is able to go _forward_ and
  _backward_ in versioning, to work with older
  clients, over [some range].
- The `kustomize edit fix` command only moves
  _forward_ within a _major_ API
  version.

At the time of writing, the YAML in a
kustomization file does not represent a [k8s API]
object, and the kustomize command and associated
library is neither a server of, nor a client to,
the k8s API.

### Additional Kustomization file rules

In addition to the [field change policy] described
above, kustomization files conform to
the following rules.

#### Eschew classic k8s fields

Field names with dedicated meaning in k8s
(`metadata`, `spec`, `status`, etc.)  aren't used.
This is enforced via code review.

#### Default values for k8s `kind` and `apiVersion`

In `v3` or below, the two [special] k8s
resource fields [`kind`] and [`apiVersion`] may
be omitted from the kustomization file.

If either field is present, they both must be.
If present, the value of `kind` must be:

> ```
> kind: Kustomization
> ```

If missing, the value of `apiVersion` defaults to

> ```
> apiVersion: kustomize.config.k8s.io/v1beta1
> ```

[field change policy]: #field-change-policy
[some range]: https://kubernetes.io/docs/reference/using-api/deprecation-policy
[proposal]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/customresources-versioning.md
[beta-level rules]: https://github.com/kubernetes/community/blob/master/contributors/devel/api_changes.md#alpha-beta-and-stable-versions
[changes]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md
[adapt]: /types/kustomization.go#L166
[special]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources
[k8s API]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[conventions]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[release process]: https://github.com/kubernetes-sigs/kustomize/tree/master/releasing/README.md
[kustomization]: /kustomize/api-reference/glossary#kustomization
[`kind`]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#types-kinds
[`apiVersion`]: https://kubernetes.io/docs/concepts/overview/kubernetes-api/#api-versioning
[semantic versioning]: https://semver.org
