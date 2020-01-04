## set

[Alpha] Set values on Resources fields values.

### Synopsis

Set values on Resources fields.  May set either the complete or partial field value.

`set` identifies setters using field metadata published as OpenAPI extensions.
`set` parses both the Kubernetes OpenAPI, as well OpenAPI published inline in
the configuration as comments.

`set` maybe be used to:

- edit configuration programmatically from the cli
- create reusable bundles of configuration with custom setters

  DIR

    A directory containing Resource configuration.

  NAME

    Optional.  The name of the setter to perform or display.

  VALUE

    Optional.  The value to set on the field.


To print the possible setters for the Resources in a directory, run `set` on
a directory -- e.g. `kustomize config set DIR/`.

#### Tips

- A description of the value may be specified with `--description`.
- The last setter for the field's value may be defined with `--set-by`.
- Create custom setters on Resources, Kustomization.yaml's, patches, etc

The description and setBy fields are left unmodified unless specified with flags.

To create a custom setter for a field see: `kustomize help config create-setter`

### Examples

  Resource YAML: Name Prefix Setter

    # DIR/resources.yaml
    ...
    metadata:
        name: PREFIX-app1 # {"type":"string","x-kustomize":{"partialFieldSetters":[{"name":"name-prefix","value":"PREFIX"}]}}
    ...
    ---
    ...
    metadata:
        name: PREFIX-app2 # {"type":"string","x-kustomize":{"partialFieldSetters":[{"name":"name-prefix","value":"PREFIX"}]}}
    ...

  List setters: Show the possible setters

    $ config set DIR/
        NAME      DESCRIPTION   VALUE     TYPE     COUNT   SETBY  
    name-prefix   ''            PREFIX    string   2

  Perform set: set a new value, owner and description

    $ kustomize config set DIR/ name-prefix "test" --description "test environment" --set-by "dev"
    set 2 values

  List setters: Show the new values

    $ config set DIR/
        NAME      DESCRIPTION         VALUE     TYPE     COUNT     SETBY 
    name-prefix   'test environment'   test     string   2          dev

  New Resource YAML:

    # DIR/resources.yaml
    ...
    metadata:
        name: test-app1 # {"description":"test environment","type":"string","x-kustomize":{"setBy":"dev","partialFieldSetters":[{"name":"name-prefix","value":"test"}]}}
    ...
    ---
    ...
    metadata:
        name: test-app2 # {"description":"test environment","type":"string","x-kustomize":{"setBy":"dev","partialFieldSetters":[{"name":"name-prefix","value":"test"}]}}
    ...
