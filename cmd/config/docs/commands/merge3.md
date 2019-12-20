## merge3

[Alpha] Merge Resource configuration files (3-way)

### Synopsis

[Alpha] Merge Resource configuration files (3-way)

Merge (3-way) reads Kubernetes Resource yaml configuration files from source packages and updated
packages then writes the result to stdout and a destination package.

Resources are merged using the Resource [apiVersion, kind, name, namespace] as the key.  If any of
these are missing, merge will default the missing values to empty.

Resources specified in the updated packages have higher-precedence and Resources specified
in the original packages have lower-precedence.  Resources specified in the destination
packages either keep, clear, or recursively merge their values.

For information on merge rules, run:

	kustomize config docs merge3

### Examples

    kustomize config merge3 --ancestor a/ --from b/ --to c/