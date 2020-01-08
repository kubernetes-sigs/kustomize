# Configuration Functions Specification

This document specifies a standard for client-side functions that operate on
Kubernetes declarative configurations. This standard enables creating
small, interoperable, and language-independent executable programs packaged as
containers that can be chained together as part of a configuration management pipeline.
The end result of such a pipeline are fully rendered configurations that can then be
applied to a control plane (e.g. Using ‘kubectl apply’ for Kubernetes control plane).
As such, although this document references Kubernetes Resource Model and API conventions,
it is completely decoupled from Kuberentes API machinery and does not depend on any
in-cluster components.

This document references terms described in [Kubernetes API Conventions][1].

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in [RFC 2119][2].

## Use Cases

_Configuration functions_ enable shift-left practices (client-side) through:

- Pre-commit / delivery validation and linting of configuration
  - e.g. Fail if any containers don't have PodSecurityPolicy or CPU / Memory limits
- Implementation of abstractions as client actuated APIs (e.g. templating)
  - e.g. Create a client-side _"CRD"_ for generating configuration checked into git
- Aspect Orient configuration / Injection of cross-cutting configuration
  - e.g. T-Shirt size containers by annotating Resources with `small`, `medium`, `large`
    and inject the cpu and memory resources into containers accordingly.
  - e.g. Inject `init` and `side-car` containers into Resources based off of Resource
    Type, annotations, etc.

Performing these on the client rather than the server enables:

- Configuration to be reviewed prior to being sent to the API server
- Configuration to be validated as part of the CI?CD pipeline
- Configuration for Resources to validated holistically rather than individually
  per-Resource
  - e.g. ensure the `Service.selector` and `Deployment.spec.template` labels
    match.
  - e.g. MutatingWebHooks are scoped to a single Resource instance at a time.
- Low-level tweaks to the output of high-level abstractions
  - e.g. add an `init container` to a client _"CRD"_ Resource after it was generated.
- Composition and layering of multiple functions together
  - Compose generation, injection, validation together

## Spec

### Input Type

A function MUST accept as input a single [Kubernetes List type][3].
The `items` field in the input will contain a sequence of [Object types][3].
A function MAY not support [Simple types][3] and List types.

An example using `v1/ConfigMapList` as input:

```yaml
apiVersion: v1
kind: ConfigMapList
items:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: config1
    data:
      p1: v1
      p2: v2
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: config2
```

An example using `v1/List` as input:

```yaml
apiVersion: v1
kind: List
items:
  spec:
   - apiVersion: foo-corp.com/v1
    kind: FulfillmentCenter
    metadata:
      name: staging
    address: "100 Main St."
  - apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: namespace-reader
    rules:
      - resources:
          - namespaces
        apiGroups:
          - ""
        verbs:
          - get
          - watch
          - list
```

In addition, a function MUST accept as input a List of kind `ResourceList` where the
`functionConfig` field, if present, will contain the invocation-specific configuration passed to the function
by the orchestrator.
Functions MAY consider this field optional so that they can be triggered in an ad-hoc fashion.

An example using `config.kubernetes.io/v1beta1/ResourceList` as input:

```yaml
apiVersion: config.kubernetes.io/v1beta1
kind: ResourceList
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    metadata:
      annotations:
        config.k8s.io/function: |
          container:
            image: gcr.io/example/foo:v1.0.0
  spec:
    address: "100 Main St."
items:
  - apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: namespace-reader
    rules:
      - resources:
          - namespaces
        apiGroups:
          - ""
        verbs:
          - get
          - watch
          - list
```

Here `FulfillmentCenter` kind with name `staging` is passed as the invocation-specific configuration
to the function.

### Output Type

A function’s output MUST be the same as the input specification above
-- i.e. `ResourceList` or `List`.
This is necessary to enable chaining two or more functions together in a pipeline.
The serialization format of the output SHOULD match that of its input on each invocation
-- e.g. if the input was a `ResourceList`, the output should also be a `ResourceList`.

### Serialization Format

A function MUST support YAML as a serialization format for the input and output.
A function MUST use utf8 encoding (as YAML is a superset of JSON, JSON will also be supported
by any conforming function).

### Operations

A function MAY Create, Update, or Delete any number of items in the `items` field and output the
resultant list.

A function MAY modify annotations with prefix `config.kubernetes.io`, but must be careful about
doing so since they’re used for orchestration purposes and will likely impact subsequent functions
in the pipeline.

A function SHOULD preserve comments when input serialization format is YAML.
This allows for human authoring of configuration to coexist with changes made by functions.

### Containerization

A function MUST be implemented as a container.

A function container MUST be capable of running as a non-root user if it does not require
access to host filesystem or makes network calls.

### stdin/stdout/stderr and Exit Codes

A function MUST accept input from stdin and emit output to stdout.

Any error messages MUST be emitted to stderr.

An exit code of zero indicates function execution was successful.
A non-zero exit code indicates a failure.

[1]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[2]: https://tools.ietf.org/html/rfc2119
[3]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#types-kinds
