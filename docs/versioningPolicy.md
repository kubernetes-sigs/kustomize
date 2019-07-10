# Versioning

Running `kustomize` means one is running a
particular version of a program, using a
particular version of underlying packages, and
reading a particular version of a [kustomization]
file.

## Program Versioning

The command `kustomize version` prints a three
field version tag (e.g. `v3.0.0`) that aspires to
[semantic versioning].

When enough changes have accumulated to
warrant a new release, a [release process]
is followed, and the fields in the version
number are bumped per semver.

## Kustomize packages

At the time of writing, the kustomize program and
the packages it uses (and exports) are in the same
Go module (see the top level `go.mod` file in the
repo).

[trailing major version indicator]: https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher

Thus, they share the module's version number, per
its git tag (e.g. `v3.0.0`), whose major verion
number matches the [trailing major version
indicator] in the module name (e.g. the `/v3` in
`sigs.k8s.io/kustomize/v3`).

The non-internal packages in the Go module
`sigs.k8s.io/kustomize/v3`, introduced in
[v3.0.0](v3.0.0.md), conform to [semantic
versioning].


## Kustomization File Versioning

At the time of writing (circa release of v2.0.0):

- A [kustomization] file is just a YAML file that
  can be successfully parsed into a particular Go
  struct defined in the `kustomize` binary.

- This struct does not have a version number,
  which is the same as saying that its version
  number matches the program's version number,
  since it's compiled in.

### Field Change Policy

- A field's meaning cannot be changed.

- A field may be deprecated, then removed.

- Deprecation means triggering a _minor_ (semver)
  version bump in the program, and
  defining a migration path in a non-fatal
  error message.

- Removal means triggering a _major_ (semver)
  version bump, and fatal error if field encountered
  (as with any unknown field).

### The `edit fix` Command

This `kustomize` command reads a Kustomization
file, converts deprecated fields to new
fields, and writes it out again in the latest
format.

This is a type version upgrade mechanism that
works within _major_ program revisions.  There is
no downgrade capability, as there's no use case
for it (see discussion below).

### Examples

At the time of writing, in v1.0.x, there were 12
minor releases, with backward compatible
deprecations fixable via `edit fix`.

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

### Kustomization file versioning

The critical difference between k8s API versioning
and kustomization file versioning is

- A k8s API server is able to go _forward_ and
  _backward_ in versioning, to work with older
  clients, over [some range].

- The `kustomize edit fix` command only moves
  _forward_ within a _major_ program
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

#### Optional use of k8s `kind` and `apiVersion`

At the time of writing two [special] k8s
resource fields are allowed, but not required, in
a kustomization file: [`kind`] and [`apiVersion`].

If either field is present, they both must be, and
they must have the following values:

``` yaml
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1
```

They are allowed to exist and have specific values
in a kustomization file only as a sort of
domain-squatting behavior for some future API.  A
kustomize user gains nothing from adding these
fields to a kustomization file.

### Why not require `kind` and `apiVersion`

#### Ease of use and setting proper expectations

Use cases for a kustomization file don't include a
server storing muliple k8s kinds and offering
version downgrades.

The kustomization file is more akin to a
`Makefile`.  A kustomize command can either read a
kustomization file, or it cannot, and in the later
case will complain as specifically as possible
about why (e.g. `unknown field Foo`).

So requiring a `kind` and `apiVersion` would just
be boilerplate in a user's files, and in all the
examples and tests.

Nevertheless, _a user still benefits from a
versioning policy_ and has a `fix` command to
upgrade files as needed.

#### We can change our minds

When/if the kustomization struct graduates to some
kind of API status, with an expectation of
"versionless" storage and downgrade capability,
whatever it looks like at that moment can be
locked into `/v1beta1` or `/v1` and the `kind`
and `apiVersion` fields can be required from that
moment forward.

[field change policy]: #field-change-policy
[some range]: https://kubernetes.io/docs/reference/using-api/deprecation-policy
[proposal]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/customresources-versioning.md
[beta-level rules]: https://github.com/kubernetes/community/blob/master/contributors/devel/api_changes.md#alpha-beta-and-stable-versions
[changes]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md
[adapt]: https://github.com/kubernetes-sigs/kustomize/blob/master/pkg/types/kustomization.go#L166
[special]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources
[k8s API]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[conventions]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[release process]: ../releasing/README.md
[kustomization]: glossary.md#kustomization
[`kind`]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#types-kinds
[`apiVersion`]: https://kubernetes.io/docs/concepts/overview/kubernetes-api/#api-versioning
[semantic versioning]: https://semver.org
