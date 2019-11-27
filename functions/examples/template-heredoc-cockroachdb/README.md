# Template HereDoc

This is an example of implementing a template function using a heredoc.

This example uses the simplest approach for building abstractions.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.
    
The template is implemented as a heredoc, which substitutes environment variables
into a static string.

This simple implementation uses `kustomize config run wrap --` to perform the
heavy lifting of implementing the function interface.

- parse functionConfig from stdin into environment variables
- merge script output with items from stdin

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.configFn` and running:

    kustomize config run local-resources/
    
This generates the `local-resources/config` directory containing the template output.

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
    
Add an annotation to the StatefulSet Resource and change the replica count of the
`kind: CockroachDB` Resource in `example-use.yaml`.  Rerun the template:

    kustomize config run local-resource/
    
The replica count should be updated, but your annotation should remain.