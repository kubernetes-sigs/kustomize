# KRM Function Spec

**Authors**: @mengqiy

**Reviewers**:
- @KnVerey
- @natasha41575
- @yuwenma

**Status**: implementable

## Summary

[KRM function spec v1] only supports one-time evaluation. Various tools have
been developed around KRM functions. But the performance limitation has been a
problem. This proposal is intended to solve the performance problem with spec v2.

[KRM function spec v1]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md

## Motivation

Currently, KRM function evaluation is a one-time operation and functions are
short-lived. According to KRM function spec v1, a function has the following
steps:

1. Spin up the runtime (exec or container)
2. (Optional) Preprocessing (e.g. fetching something from the network or parsing OpenAPI schema)
3. Read a resource list from stdin
4. Mutate or validate the resource list
5. Write the resource list to stdout
6. Shut down with exit code

Noteworthy that spinning a container takes seconds (step 1). Fetching an OpenAPI
schema from network may take 0.5 second (step 2). Parsing OpenAPI schema may
take 0.5 second (step 2).
Performance can be improved if we can reuse step 1, 2 and 6 for functions.

**Goals:**

1. Design KRM function spec v2 that enables the performance improvement when
   evaluating same functions. 
2. Spec v2 should be compatible with spec v1 as much as possible.

**Non-goals:**

1. Develop low-level container runtime (e.g. runc and crun alternatives) to improve performance.
1. Develop container engine (e.g. docker and podman alternatives) to improve performance.

## Proposal

Spec v2 is a superset of spec v1. Spec v2 supports running a KRM function as a
server.

There are no changes to the resource list schema.

### Function Metadata Schema

We have defined the function metadata schema [here](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/2906-kustomize-function-catalog/README.md#function-metadata-schema).
We will introduce a new field called `conformWithSpecVersions` indicating which
FRM function spec version(s) the function implements.

```yaml
group: example.com
names:
  kind: SetNamespace
description: "A short description of the KRM function"
publisher: example.com
versions:
  - name: v1
    schema:
      openAPIV3Schema: ... # inline schema like CRD
    # conformWithSpecVersions is a list KRM function spec versions that the function conforms.
    conformWithSpecVersions: # New field
      - v2
      - v1
    idempotent: true|false
    runtime:
      container:
        image: docker.example.co/functions/set-namespace:v1.2.3
        sha256: a428de... # The digest of the image which can be verified against. This field is required if the version is semver.
        requireNetwork: true|false
        requireStorageMount: true|false
    usage: <a URL pointing to a README.md>
    examples:
      - <a URL pointing to a README.md>
      - <another URL pointing to another README.md>
    license: Apache 2.0
  - name: v1beta1
    ...
```

For KRM functions that use container as runtime, we should leverage OCI
annotations or something similar to enable distributing KRM function metadata
with the image.

If a container image doesn't have this metadata with the image, we will treat it
as using v1 spec.

### HTTP Server

KRM function can be run as an HTTP server.

It must support the following configuration. The flag if specified take
precedence over the environment variable.

| Description       | Flag        | Environment Variable      |
|-------------------|-------------|---------------------------|
| Address to listen | --http-addr | KRM_FUNCTION_HTTP_ADDRESS |

It only accepts POST requests.

#### Request

Body: The input `resourceList`.

Headers:

| Header       | Valid values                           | Description                                                                                 |
|--------------|----------------------------------------|---------------------------------------------------------------------------------------------|
| Content-Type | `text/yaml`, `application/json`        | Must be one of the valid values.                                                            |
| Accept       | `text/yaml`, `application/json`, `*/*` | `Accept` can be set to be one of the valid values.<br/>If `*/*` or omitted, yaml will be served. |

#### Response

Body: The output `resourceList`.

| Header       | Valid values                     | Description                       |
|--------------|----------------------------------|-----------------------------------|
| Content-Type | `text/yaml`, `application/json`  | Must be one of the valid values.  |

### GRPC Server

KRM function can be run as a GRPC server.

It must support the following configuration. The flag if specified take
precedence over the environment variable.

| Description       | Flag        | Environment Variable      |
|-------------------|-------------|---------------------------|
| Address to listen | --grpc-addr | KRM_FUNCTION_GRPC_ADDRESS |

<details>

<summary>
protobuf definition
</summary>

TODO: add protobuf here

</details>



### User Stories

#### Story 1

As an end user of KRM functions, I can reference a function's schema to find out
if it implements spec v2. If so, I can configure it to run as either an HTTP
server or a GRPC server.

#### Story 2

As a KRM function orchestration tool developer, I can develop tools that work
with both spec v1 functions and spec v2 functions. My tool can choose tp reuse
KRM function runtime when the function conforms spec v2.

#### Story 3

As a KRM function developer, I implement my function to conform spec v2 and I
publish my function in the [KRM function registry]. If I'm developing a
container-based function, I also save the function metadata in an OCI annotation.

[KRM function registry]: https://github.com/kubernetes-sigs/krm-functions-registry

### Scalability

With this change, we will be able to evaluate more KRM functions given the same
time duration and same amount of resources (CPU and memory).

## Drawbacks

Supporting server-mode KRM functions won't help everyone. It will help only when
there are a great amount of KRM function evaluation demand and same function are
evaluated over and over again.

## Alternatives

### No spec v1 in spec v2

Instead of making spec v2 a superset of spec v1, we can make spec v2 to support
server mode only.

### Multiple Resource Lists in stdin and stdout

A KRM function keep reading from stdin. It starts to process the input resource
list after seeing a yaml separator i.e. `---`. The function will write the
resource list to stdout with a yaml separator at the end after finishing
processing one resource list. It will wait until reading the next the resource
list from stdin.

## Rollout Plan

### Alpha

The new spec will be using version v2alpha1 initially.

### Beta

The new spec will be using version v2beta1 when it becomes more mature.

### GA

The new spec will graduate as v2 when it's mature.
