## grep

Search for matching Resources in a directory or from stdin

### Synopsis

  Search for matching Resources in a directory or from stdin.

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
    kyaml grep "kind=Deployment" my-dir/
    
    # find Resources named nginx
    kyaml grep "metadata.name=nginx" my-dir/
    
    # use tree to display matching Resources
    kyaml grep "metadata.name=nginx" my-dir/ | kyaml tree
    
    # look for Resources matching a specific container image
    kyaml grep "spec.template.spec.containers[name=nginx].image=nginx:1\.7\.9" my-dir/ | kyaml tree