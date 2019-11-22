## cat

Print Resource Config from a local directory.

### Synopsis

Print Resource Config from a local directory.

  DIR:
    Path to local directory.

### Examples

    # print Resource config from a directory
    kyaml cat my-dir/
    
    # wrap Resource config from a directory in an ResourceList
    kyaml cat my-dir/ --wrap-kind ResourceList --wrap-version config.kubernetes.io/v1alpha1 --function-config fn.yaml
    
    # unwrap Resource config from a directory in an ResourceList
    ... | kyaml cat
