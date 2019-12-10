## Function Basics

### Synopsis

  `kustomize config` enables encapsulating function for manipulating Resource
  configuration inside containers, which are run using `run`.

  First fetch the kustomize repository, which contains a collection of example
  functions

	git clone https://github.com/kubernetes-sigs/kustomize
	cd kustomize/functions/examples/

### Templating -- CockroachDB

  This section demonstrates how to leverage templating based solutions from
  `kustomize config`.  The templating function is implemented as a `bash` script
  using a `heredoc`.

  #### 1: Generate the Resources

  `cd` into the `kustomize/functions/examples/template-heredoc-cockroachdb/`
  directory, and invoke `run` on the `local-resource/` directory.

    cd template-heredoc-cockroachdb/

    # view the Resources
    kustomize config tree local-resource/ --name --image --replicas

    # run the function
    kustomize config run local-resource/

    # view the generated Resources
    kustomize config tree local-resource/ --name --image --replicas

  `run` generated the directory ` local-resource/config` containing the generated
  Resources.

  #### 2. Modify the Generated Resources

  - modify the generated Resources by adding an annotation, sidecar container, etc.
  - modify the `local-resources/example-use.yaml` by changing the replicas

  re-run `run`.  this will apply the updated replicas to the generated Resources,
  but keep the fields that you manually added to the generated Resource configuration.

    # run the function
    kustomize config run local-resource/

  `run` facilitates a non-destructive *smart templating* approach that allows templating
  to be composed with manual modifications directly to the template output, as well as
  composition with other functions which may appy validation or injection of values.

  #### 3. Function Implementation

  the function implementation is located under the `image/` directory as a `Dockerfile`
  and a `bash` script.

### Templating -- Nginx

  The steps in this section are identical to the CockroachDB templating example,
  but the function implementation is very different, and implemented as a `go`
  program rather than a `bash` script.

  #### 1: Generate the Resources

  `cd` into the `kustomize/functions/examples/template-go-nginx/`
  directory, and invoke `run` on the `local-resource/` directory.

    cd template-go-nginx/

    # view the Resources
    kustomize config tree local-resource/ --name --image --replicas

    # run the function
    kustomize config run local-resource/

    # view the generated Resources
    kustomize config tree local-resource/ --name --image --replicas

  `run` generated the directory ` local-resource/config` containing the generated
  Resources.  this time it put the configuration in a single file rather than multiple
  files.  The mapping of Resources to files is controlled by the function itself through
  annotations on the generated Resources.

  #### 2. Modify the Generated Resources

  - modify the generated Resources by adding an annotation, sidecar container, etc.
  - modify the `local-resources/example-use.yaml` by changing the replicas

  re-run `run`.  this will apply the updated replicas to the generated Resources,
  but keep the fields that you manually added to the generated Resource configuration.

    # run the function
    kustomize config run local-resource/

  Just like in the preceding section, the function is implemented using a non-destructive
  approach which merges the generated Resources into previously generated instances.

  #### 3. Function Implementation

  the function implementation is located under the `image/` directory as a `Dockerfile`
  and a `go` program.

### Validation -- resource reservations

  This section uses `run` to perform validation rather than generate Resources.

  #### 1: Run the Validator

  `cd` into the `kustomize/functions/examples/validator-resource-requests`
  directory, and invoke `run` on the `local-resource/` directory.

    # run the function
    kustomize config run local-resource/
    cpu-requests missing for a container in Deployment nginx (example-use.yaml [1])
    Error: exit status 1
    Usage:
    ...

  #### 2: Fix the validation issue

  The command will fail complaining that the nginx Deployment is missing `cpu-requests`,
  and print the name of the file + Resource index.  Edit the file and uncomment the resources,
  then re-run the functions.

    kustomize config run local-resource/

  The validation now passes.

### Injection -- resource reservations

  This section uses `run` to perform injection of field values based off annotations
  on the Resource.

  #### 1: Run the Injector

  `cd` into the `kustomize/functions/examples/inject-tshirt-sizes`
  directory, and invoke `run` on the `local-resource/` directory.

    # print the resources
    kustomize config tree local-resource --resources --name
    local-resource
    ├── [example-use.yaml]  Validator
    └── [example-use.yaml]  Deployment nginx
        └── spec.template.spec.containers
            └── 0
                └── name: nginx

    # run the functions
    kustomize config run local-resource/

    # print the new resources
    kustomize config tree local-resource --resources --name
    ├── [example-use.yaml]  Validator
    └── [example-use.yaml]  Deployment nginx
        └── spec.template.spec.containers
            └── 0
                ├── name: nginx
                └── resources: {requests: {cpu: 4, memory: 1GiB}}

  #### 2: Change the tshirt-size

  Change the `tshirt-size` annotation from `medium` to `small` and re-run the functions.

    kustomize config run local-resource/
    kustomize config tree local-resource/
    local-resource
    ├── [example-use.yaml]  Validator
    └── [example-use.yaml]  Deployment nginx
        └── spec.template.spec.containers
            └── 0
                ├── name: nginx
                └── resources: {requests: {cpu: 200m, memory: 50MiB}}

  The function has applied the reservations for the new tshirt-size

### Function Composition

Functions may be composed together.  Try putting the Injection (tshirt-size) and
Validation functions together in the same .yaml file (separated by `---`).  Run
`run` and observe that the first function in the file is applied to the Resources,
and then the second function in the file is applied.
