## fmt

[Alpha] Format yaml configuration files.

### Synopsis

[Alpha] Format yaml configuration files.

Fmt will format input by ordering fields and unordered list items in Kubernetes
objects.  Inputs may be directories, files or stdin, and their contents must
include both apiVersion and kind fields.

- Stdin inputs are formatted and written to stdout
- File inputs (args) are formatted and written back to the file
- Directory inputs (args) are walked, each encountered .yaml and .yml file
  acts as an input

For inputs which contain multiple yaml documents separated by \n---\n,
each document will be formatted and written back to the file in the original
order.

Field ordering roughly follows the ordering defined in the source Kubernetes
resource definitions (i.e. go structures), falling back on lexicographical
sorting for unrecognized fields.

Unordered list item ordering is defined for specific Resource types and
field paths.

- .spec.template.spec.containers (by element name)
- .webhooks.rules.operations (by element value)

### Examples

	# format file1.yaml and file2.yml
	kustomize config fmt file1.yaml file2.yml

	# format all *.yaml and *.yml recursively traversing directories
	kustomize config fmt my-dir/

	# format kubectl output
	kubectl get -o yaml deployments | kustomize config fmt

	# format kustomize output
	kustomize build | kustomize config fmt