## tree

[Alpha] Display Resource structure from a directory or stdin.

### Synopsis

[Alpha] Display Resource structure from a directory or stdin.

kustomize cfg tree may be used to print Resources in a directory or cluster, preserving structure

Args:

  DIR:
    Path to local directory directory.

Resource fields may be printed as part of the Resources by specifying the fields as flags.

kustomize cfg tree has build-in support for printing common fields, such as replicas, container images,
container names, etc.

kustomize cfg tree supports printing arbitrary fields using the '--field' flag.

By default, kustomize cfg tree uses Resource graph structure if any relationships between resources (ownerReferences)
are detected, as is typically the case when printing from a cluster. Otherwise, directory graph structure is used. The
graph structure can also be selected explicitly using the '--graph-structure' flag.

### Examples

    # print Resources using directory structure
    kustomize cfg tree my-dir/

    # print replicas, container name, and container image and fields for Resources
    kustomize cfg tree my-dir --replicas --image --name

    # print all common Resource fields
    kustomize cfg tree my-dir/ --all

    # print the "foo"" annotation
    kustomize cfg tree my-dir/ --field "metadata.annotations.foo"

    # print the "foo"" annotation
    kubectl get all -o yaml | kustomize cfg tree \
      --field="status.conditions[type=Completed].status"

    # print live Resources from a cluster using owners for graph structure
    kubectl get all -o yaml | kustomize cfg tree --replicas --name --image

    # print live Resources with status condition fields
    kubectl get all -o yaml | kustomize cfg tree \
      --name --image --replicas \
      --field="status.conditions[type=Completed].status" \
      --field="status.conditions[type=Complete].status" \
      --field="status.conditions[type=Ready].status" \
      --field="status.conditions[type=ContainersReady].status"
