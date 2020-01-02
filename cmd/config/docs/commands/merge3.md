## merge3

[Alpha] Merge diff of Resource configuration files into a destination (3-way)

### Synopsis

[Alpha] Merge diff of Resource configuration files into a destination (3-way)

Merge3 performs a 3-way merge by applying the diff between 2 sets of Resources to a 3rd set.

Merge3 may be for rebasing changes to a forked set of configuration -- e.g. compute the difference between the original
set of Resources that was forked and an updated set of those Resources, then apply that difference to the fork.

If a field value differs between the ORIGINAL_DIR and UPDATED_DIR, the value from the UPDATED_DIR is taken and applied
to the Resource in the DEST_DIR.

For information on merge rules, run:

	kustomize config docs-merge3

### Examples

    kustomize config merge3 --ancestor a/ --from b/ --to c/