## tree

[Alpha] Display Resource structure from a directory or stdin.

### Synopsis

[Alpha] Display Resource structure from a directory or stdin.

kustomize config tree may be used to print Resources in a directory or cluster, preserving structure

Args:

  DIR:
    Path to local directory directory.

Resource fields may be printed as part of the Resources by specifying the fields as flags.

kustomize config tree has build-in support for printing common fields, such as replicas, container images,
container names, etc.

kustomize config tree supports printing arbitrary fields using the '--field' flag.

By default, kustomize config tree uses the directory structure for the tree structure, however when printing
from the cluster, the Resource graph structure may be used instead.

### Examples

    # print Resources using directory structure
    kustomize config tree my-dir/
    
    # print replicas, container name, and container image and fields for Resources
    kustomize config tree my-dir --replicas --image --name
    
    # print all common Resource fields
    kustomize config tree my-dir/ --all
    
    # print the "foo"" annotation
    kustomize config tree my-dir/ --field "metadata.annotations.foo" 
    
    # print the "foo"" annotation
    kubectl get all -o yaml | kustomize config tree my-dir/ --structure=graph \
      --field="status.conditions[type=Completed].status"
    
    # print live Resources from a cluster using graph for structure
    kubectl get all -o yaml | kustomize config tree --replicas --name --image --structure=graph
    
    
    # print live Resources using graph for structure
    kubectl get all,applications,releasetracks -o yaml | kustomize config tree --structure=graph \
      --name --image --replicas \
      --field="status.conditions[type=Completed].status" \
      --field="status.conditions[type=Complete].status" \
      --field="status.conditions[type=Ready].status" \
      --field="status.conditions[type=ContainersReady].status"
