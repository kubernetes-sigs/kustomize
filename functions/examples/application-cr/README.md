# Append Application CR

This is an example of appending an Application CR to a group of resources.

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

The template is implemented as a go program, which reads a collection of input
Resource configuration, and looks for invalid configuration.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize config run local-resource/ --fn-path config/

This exits non-zero if there is an error.

## Running the Example

Run the validator with:

    kustomize config run local-resource/ --fn-path config/

This will append an Application CR.  
