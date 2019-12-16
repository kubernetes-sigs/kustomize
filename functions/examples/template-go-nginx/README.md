# Template GoTemplate

This is an example of implementing a template function using a `go` template.

This example uses a more sophisticated approach for building abstractions.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

The template is implemented as a go program with a `go` template.  It parses the
functionConfig into a `go` struct and uses the `kyaml` module for parsing the
function input, and writing the function output.

1. read the inputs (stdin)
2. apply filters
  - (custom filter) generate the nginx Deployment and Service from go templates using the
    functionConfig as the template input.
  - merge the generated Deployment and Service with the input Deployment and Service if
    present
  - set filenames on the Resources (`NAME.yaml`)
  - format the Resources
3. write the outputs (stdout)

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.configFn` and running:

    kustomize config run local-resource/

This generates the `local-resource/config` directory containing the template output.

- the template output may be modified by adding fields -- such as initContainers,
  sidecarConatiners, cpu resource limits, etc -- and these fields will be retained
  when re-running `run`
- the function input `example-use.yaml` may be changed and rerunning `run` will update
  only the parts changed in the template output.

## Running the Example

Run the config with:

     kustomize config run local-resource/

This will create the directory

    local-resource/config

Add an annotation to the Deployment Resource and change the replica count of the
`kind: Nginx` Resource in `example-use.yaml`.  Rerun the template:

    kustomize config run local-resource/

The replica count should be updated, but your annotation should remain.
