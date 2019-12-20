# Validation

This is an example of implementing a validation function against
[kubeval](https://github.com/instrumenta/kubeval).

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

The function is implemented as a go program, which reads a collection of input
Resource configuration, passing each to kubeval.

### Function configuration

A number of settings can be modified for `kubeval` in the function `spec`. See
the `API` struct definition in [main.go](image/main.go) for documentation.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.configFn` and running:

    kustomize config run local-resource/

This exists non-zero if kubeval detects an invalid Resource.

## Running the Example

Run the validator with:

    kustomize config run local-resource/

This will return an error:

    Resource invalid: (Kind: Service, Name: svc)
    prots: Additional property prots is not allowed
    Error: exit status 1

Now fix the typo in [example-use.yaml](local-resource/example-use.yaml) and
run:

    kustomize config run local-resource/

This will return success (no output).
