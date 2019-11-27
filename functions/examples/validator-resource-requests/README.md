# Validation

This is an example of implementing a validation function.

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.
    
The template is implemented as a go program, which reads a collection of input
Resource configuration, and looks for invalid configuration.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.configFn` and running:

    kustomize config run local-resources/
    
This exists non-zero if there is an error.

## Running the Example

Run the validator with:

    kustomize config run local-resource/
    
This will return an error:

    cpu-requests missing for container nginx

Now uncomment the resource reservations and run again:

    kustomize config run local-resource/
    
This will return success