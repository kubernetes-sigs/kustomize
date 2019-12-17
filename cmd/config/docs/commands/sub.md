## set

[Alpha] Set values on Resources fields by substituting values.

### Synopsis

Set values on Resources fields by substituting predefined markers for new values.

`set` looks for markers specified on Resource fields and substitute a new user defined
value for the existing value.

`set` maybe be used to:

- edit configuration programmatically from the cli or scripts
- create reusable bundles of configuration

  DIR

    A directory containing Resource configuration.

  NAME

    Optional.  The name of the substitution to perform or display.

  VALUE

    Optional.  The new value to substitute into the field.


To print the possible substitutions for the Resources in a directory, run `set` on
a directory -- e.g. `kustomize config set DIR/`.

#### Tips

- A description of the value may be specified with `--description`.
- An owner for the field's value may be defined with `--owned-by`.
- Prevent overriding previous substitutions with `--override=false`.
- Revert previous substitutions with `--revert`.
- Create substitutions on Kustomization.yaml's, patches, etc

When overriding or reverting previous substitutions, the description and owner are left
unmodified unless specified with flags.

To create a substitution for a field see: `kustomize help config set create`

### Examples

  Resource YAML: Name substitution

    # dir/resources.yaml
    ...
    metadata:
        name: PREFIX-app1 # {"substitutions":[{"name":"prefix","marker":"PREFIX-"}]}
    ...
    ---
    ...
    metadata:
        name: PREFIX-app2 # {"substitutions":[{"name":"prefix","marker":"PREFIX-"}]}
    ...

  Show substitutions: Show the possible substitutions

    $ config set dir
    NAME    DESCRIPTION    VALUE     TYPE    COUNT   SUBSTITUTED   OWNER
    prefix   ''            PREFIX-   string   2       false

  Perform substitution: set a new value, owner and description

    $ config set dir prefix "test-" --description "test environment" --owned-by "dev"
    performed 2 substitutions

  Show substitutions: Show the new values

    $ config set dir
    NAME       DESCRIPTION       VALUE    TYPE    COUNT   SUBSTITUTED   OWNER
    prefix   'test environment'   test-   string   2       true          dev

  New Resource YAML:

    # dir/resources.yaml
    ...
    metadata:
      name: test-app1 # {"substitutions":[{"name":"prefix","marker":"PREFIX-","value":"test-"}],"setBy":"dev","description":"test environment"}
    ...
    ---
    ...
    metadata:
      name: test-app2 # {"substitutions":[{"name":"prefix","marker":"PREFIX-","value":"test-"}],"setBy":"dev","description":"test environment"}
    ...

  Revert substitution:

    config set dir prefix --revert
    performed 2 substitutions

    config set dir
    NAME       DESCRIPTION       VALUE    TYPE    COUNT   SUBSTITUTED   OWNER  
    prefix   'test environment'   PREFIX-  string   2       false          dev    
