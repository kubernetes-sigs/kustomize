# Built-in OpenAPI data

Kustomize embeds a compiled OpenAPI bundle for Kubernetes built-in types. The
runtime artifact is:

```
kubernetesapi/data/kubernetes-openapi-union.bundle-v1.json.gz
```

The path is stable across Kubernetes releases. `bundle-v1` is the independent
artifact format version. The bundle metadata records its exact Kubernetes
source range and, for each source, the peeled Git commit and uncompressed
SHA-256.

`DefaultOpenAPI` is generated from the newest selected Kubernetes minor, and
`kustomize openapi info` prints its current value. It identifies the union
artifact's ceiling minor, not a claim that the bundle contains only that
minor's schema. The stable selector `builtin` and minor versions covered by
the bundle are accepted as aliases for the same union artifact. Use `builtin`
in configuration that should continue to select the built-in schema as its
ceiling advances. The historical exact value `v1.21.2` is also retained as a
compatibility alias.

The bundle contains the complete OpenAPI definitions and a compact index from
GVK to root definition and resource scope. API paths and other top-level
OpenAPI fields are compiler inputs and are not embedded in the runtime binary.

The compiler processes sources newest-first. The schema fields of a definition
from the newest source containing that definition are retained as a whole;
schemas are never merged field-by-field. The compiler separately unions the
GVK inventory from every source, then normalizes each definition's
`x-kubernetes-group-version-kind` extension from that inventory. Older sources
therefore fill definition names and GVKs that are absent from newer Kubernetes
releases without changing a newer schema's fields. Known resource scopes must
agree across every source. This policy preserves removed APIs while using the
newest available schema for APIs that remain present.

## Kubernetes source selection policy

The union uses exactly one OpenAPI snapshot for every Kubernetes minor in its
covered range. A new minor is represented by its `v1.N.0` release. Publishing a
new patch release does not by itself trigger a source update or a tracking
issue. Once selected, a source remains pinned unless one of the documented
exceptions below applies.

A non-zero patch release is allowed only for one of these documented
exceptions:

- preserving a historical compatibility baseline; or
- taking an upstream correction to OpenAPI information retained or consumed
  by Kustomize, such as GVK and scope discovery, field names and references,
  types, patch strategies and merge keys, or Kubernetes list semantics.

For an OpenAPI correction, use the first stable patch release containing the
fix and record the upstream Kubernetes pull request in the version list.
Unrelated fixes, a newer patch number, and description-only changes are not
reasons to replace a source. This keeps source selection deterministic and
avoids recurring source churn.

The current exceptions are:

| Minor | Source | Reason |
| --- | --- | --- |
| v1.21 | v1.21.2 | Retains the OpenAPI source used by the historical built-in schema. |
| v1.26 | v1.26.2 | Includes the corrected `ResourceClaim.status.reservedFor` list semantics from [kubernetes/kubernetes#115400]. |

[kubernetes/kubernetes#115400]: https://github.com/kubernetes/kubernetes/pull/115400

## Updating and regenerating the bundle

The selected Kubernetes versions are stored oldest-first in:

```
kubernetesapi/builtin-versions.json
```

To add a newly covered Kubernetes minor, append its `v1.N.0` version to the
`versions` array and run the generator. Do not remove older versions: their
APIs may no longer exist in newer Kubernetes releases. A non-zero patch also
requires an entry in `patchVersionExceptions` as described above.

For an ordinary new minor, that one version entry is the only hand-written
change; commit it together with the files produced by the generator. The
download cache remains outside the repository and must not be committed.

Raw Kubernetes OpenAPI documents are generator inputs and are not stored in
this repository. For versions already represented by the checked-in bundle,
the bundle's source metadata locks the peeled Git commit and uncompressed
SHA-256. For a newly appended version, the generator resolves the official
Kubernetes release tag to its peeled commit. It then downloads this immutable
upstream file:

```
https://raw.githubusercontent.com/kubernetes/kubernetes/<gitCommit>/api/openapi-spec/swagger.json
```

Downloaded documents are validated and cached by their uncompressed SHA-256.
The default cache is the platform user cache under `kustomize/openapi`. Set
`KUSTOMIZE_OPENAPI_CACHE_DIR` to use a different directory. Cache entries are
always re-hashed before use; an invalid entry is discarded.

Regenerate the runtime bundle with:

```
make -C kyaml/openapi generate
```

Generation may access GitHub for cache misses and to resolve a newly selected
release tag. To prohibit network access, invoke the generator command shown in
`builtin_schema.go` with `-offline`. Offline generation succeeds only when all
required documents are already cached; a new version cannot be resolved in
offline mode. `GITHUB_TOKEN` may be set for GitHub API authentication during
tag resolution; it is not sent when downloading the raw OpenAPI document.

The checked-in bundle is also the default source lock for regeneration. The
generator's `-lock` flag can point at that bundle while `-output` points at a
temporary file, allowing an independent byte comparison. Restore the bundle
from Git before regenerating if it was deleted.

The compiler validates each source independently, constructs the definition
and GVK union, validates final local references, writes canonical JSON, and
uses a deterministic gzip header. It also generates the built-in version
metadata, API provenance test data, and precomputed resource-scope index.
Provenance examples for later minors are selected from GVKs first seen in that
source. Because the compatibility floor has no earlier input to compare, its
known Kubernetes v1.21 additions are curated and verified against the selected
source during generation.

These generated files and the stable runtime bundle are checked in. Do not
edit their paths, version constants, or provenance data manually.

To download or reuse every locked source, regenerate all outputs, and fail if
they differ from the checked-in files, run:

```
make -C kyaml/openapi verify-generated
```

This target has the same GitHub and cache requirements as `generate`.

Normal builds and unit tests read only checked-in generated data and do not
download Kubernetes OpenAPI documents. Verify the generated data and run the
offline OpenAPI tests with:

```
make -C kyaml/openapi verify
```

The compiler retains single protobuf input support for the legacy asset. The
historical `kubernetesapi/v1_21_2/swagger.pb.gz` archive remains checked in only
to preserve its legacy public asset API; the union compiler and normal
Kustomize runtime do not import it.

The small Kustomization schema remains as source JSON at
`kustomizationapi/swagger.json` and is embedded directly with `go:embed`.
