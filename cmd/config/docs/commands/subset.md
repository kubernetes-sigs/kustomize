## sub-set-marker

[Alpha] Create a new substitution for a Resource field

### Synopsis

Create a new substitution for a Resource field -- recognized by `kustomize config set`.

  DIR

    A directory containing Resource configuration.

  NAME

    The name of the substitution to create.

  VALUE

    The current value of the field, or a substring of the field.

#### Tips: Picking Good Marker

Substitutions may be defined by directly editing yaml **or** by running `kustomize config set create`
to create a new substitution.

Given the YAML:
    
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

Create a new set marker:

    # create a substitution for ports
    $ kustomize config set create dir/ http-port 8080 --type "int" --field "port"

Modified YAML:

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
        port: 8080 # {"substitutions":[{"name":"port","marker":"[MARKER]"}],"type":"int"}
        ...

Change the value using the `set` command:

    # change the http-port value to 8081
    $ kustomize config set dir/ http-port 8081

Resources fields with a field name matching `--field` and field value matching `VALUE` will
have a line comment added marking this field as settable.

Substitution markers may be:

- valid field values (e.g. `8080` for a port)
  - Note: `008080` would be preferred because it is more recognizable as a marker
- invalid values that adhere to the schema (e.g. `0000` for a port)
- values that do not adhere to the schema (e.g. `[PORT]` for port)

Markers **SHOULD be clearly identifiable as a marker and either**:

- **adhere to the field schema** -- e.g. use a valid value


    port: 008080 # {"substitutions":[{"name":"port","marker":"008080"}],"type":"int"}

- **be pre-filled in with a value** -- e.g. set the value when setting the marker


    port: 8080 # {"substitutions":[{"name":"port","marker":"[MARKER]","value":"8080""}],"type":"int"}

**Note:** The important thing is that in both cases the Resource configuration may be directly
applied to a cluster and validated by tools without the tool knowing about the substitution
marker.

The difference between the preceding examples is that:

- the former will be shown as `SUBSTITUTED=false` (`config sub dir/` exits non-0)
- the latter with show up as `SUBSTITUTED=true` (`config sub dir/` exits 0)

When choosing the which to use, consider that checks for unsubstituted values MAY be
configured as pre-commit checks -- if you want to these checks to fail if the value
hasn't been substituted, then don't specify a `value`.

Markers which are invalid field values MAY be chosen in cases where it is preferred to have
the create or update request fail rather than succeed if the substitution has not yet been
performed.

A substitution may be a substring of the full field:

    $ kustomize config set create dir/ app-image-tag v1.0.01 --type "string" --field "image"

    image: gcr.io/example/app:v1.0.1 # {"substitutions":[{"name":"app-image-tag","marker":"[MARKER]","value":"v1.0.1"}]}


A single field value may have multiple substitutions applied to it:

    name: PREFIX-app-SUFFIX # {"substitutions":[{"name":"prefix","marker":"PREFIX-"},{"name":"suffix","marker":"-SUFFIX"}]}

#### Substitution Format

Substitutions are defined as json encoded FieldMeta comments on fields.

FieldMeta Schema read by `sub`:

    {
      "title": "FieldMeta",
      "type": "object",
      "properties": {
        "substitutions": {
          "type": "array",
          "description": "Possible substitutions that may be performed against this field.",
          "items": {
            "type": "object",
            "properties": {
              "name": "Name of the substitution.",
              "marker": "Marker for the value to be substituted.",
              "value": "Current substituted value"
            }   
          }
        },
        "type": {
          "type": "string",
          "description": "The value type.  Defaults to string."
          "enum": ["string", "int", "float", "bool"]
        },
        "description": {
          "type": "string",
          "description": "A description of the field's current value.  Optional."
        },
        "setBy": {
          "type": "string",
          "description": "The current owner of the field.  Optional."
        },
      }
    }

### Examples

    # set a substitution for port fields matching "8080"
    kustomize config sub create dir/ port 8080 --type "int" --field port \
         --description "default port used by the app"

    # set a substitution for port fields matching "8080", using "0000" as a marker.
    kustomize config sub dir/ port 8080 --marker "0000" --type "int" \
        --field port --description "default port used by the app"

    # substitute a substring of a field rather than the full field -- e.g. only the
    # image tag, not the full image
    kustomize config sub dir/ app-image-tag v1.0.1 --type "string" --substring \
        --field port --description "current stable release"