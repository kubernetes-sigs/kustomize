## annotate

[Alpha] Set an annotation on Resources.

### Synopsis

[Alpha]  Set an annotation on Resources.

  DIR:
    Path to local directory.

### Examples

    kustomize config annotate my-dir/ --kv foo=bar

    kustomize config annotate my-dir/ --kv foo=bar --kv a=b

    kustomize config annotate my-dir/ --kv foo=bar --kind Deployment --name foo
