# OpenAPI Data for Builtin Kubernetes Types

[Kustomization schema]: ./kustomizationapi/
[issue 5016]: https://github.com/kubernetes-sigs/kustomize/issues/5016

Kustomize no longer embeds a full Kubernetes OpenAPI schema document.
The strategic-merge patch metadata the kyaml merge walker needs from that
document (the `x-kubernetes-patch-strategy` and `x-kubernetes-patch-merge-key`
extensions, plus the definition structure needed to reach them) is compiled
in as a small precomputed table, `zz_generated_patchmeta.go`, generated from
the published OpenAPI document of an exact kubernetes/kubernetes tag.

The [Kustomization schema] (the schema for the Kustomization type itself)
remains embedded; it is small and owned by this repository.

Users who need a full schema document at runtime (for example for custom
resources, or for a different Kubernetes version) can supply one with the
`openapi` field in kustomization.yaml; that path is unchanged.

## Regenerating the patch-metadata table

The Makefile in this directory downloads the OpenAPI document for the
kubernetes/kubernetes tag pinned by `API_VERSION` and regenerates the table
from it:

```
make zz_generated_patchmeta.go
```

To move to a newer Kubernetes version, update `API_VERSION` in the Makefile
and rerun. Note the guidance in [issue 5016] before bumping: schema versions
newer than v1.21 drop long-deprecated beta APIs that some users still
kustomize, so a version bump (or a union across versions — the generator
accepts multiple documents) is a deliberate decision, not routine
maintenance.

## Verifying the table

```
make verify-patchmeta
```

This downloads the pinned tag's document and runs
`TestPrecomputedPatchMetaEquivalence`, which walks every resource root in
lockstep against the parsed document and asserts the table reports identical
patch strategies and merge keys at every path (the test skips when the
document is not on disk).

## Precomputations

To avoid expensive schema lookups, some functions have precomputed results
based on the schema (`precomputedIsNamespaceScoped`, and the patch-metadata
table above). Unit tests ensure these stay in sync; if they fail, follow the
suggested diff or regenerate.

## Kustomization schema

```
make kustomizationapi/swagger.go
```

## Run all tests

At the top of the repository, run the tests.

```
make prow-presubmit-check >& /tmp/k.txt; echo $?
```

The exit code should be zero; if not, examine `/tmp/k.txt`.
