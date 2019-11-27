## run

[Alpha] Reoncile config functions to Resources.

### Synopsis

[Alpha] Reconcile config functions to Resources.

run sequentially invokes all config functions in the directly, providing Resources
in the directory as input to the first function, and writing the output of the last
function back to the directory.

The ordering of functions is determined by the order they are encountered when walking the
directory.  To clearly specify an ordering of functions, multiple functions may be
declared in the same file, separated by '---' (the functions will be invoked in the
order they appear in the file).

#### Arguments:

  DIR:
    Path to local directory.

#### Config Functions:

  Config functions are specified as Kubernetes types containing a metadata.configFn.container.image
  field.  This fields tells run how to invoke the container.

  Example config function:

	# in file example/fn.yaml
	apiVersion: fn.example.com/v1beta1
	kind: ExampleFunctionKind
	metadata:
	  configFn:
	    container:
	      # function is invoked as a container running this image
	      image: gcr.io/example/examplefunction:v1.0.1
	  annotations:
	    config.kubernetes.io/local-config: "true" # tools should ignore this
	spec:
	  configField: configValue

  In the preceding example, 'kustomize config run example/' would identify the function by
  the metadata.configFn field.  It would then write all Resources in the directory to
  a container stdin (running the gcr.io/example/examplefunction:v1.0.1 image).  It
  would then writer the container stdout back to example/, replacing the directory
  file contents.

  See `kustomize config help docs-fn` for more details on writing functions.

### Examples

kustomize config run example/
