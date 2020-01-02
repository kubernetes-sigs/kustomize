## events

[Alpha] Poll the cluster until all provided resources have become Current and list the status change events.

### Synopsis

[Alpha] Poll the cluster for the state of all the provided resources until either they have all become 
Current or the timeout is reached. The output will be status change events.

The list of resources which should be polled are provided as manifests either on the filesystem or
on StdIn. 

  DIR:
    Path to local directory. If not provided, input is expected on StdIn.

### Examples

    # Read resources from the filesystem and wait up to 1 minute for all of them to become Current
    resource status events my-dir/

    # Fetch all resources in the cluster and wait up to 5 minutes for all of them to become Current
    kubectl get all --all-namespaces -oyaml | resource status events --timeout=5m
