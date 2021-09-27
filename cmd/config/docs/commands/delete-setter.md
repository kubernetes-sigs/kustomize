## delete-setter

[Alpha] Delete a custom setter for a Resource field

### Synopsis

Delete a custom setter for a Resource field.

  DIR

    A directory containing Resource configuration.

  NAME

    The name of the setter to create.

### Deleting a Custom Setter

**Given the YAML:**

    # resource.yaml
    apiVersion: v1
    kind: Service
    metadata:
      ...
    spec:
      ...
      ports:
        ...
      - name: http
        port: 8080  # {"type":"integer","x-kustomize":{"partialFieldSetters":[{"name":"http-port","value":"8080"}]}}
        ...

**Delete setter:**

    # delete a setter for ports
    $ kustomize cfg set create DIR/ http-port

comment will be removed for this field is not settable any more.

**Newly modified YAML:**

    # resource.yaml
    apiVersion: v1
    kind: Service
    metadata:
      ...
    spec:
      ...
      ports:
        ...
      - name: http
        port: 8080
        ...


### Deleting a setter used in substitution

If the setter is also used in substitution, it will ask you to delete the substitution first.


### Examples

    # delete a setter for port
    kustomize cfg create-setter DIR/ port