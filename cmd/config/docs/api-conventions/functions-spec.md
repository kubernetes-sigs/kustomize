# KRM Functions Specification

_apiVersion: v1_

## Overview

This document specifies a standard for client-side functions that operate on
Kubernetes declarative configurations referred to as _KRM Functions_. This
standard enables creating small, interoperable, and language-independent
executable programs packaged as containers that can be chained together as part
of a configuration management pipeline. The end result of such a pipeline are
fully rendered configurations that can then be applied to a control plane (e.g.
Using ‘kubectl apply’ for Kubernetes control plane). As such, although this
document references Kubernetes Resource Model and API conventions, it is
completely decoupled from Kubernetes API machinery and does not depend on any
in-cluster components.

This document references terms described in [Kubernetes API Conventions][1].

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in [RFC 2119][2].

## Use Cases

KRM functions enable shift-left practices (client-side) through:

- Pre-commit / delivery validation and linting of configuration
  - e.g. Fail if any containers don't have PodSecurityPolicy or CPU / Memory
    limits
- Implementation of abstractions as client actuated APIs
  - e.g. Create a client-side _"CRD"_ for generating configuration checked into
    git
- Injection of cross-cutting configuration
  - e.g. T-Shirt size containers by annotating resources with `small`, `medium`,
    `large` and inject the cpu and memory resources into containers accordingly.
  - e.g. Inject `init` and `side-car` containers into resources based off of
    resource type, annotations, etc.

Performing these on the client rather than the server enables:

- Configuration to be reviewed prior to being sent to the API server
- Configuration to be validated as part of the CI/CD pipeline
- Configuration for resources to validated holistically rather than individually
  per-resource
  - e.g. ensure the `Service.selector` and `Deployment.spec.template` labels
    match.
  - e.g. MutatingWebHooks are scoped to a single resource instance at a time.
- Low-level tweaks to the output of high-level abstractions
  - e.g. add an `init container` to a client _"CRD"_ resource after it was
    generated.
- Composition and layering of multiple functions together
  - Compose generation, injection, validation together

## Definitions

- **function:** A containerized program conforming to the spec described in this
  document.
- **orchestrator:** A program that invokes the function container, passing
  arguments and processing its output.

## Interface

The inter-process communication between the orchestrator and a function works as
follows:

1. Orchestrator runs the function container and provides the input on `stdin`.
   The input is a Kubernetes object of kind `ResourceList` as described below.
2. Function reads the input from `stdin`, performs computations, and provides
   the output as a `ResourceList` to `stdout`. The function MAY also emit
   non-structured error message on `stderr`.
3. Orchestrator uses the `stdout`, `stderr`, and the exit code of the function
   as it sees fit following to the semantics described below.

### Breaking Changes

1. KRM Function spec `v2` will drop support for `results[].field.proposedValue`
   and instead use the new array field `results[].field.proposedValues[]`.

### Schema

A function MUST accept input from `stdin` and MUST output to `stdout` a
Kubernetes object of kind `ResourceList` with the following OpenAPI schema:

```yaml
swagger: "2.0"
info:
  title: KRM Functions Specification (ResourceList)
  version: v1
definitions:
  ResourceList:
    type: object
    description: ResourceList is the input/output wire format for KRM functions.
    x-kubernetes-group-version-kind:
      - group: config.kubernetes.io
        kind: ResourceList
        version: v2alpha1
      - group: config.kubernetes.io
        kind: ResourceList
        version: v1
      - group: config.kubernetes.io
        kind: ResourceList
        version: v1beta1
    required:
      - items
    properties:
      apiVersion:
        description: apiVersion of ResourceList
        type: string
      kind:
        description: kind of ResourceList i.e. `ResourceList`
        type: string
      items:
        type: array
        description: |
          [input/output]
          Items is a list of Kubernetes objects: 
          https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#types-kinds).

          A function will read this field in the input ResourceList and populate
          this field in the output ResourceList.
        items:
          type: object
      functionConfig:
        type: object
        description: |
          [input]
          FunctionConfig is an optional Kubernetes object for passing arguments to a
          function invocation.
      results:
        type: array
        description: |
          [output]
          Results is an optional list that can be used by function to emit results
          for observability and debugging purposes.
        items:
          "$ref": "#/definitions/Result"
  Result:
    type: object
    required:
      - message
    properties:
      message:
        type: string
        description: Message is a human readable message.
      severity:
        type: string
        enum:
          - error
          - warning
          - info
        default: error
        description: |
          Severity is the severity of a result:

          "error": indicates an error result.
          "warning": indicates a warning result.
          "info": indicates an informational result.
      resourceRef:
        type: object
        description: |
          ResourceRef is the metadata for referencing a Kubernetes object
          associated with a result.
        required:
          - apiVersion
          - kind
          - name
        properties:
          apiVersion:
            description:
              APIVersion refers to the `apiVersion` field of the object
              manifest.
            type: string
          kind:
            description: Kind refers to the `kind` field of the object.
            type: string
          namespace:
            description:
              Namespace refers to the `metadata.namespace` field of the object
              manifest.
            type: string
          name:
            description:
              Name refers to the `metadata.name` field of the object manifest.
            type: string
      field:
        type: object
        description: |
          Field is the reference to a field in the object.
          If defined, `ResourceRef` must also be provided.
        required:
          - path
        properties:
          path:
            type: string
            description: |
              Path is the JSON path of the field
              e.g. `spec.template.spec.containers[3].resources.limits.cpu`
          currentValue:
            description: |
              CurrentValue is the current value of the field.
              Can be any value - string, number, boolean, array or object.
          proposedValues:
            type: array
            description: |
              ProposedValues is the set of proposed values of the field to fix an issue.
              Elements can be any value - string, number, boolean, array or object.

              Note: Only supported on spec v2alpha1 and above.
          proposedValue:
            description: |
              ProposedValue is the proposed value of the field to fix an issue.
              Can be any value - string, number, boolean, array or object.
              
              Note: Only supported on spec v1 and below. Use proposedValues instead.
      file:
        type: object
        description: File references a file containing the resource.
        required:
          - path
        properties:
          path:
            type: string
            description: |
              Path is the OS agnostic, slash-delimited, relative path.
              e.g. `some-dir/some-file.yaml`.
          index:
            type: number
            default: 0
            description: Index of the object in a multi-object YAML file.
      tags:
        type: object
        additionalProperties:
          type: string
        description: |
          Tags is an unstructured key value map stored with a result that may be set
          by external tools to store and retrieve arbitrary metadata.
paths: {}
```

#### Examples

The following is an example input, where the custom resource of kind
`FulfillmentCenter` is provided as `functionConfig`. The function will operate
on one resource of kind `Service`.

```yaml
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
  spec:
    address: "100 Main St."
items:
  - apiVersion: v1
    kind: Service
    metadata:
      name: wordpress
      labels:
        app: wordpress
      annotations:
        internal.config.kubernetes.io/index: "0"
        internal.config.kubernetes.io/path: "service.yaml"
    spec: # Example comment
      type: LoadBalancer
      selector:
        app: wordpress
        tier: frontend
      ports:
        - protocol: TCP
          port: 80
```

The following is an example output containing one result representing a
validation error which also includes a proposal to fix the problem, tools that
make use of this API can use these hints to apply automatic fixes.

```yaml
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
  - apiVersion: v2alpha1
    kind: Service
    metadata:
      name: wordpress
      labels:
        app: wordpress
      annotations:
        internal.config.kubernetes.io/index: "0"
        internal.config.kubernetes.io/path: "service.yaml"
    spec: # Example comment
      type: LoadBalancer
      selector:
        app: wordpress
        tier: frontend
      ports:
        - protocol: TCP
          port: 80
results:
  - message: "Invalid type. Expected: integer, given: string"
    severity: error
    resourceRef:
      apiVersion: v1
      kind: Service
      name: wordpress
    field:
      path: spec.ports.0.port
      currentValue: cool-port
      proposedValues:
      - 1337
      - 4242
    file:
      path: service.yaml
```

### Serialization Format

A function MUST support YAML as a serialization format for the input and output.
A function MUST use utf8 encoding (as YAML is a superset of JSON, JSON will also
be supported by any conforming function).

### Containerization

A function MUST be implemented as a container.

A function container MUST be capable of running as a non-root user `nobody` if
it does not require access to host filesystem.

### stderr

Any non-structured error messages MUST be emitted to `stderr`. `stdout` is
reserved for `ResourceList` as described above.

### Exit Code

An exit code of zero indicates function execution was successful. A non-zero
exit code indicates a failure.

### Operations

A function MAY Create, Update, or Delete any number of items in the `items`
field and output the resultant list in the corresponding `items` field of the
output.

A function SHOULD preserve comments when input serialization format is YAML.
This allows for human authoring of configuration to coexist with changes made by
functions.

### Internal Annotations

For orchestration purposes, the orchestrator will use a set of annotations,
referred to as _internal annotations_, on resources in `Resources.items`. These
annotations are not persisted to resource manifests on the filesystem: The
orchestrator sets this annotation when reading files from the local filesystem
and removes the annotation when writing the output of functions back to the
filesystem.

Annotation prefix `internal.config.kubernetes.io` is reserved for use for
internal annotations. In general, a function MUST NOT modify these annotations with
the exception of the specific annotations listed below. This enables orchestrators to add additional internal annotations, without requiring changes to existing functions.

#### `internal.config.kubernetes.io/path`

Records the slash-delimited, OS-agnostic, relative file path to a resource. The
path is relative to a fix location on the filesystem. Different orchestrator
implementations can choose different fixed points.

A function SHOULD NOT modify these annotations.

Example:

```yaml
metadata:
  annotations:
    internal.config.kubernetes.io/path: "relative/file/path.yaml"
```

#### `internal.config.kubernetes.io/index`

Records the index of a Resource in file. In a multi-object YAML file, resources
are separated by three dashes (`---`), and the index represents the position of
the Resource starting from zero. When this annotation is not specified, it
implies a value of `0`.

A function SHOULD NOT modify these annotations.

Example:

```yaml
metadata:
  annotations:
    internal.config.kubernetes.io/path: "relative/file/path.yaml"
    internal.config.kubernetes.io/index: 2
```

This represents the third resource in the file.

[1]:
  https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[2]: https://tools.ietf.org/html/rfc2119
[3]:
  https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#types-kinds
