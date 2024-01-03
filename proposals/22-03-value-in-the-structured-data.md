<!--
**Note:** When your proposal is complete, all of these comment blocks should be removed.

To get started with this template:

- [ ] **Make a copy of this file.**
  Name it `YY-MM-short-descriptive-title.md` (where `YY-MM` is the current year and month).
- [ ] **Fill out this file as best you can.**
  At minimum, you should fill in the "Summary" and "Motivation" sections.
- [ ] **Create a PR.**
  Ping `@kubernetes-sigs/kustomize-admins` and `@kubernetes-sigs/kustomize-maintainers`.
-->

# Replacements and Patch value in the structured data.

**Authors**:
- koba1t

**Reviewers**:
- natasha41575
- knverey

**Status**: implementable
<!--
In general, all proposals made should be merged for the record, whether or not they are accepted.
Use the status field to record the results of the latest review:
- implementable: The default for this repo. If the proposal is merged, you can start working on it.
- deferred: The proposal may be accepted in the future, but it has been shelved for the time being.
A new PR must be opened to update the proposal and gain reviewer consensus before work can begin.
- withdrawn: The author changed their mind and no longer wants to pursue the proposal.
A new PR must be opened to update the proposal and gain reviewer consensus before work can begin.
- rejected: This proposal should not be implemented.
- replaced: If you submit a new proposal that supersedes an older one,
update the older one's status to "replaced by <link>".
-->

## Summary

This proposal decides the interfaces to change values in the structured data (like json,yaml) inside a Kubernetes objects' field value and implements changing function target a few formats (mainly json).


## Motivation

<!--
If this proposal is an expansion of an existing GitHub issue, link to it here.
-->

kustomize can apply structured edits to Kubernetes objects defined in yaml files.\
Sometimes structured multi-line or long single line string (ex. json,yaml, and other structured format data) is injected in Kubernetes objects' string literal field. From the kustomize perspective, these "structured" multiline strings are just arbitrary unstructured strings.\
So, kustomize can't manipulate one value on structured, formatted data in the Kubernetes object's string literal field. This function is expected behavior, but kustomize will be very helpful if it can change the value of structured data like json and yaml substrings in a string literal.\
This proposal allows the user to identify strings literals containing JSON/YAML data so that Kustomize can make structured edits to the data they contain. This allows the requested functionality to be added without violating Kustomize's core principle of supporting structured edits exclusively.\

For example, kustomize can't change the value `"REPLACE_TARGET_HOSTNAME"` in this yaml file straightforwardly.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: target-configmap
data:
  config.json: |-
    {"config": {
      "id": "42",
      "hostname": "REPLACE_TARGET_HOSTNAME"
    }}
```

This function has been wanted for a long time.
- https://github.com/kubernetes-sigs/kustomize/issues/680
- https://github.com/kubernetes-sigs/kustomize/issues/3787
- https://github.com/kubernetes-sigs/kustomize/issues/4517

This function can be an alternative deprecated [vars](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/vars/) function in many use cases.

**Goals:**
<!--
List the specific goals of the proposal. What is it trying to achieve? How will we
know that this has succeeded?
-->

1. Provide the way to update values in the structured data like kubernetes objects with JSON/YAML formats.


**Non-goals:**
<!--
What is out of scope for this proposal? Listing non-goals helps to focus discussion
and make progress.
-->

1. Do not provide [Unstructured edits, because kustomize eschews parameterization](https://kubectl.docs.kubernetes.io/faq/kustomize/eschewedfeatures/#unstructured-edits).

1. Do not provide a way for targeting/merging values within strings from the [patches](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patches/) field.

1. Do not provide a way for customizing field merge strategies for the embedded data.

## Proposal

<!--
This is where we get down to the specifics of what the proposal actually is.
Include enough information to illustrate your proposal, but try not to
overwhelm reviewers with details. Focus on APIs and interfaces rather than implementation details,
e.g.:
- Does this proposal require new kinds, fields or CLI flags?
- Will this feature require extending the public interface of Kustomize's Go packages?
(it's ok if you're not sure yet)

A proof of concept PR is NOT required but is preferable to including large amounts of code
inline here, if you feel such implementation details are required to adequately explain your design.
If you have a PR, link to it at the top of this section.
-->

### Replacement the value in structured data

I propose to extend the `fieldPath` and `fieldPaths` fields to replace the value in structured data with the replacements function.\
This idea is extending how to select any field in replacement [TargetSelector](https://github.com/kubernetes-sigs/kustomize/blob/8668691ade05bc17b3c6f44bcd4723735033196e/api/types/replacement.go#L52-L64). If the `source.fieldPath` and `targets.fieldPaths` had extra values after a specific string literal in Yaml, Kustomize tries to parse that string as structure data and tries to drill down using that additional values.

#### Example.

```yaml
## replacement
replacements:
- source:
    kind: ConfigMap
    name: source-configmap
    fieldPath: data.HOSTNAME
  targets:
  - select:
      kind: ConfigMap
      name: target-configmap
    fieldPaths:
    - data.config\.json.config.hostname # A path after `config\.json` is pointing one place in the structured data.
```

Please check [Story 1](#Story-1).

### Disciplined merge the value in structured data with configMapGenerator and secretGenerator

This Proposal is to add parameters for [configMapGenerator](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/configmapgenerator/) and [secretGenerator](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/secretgenerator/) to allow the merging of two string literals when the behavior option is set to merge and string literals value is structured.\
This idea is to add one parameter, `mergeValues` to [GeneratorArgs](https://github.com/kubernetes-sigs/kustomize/blob/672c751715be7dd0b43b4a2fce956c84452e0db9/api/types/generatorargs.go#L7-L27) used from configMapGenerator and secretGenerator.
The `mergeValues` option is a list that contains two parameters, `key` and `format`. One element corresponds to one string literal, `key` is used to select a string literal to merge, and `format` is used to designate a format for string literal which is JSON or YAML.\
This merge operation will be implemented for a part of [Overriding Base ConfigMap Values](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/configmapgenerator/#overriding-base-configmap-values). It will execute to merge two string literal having same [key](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#define-the-key-to-use-when-creating-a-configmap-from-a-file) name when merging two configMap or secret.\

#### Example

```yaml
configMapGenerator:
- name: demo-settings
  behavior: merge                     # This function requires `behavior: merge`.
  mergeValues:
  - key: config.json # Key with a target to merge.
    format: json # Setting structured data format MUST be YAML/JSON.
  literals:
  - config.json: |-
      {
        "config": {
          "hostname": "REPLACE_TARGET_HOSTNAME",
          "value": {
            "foo": "bar"
          }
        }
      }
```

Please check [Story 2](#Story-2).

#### Appendix
- https://github.com/kubernetes-sigs/kustomize/issues/680#issuecomment-458834785

## User Stories
<!--
Describe what people will be able to do if this KEP is implemented. If different user personas
will use the feature differently, consider writing separate stories for each.
Include as much detail as possible so that people can understand the "how" of the system.
The goal here is to make this feel real for users without getting bogged down.
-->

#### Story 1

Scenario summary: Replacement the value inside for structured data(json) in the configMap.
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

kustomize patching overlay is very strong to manage common yaml when using many cluster.\
But, if you want to set cluster specific change value in the json with configMap data field, you have to replacement whole json file.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: target-configmap-dev
data:
  config.json: |-
    {"config": {
      "id": "42",
      "hostname": "dev.example.com
    }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: target-configmap-prod
data:
  config.json: |-
    {"config": {
      "id": "42",
      "hostname": "prod.example.com"
    }}
```

So if we can replacement this value in the substring formatted with json, we can easy to overlay this difference.

```yaml
## source
apiVersion: v1
kind: ConfigMap
metadata:
  name: source-configmap
data:
  HOSTNAME: www.example.com
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: target-configmap
data:
  config.json: |-
    {"config": {
      "id": "42",
      "hostname": "REPLACE_TARGET_HOSTNAME"
    }}
```

```yaml
## replacement
replacements:
- source:
    kind: ConfigMap
    name: source-configmap
    fieldPath: data.HOSTNAME
  targets:
  - select:
      kind: ConfigMap
      name: target-configmap
    fieldPaths:
    - data.config\.json.config.hostname
```

```yaml
## expected
apiVersion: v1
kind: ConfigMap
metadata:
  name: source-configmap
data:
  HOSTNAME: www.example.com
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: target-configmap
data:
  config.json: '{"config":{"hostname":"www.example.com","id":"42"}}'
```

#### Story 2

Scenario summary: Merge the two values formatted by structured data(json) with configMapGenerator.

<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

Many application needs to be set with json format file. And, That file is handled with configMap when the application is running on kubernetes.\
So, If kustomize configMapGenerator can merge for json inside a configMap data value, A json format file will be simple and easy to handle.

#### Source
```yaml
# base/kustomization.yaml
configMapGenerator:
- name: demo
  literals:
  - config.json: |-
      {
        "config": {
          "loglevel": debug,
          "parameter": {
            "foo": "bar"
          }
        }
      }
```

```yaml
# overlay/kustomization.yaml
resources:
- ../base
configMapGenerator:
- name: demo
  behavior: merge
  mergeValues:
  - key: config.json # Key with a target to merge.
    format: json # Setting structured data format MUST be YAML/JSON.
  literals:
  - config.json: |-
      {
        "config": {
          "hostname": "www.example.com",
          "parameter": {
            "baz": "qux"
          }
        }
      }
```

#### Result
```yaml
apiVersion: v1
data:
  config.json: |-
    {
      "config": {
        "loglevel": debug,
        "hostname": "www.example.com",
        "parameter": {
          "foo": "bar",
          "baz": "qux"
        }
      }
    }
kind: ConfigMap
metadata:
  name: demo-xxxxxxxxxx  # name suffix hash
```


#### Story 3

Scenario summary: Replacement the value inside for yaml in the configMap.
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

A few applications require to use yaml format config file, and some major cloudnative applications are using that.(ex, Prometheus,AlertManager)\
A value you want to overlay in yaml is usually into a nested yaml structure. So if you can overlay value inside yaml, you won't need to copy a whole yaml file.

```yaml
## source
apiVersion: v1
kind: ConfigMap
metadata:
  name: environment-config
data:
  env: dev
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |-
    global:
      external_labels:
        prometheus_env: TARGET_ENVIROMENT
    scrape_configs:
      - job_name: "prometheus"
        static_configs:
          - targets: ["localhost:9090"]
```

```yaml
## replacement
replacements:
- source:
    kind: ConfigMap
    name: environment-config
    fieldPath: data.env
  targets:
  - select:
      kind: ConfigMap
      name: prometheus-config
    fieldPaths:
    - data.prometheus\.yml.global.external_labels?prometheus_env
```

```yaml
## expected
apiVersion: v1
kind: ConfigMap
metadata:
  name: environment-config
data:
  env: dev
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |-
    global:
      external_labels:
        prometheus_env: dev
    scrape_configs:
      - job_name: "prometheus"
        static_configs:
          - targets: ["localhost:9090"]
```

#### Story 4

Scenario summary: Replacement the value inside for json in the annotations.
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

A few times, an application on your cluster requires to set json format config for the *Annotations* on kubernetes yaml resources. We need to overlay in this position value.

##### Example
```yaml
## source
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: debug-backend-config
spec:
  securityPolicy:
    name: "debug-security-policy"
---
apiVersion: v1
kind: Service
metadata:
  name: appA-svc
  annotations:
    cloud-provider/backend-config: '{"ports": {"appA":"gke-default-backend-config"}}'
spec:
  ports:
  - name: appA
    port: 1234
    protocol: TCP
    targetPort: 8080
```

```yaml
## replacement
replacements:
- source:
    kind: BackendConfig
    name: debug-backend-config
    fieldPath: metadata.name
  targets:
  - select:
      kind: ConServicefigMap
      name: appA-svc
    fieldPaths:
    - metadata.annotations.cloud-provider/backend-config.ports.appA
```

```yaml
## expected
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: debug-backend-config
spec:
  securityPolicy:
    name: "debug-security-policy"
---
apiVersion: v1
kind: Service
metadata:
  name: appA-svc
  annotations:
    cloud-provider/backend-config: '{"ports": {"appA":"debug-backend-config"}}'
spec:
  ports:
  - name: appA
    port: 1234
    protocol: TCP
    targetPort: 8080
```

- https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-features#unique_backendconfig_per_service_port
- https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#listen-ports


### Risks and Mitigations
<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security, end-user privacy, and how this will
impact the larger Kubernetes ecosystem.
-->

### Dependencies
<!--
Kustomize tightly controls its Go dependencies in order to remain approved for
integration into kubectl. It cannot depend directly on kubectl or apimachinery code.
Identify any new Go dependencies this proposal will require Kustomize to pull in.
If any of them are large, is there another option?
-->

### Scalability
<!--
Is this feature expected to have a performance impact?
Explain to what extent and under what conditions.
-->

## Drawbacks
<!--
Why should this proposal _not_ be implemented?
-->

## Alternatives
<!--
What other approaches did you consider, and why did you rule them out? Be concise,
but do include enough information to express the idea and why it was not acceptable.
-->

## Rollout Plan
<!--
Depending on the scope of the features and the risks enabling it implies,
you may need to use a formal graduation process. If you don't think this is
necessary, explain why here, and delete the alpha/beta/GA headings below.
-->

### Alpha
<!--
New Kinds should be introduced with an alpha group version.
New major features should often be gated by an alpha flag at first.
New transformers can be introduced for use in the generators/validators/transformers fields
before they get their own top-level field in Kustomization.
-->

- Will the feature be gated by an "alpha" flag? Which one?
- Will the feature be available in `kubectl kustomize` during alpha? Why or why not?

### Beta
<!--
If the alpha was not available in `kubectl kustomize`, you need a beta phase where it is.
Full parity with `kubectl kustomize` is required at this stage.
-->

### GA
<!--
You should generally wait at least two `kubectl` release cycles before promotion to GA,
to ensure that the broader user base has time to try the feature and provide feedback.
For example, if your feature first appears in kubectl 1.23, promote it in 1.25 or later.
-->
