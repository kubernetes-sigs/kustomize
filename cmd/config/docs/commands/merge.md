## merge

[Alpha] Merge Resource configuration files

### Synopsis

[Alpha] Merge Resource configuration files

Merge reads Kubernetes Resource yaml configuration files from stdin or sources packages and write
the result to stdout or a destination package.

Resources are merged using the Resource [apiVersion, kind, name, namespace] as the key.  If any of
these are missing, merge will default the missing values to empty.

Resources specified later are high-precedence (the source) and Resources specified
earlier are lower-precedence (the destination).

For information on merge rules, run:

	kustomize config docs merge

### Examples

    cat resources_and_patches.yaml | kustomize config merge > merged_resources.yaml