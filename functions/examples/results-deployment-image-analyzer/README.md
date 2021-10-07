# Validation

This is an example of implementing an "analyzer" KRM function that returns a
`ResourceList` with a list of `Results` (or findings) instead of erroring out.

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

The template is implemented as a go program, which reads a collection of input
Resource configuration, and looks for valid configuration with common mistakes,
instead of failing it returns a processed `ResourceList` with `Results` with a
summary of all the findings in the given configuration.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize fn run local-resource/

This exists non-zero if there is an unexpected error.

## Running the Example

Run the analyzer with:

    kustomize fn run local-resource/ --dry-run

This will generate `Results` for deployments missing a version on the image.