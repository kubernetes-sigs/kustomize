## cat

[Alpha] Print Resource Config from a local directory.

### Synopsis

[Alpha]  Print Resource Config from a local directory.

  DIR:
    Path to local directory.

### Examples

    # print Resource config from a directory
    kustomize config cat my-dir/
    
    # wrap Resource config from a directory in an ResourceList
    kustomize config cat my-dir/ --wrap-kind ResourceList --wrap-version config.kubernetes.io/v1alpha1 --function-config fn.yaml
    
    # unwrap Resource config from a directory in an ResourceList
    ... | kustomize config cat
