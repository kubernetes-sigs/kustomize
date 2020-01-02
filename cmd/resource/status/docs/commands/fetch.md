## fetch

[Alpha] Fetch the state of the provided resources from the cluster and display status in a table.

### Synopsis

[Alpha] Fetches the state of all provided resources from the cluster and displays the status in
a table.

The list of resources are provided as manifests either on the filesystem or on StdIn. 

  DIR:
    Path to local directory.

### Examples

    # Read resources from the filesystem and wait up to 1 minute for all of them to become Current
    resource status fetch my-dir/

    # Fetch all resources in the cluster and wait up to 5 minutes for all of them to become Current
    kubectl get all --all-namespaces -oyaml | resource status fetch
