# Configuration Functions API Semantics

  Configuration Functions are containers which enable **shift-left practices**
  for configuring applications and infrastructure through Kubernetes style
  Resource Configuration.
  
  Configuration functions enable shift-left practices (client-side) through:
  
  - Pre-commit / delivery validation and linting of configuration
    - Fail if any containers don't have PodSecurityPolicy or CPU / Memory limits
  - Implementation of abstractions as APIs
    - Create a client-side "CRD" for generating configuration  
  - Injection of cross-cutting configuration
    - Inject init-containers into all Resources with `foo` annotation
  - Aspect Oriented authoring of configuration
    - Set CPU / Memory for containers on Resources annotated with `container-resources=small`

  Performing these through shift-left enables:
  
  - Configuration to be reviewed prior to being sent to the API server
  - Configuration to be validated as part of the CD pipeline
  - Configuration for Resources to validated holistically rather than individually
    per-Resource -- e.g. ensure the `Service.selector` and `Deployment.spec.template` labels
    match.
  - Low-level tweaks to the output of high-level abstractions -- e.g. add an `initContainer`
    to a generated Resource.
  - Composition and layering of multiple functions

  Configuration Functions are implemented as executable programs published in containers which:
   
  - Accept as input (stdin):
    - A list of Resource Configuration
    - A function Configuration (to configure the function itself)
  - Emit as output (stdout + exit):
    - A list of Resource Configuration
    - An exit code for success / failure
  
### Function Specification

  - Functions **SHOULD** be published as container images containing a `CMD` invoking an executable.
  - Functions **MUST** accept input on STDIN a `ResourceList` containing the Resources and
    `functionConfig`.
  - Functions **MUST** emit output on STDOUT a `ResourceList` containing the modified
    Resources.
  - Functions **MUST** exit non-0 on failure, and exit 0 on success.
  - Functions **MAY** emit output on STDERR with error messaging.
  - Functions performing validation **SHOULD** exit failure and emit error messaging
    on a validation failure.
  - Functions generating Resources **SHOULD** retain non-conflicting changes on the
    generated Resources -- e.g. 1. the function generates a Deployment, but doesn't
    specify `cpu`, 2. the user sets the `cpu` on the generated Resource, 3. the
    function should keep the `cpu` when regenerating the Resource a second time.
  - Functions **SHOULD** be usable outside `kyaml run-fns` -- e.g. though pipeline
    mechanisms such as Tekton.

#### Input Format

  Functions must accept on STDIN:

  `ResourceList`:
  - contains `items` field, same as `List.items`
  - contains `functionConfig` field -- a single item with the configuration for the function itself

  Example `ResourceList` Input:
  
    apiVersion: config.kubernetes.io/v1alpha1
    kind: ResourceList
    functionConfig:
      apiVersion: example.com/v1beta1
      kind: Nginx
      metadata:
        name: my-instance
        annotations:
          config.kubernetes.io/local-config: "true"
      spec:
        replicas: 5
    items:
    -  apiVersion: apps/v1
       kind: Deployment
       metadata:
         name: my-instance
       spec:
         replicas: 3
         ...
    - apiVersion: v1
      kind: Service
      metadata:
        name: my-instance
      spec:
        ...

#### Output Format

  Functions must emit on STDOUT:

  `ResourceList`:
  - contains `items` field, same as `List.items`

  Example `ResourceList` Output:
  
    apiVersion: config.kubernetes.io/v1alpha1
    kind: ResourceList
    items:
    -  apiVersion: apps/v1
       kind: Deployment
       metadata:
         name: my-instance
       spec:
         replicas: 5
         ...
    - apiVersion: v1
      kind: Service
      metadata:
        name: my-instance
      spec:
        ...

#### Container Environment

  When run by `kyaml run-fns`, functions are run in containers with the following environment:

  - Network: `none`
  - User: `nobody`
  - Security Options: `no-new-privileges`
  - Volumes: the volume containing the `functionConfig` yaml is mounted under `/local` as `ro`

### Example Function Implementation

  Following is an example for implementing an nginx abstraction using a config
  function.

#### `nginx-template.sh`

  `nginx-template.sh` is a simple bash script which uses a heredoc as a templating solution
  for generating Resources from the functionConfig input fields.

  The script wraps itself using `kyaml run-fns wrap -- $0` which will:

  1. Parse the `ResourceList.functionConfig` (provided to the container stdin) into env vars
  2. Merge the stdout into the original list of Resources
  3. Defaults filenames for newly generated Resources (if they are not set as annotations)
     to `config/NAME_KIND.yaml`
  4. Format the output

    #!/bin/bash
    # script must run wrapped by kyaml for parsing input
    # the functionConfig into env vars
    if [ -z ${WRAPPED} ]; then
      export WRAPPED=true
      kyaml run-fns wrap -- $0
      exit $?
    fi

    cat <<End-of-message
    apiVersion: v1
    kind: Service
    metadata:
      name: ${NAME}
      labels:
        app: nginx
        instance: ${NAME}
    spec:
      ports:
      - port: 80
        targetPort: 80
        name: http
      selector:
        app: nginx
        instance: ${NAME}
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ${NAME}
      labels:
        app: nginx
        instance: ${NAME}
    spec:
      replicas: ${REPLICAS}
      selector:
        matchLabels:
          app: nginx
          instance: ${NAME}
      template:
        metadata:
          labels:
            app: nginx
            instance: ${NAME}
        spec:
          containers:
          - name: nginx
            image: nginx:1.7.9
            ports:
            - containerPort: 80
    End-of-message

#### `Dockerfile`

  `Dockerfile` installs `kyaml` and copies the script into the container image.    

    FROM golang:1.13-stretch
    RUN go get sigs.k8s.io/kustomize/cmd/kyaml
    RUN mv /go/bin/kyaml /usr/bin/kyaml
    COPY nginx-template.sh /usr/bin/nginx-template.sh
    CMD ["nginx-template.sh]

### Example Function Usage

Following is an example of running the `kyaml run-fns` using the preceding API.

#### `nginx.yaml` (Input)

  `dir/nginx.yaml` contains a reference to the Function.  The contents of `nginx.yaml`
  are passed to the Function through the `ResourceList.functionConfig` field.

    apiVersion: example.com/v1beta1
    kind: Nginx
    metadata:
      name: my-instance
      annotations:
        config.kubernetes.io/local-config: "true"
      configFn:
        container:
          image: gcr.io/example-functions/nginx-template:v1.0.0
    spec:
      replicas: 5

  - `configFn.container.image`: the image to use for this API
  - `annotations[config.kubernetes.io/local-config]`: mark this as not a Resource that should
    be applied
    
#### `kyaml run-fns dir/` (Output)

  `dir/my-instance_deployment.yaml` contains the Deployment:
    
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: my-instance
      labels:
        app: nginx
        instance: my-instance
    spec:
      replicas: 5
      selector:
        matchLabels:
          app: nginx
          instance: my-instance
      template:
        metadata:
          labels:
            app: nginx
            instance: my-instance
        spec:
          containers:
          - name: nginx
            image: nginx:1.7.9
            ports:
            - containerPort: 80

  `dir/my-instance_service.yaml` contains the Service:

    apiVersion: v1
    kind: Service
    metadata:
      name: my-instance
      labels:
        app: nginx
        instance: my-instance
    spec:
      ports:
      - port: 80
        targetPort: 80
        name: http
      selector:
        app: nginx
        instance: my-instance
 
