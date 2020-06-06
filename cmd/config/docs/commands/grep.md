## grep

[Alpha] Search for matching Resources in a directory or from stdin

### Synopsis

[Alpha] Search for matching Resources in a directory or from stdin.

  QUERY:
    Query to match expressed as 'path.to.field=value'.
    Maps and fields are matched as '.field-name' or '.map-key'
    List elements are matched as '[list-elem-field=field-value]'
    The value to match is expressed as '=value'
    '.' as part of a key or value can be escaped as '\.'

  DIR:
    Path to local directory.

### Examples

    # find Deployment Resources
    kustomize cfg grep "kind=Deployment" my-dir/

    # find Resources named nginx
    kustomize cfg grep "metadata.name=nginx" my-dir/

    # use tree to display matching Resources
    kustomize cfg grep "metadata.name=nginx" my-dir/ | kustomize cfg tree

    # look for Resources matching a specific container image
    kustomize cfg grep "spec.template.spec.containers[name=nginx].image=nginx:1\.7\.9" my-dir/ | kustomize cfg tree
