# Built-in OpenAPI data

Kustomize embeds a compiled OpenAPI bundle for Kubernetes built-in types. The
runtime artifact is:

```
kubernetesapi/data/kubernetes-openapi-union-v1.36.bundle-v1.json.gz
```

The Kubernetes suffix identifies the newest Kubernetes schema represented by
the bundle at minor-version granularity. `bundle-v1` is the independent
artifact format version. The bundle coverage metadata records the exact source
range, currently v1.21.2 through v1.36.2.

`DefaultOpenAPI` and `kustomize openapi info` identify this artifact as
`v1.36`; that value is the union artifact's ceiling minor, not a claim that the
bundle contains only the v1.36.2 schema. The previous `openapi.version:
v1.21.2` value remains accepted as a compatibility alias and loads the same
union artifact.

The bundle contains the complete OpenAPI definitions and a compact index from
GVK to root definition and resource scope. API paths and other top-level
OpenAPI fields are compiler inputs and are not embedded in the runtime binary.

The compiler processes sources newest-first. A definition from the newest
source containing that definition is retained as a whole; definitions are
never merged field-by-field. Older sources fill definition names and GVKs that
are absent from newer Kubernetes releases. Known resource scopes must agree
across every source. This policy preserves removed APIs while using the newest
available schema for APIs that remain present.

## Regenerating the bundle

The source manifest is:

```
kubernetesapi/sources/builtin-union-v1.36.json
```

It lists one official Kubernetes `api/openapi-spec/swagger.json` snapshot for
every minor release in the covered range. v1.21.2 is retained as the existing
baseline; later end-of-life minors use their final patch release, and supported
minors use the patch release current when this bundle was generated. Each entry
pins the release's peeled Git commit and the uncompressed source SHA-256. The
gzip-compressed JSON files live beside the manifest in a compiler-only
directory: they are checked in for offline reproducibility but are not embedded
or imported by the kyaml runtime.

Source acquisition is deliberately separate from compilation. A manifest
entry identifies this immutable upstream file:

```
https://raw.githubusercontent.com/kubernetes/kubernetes/<gitCommit>/api/openapi-spec/swagger.json
```

When updating Kubernetes, verify the release tag's peeled commit, download the
file from that commit, verify its SHA-256, and store it gzip-compressed next to
the manifest. Update the manifest, embedded path, version constants, and
generated files together.

Regenerate the runtime bundle with:

```
make -C kyaml/openapi generate
```

The compiler accepts the manifest's JSON inputs (and retains single protobuf
input support for the legacy asset), validates each source independently,
constructs the definition and GVK union, validates final local references,
writes canonical JSON, and uses a deterministic gzip header. It also generates
the precomputed resource-scope index used by kyaml. Both generated files must
be byte-for-byte reproducible. Verify them and run the OpenAPI tests with:

```
make -C kyaml/openapi verify
```

The historical `kubernetesapi/v1_21_2/swagger.pb.gz` archive is retained only
to preserve the legacy public asset API. The union compiler no longer uses it,
and the normal Kustomize runtime does not import that compatibility package.

The small Kustomization schema remains as source JSON at
`kustomizationapi/swagger.json` and is embedded directly with `go:embed`.
