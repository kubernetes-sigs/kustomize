---
title: "Builtin Plugins"
linkTitle: "Builtin Plugins"
type: docs
description: >
    Builtin Plugins
---


# Builtin Plugins

A list of kustomize's builtin plugins - both
generators and transformers.

For each plugin, an example is given for

* implicitly triggering
the plugin via a dedicated kustomization
file field (e.g. the `AnnotationsTransformer` is
triggered by the `commonAnnotations` field).

* explicitly triggering the plugin
via the `generators` or `transformers` field
(by providing a config file specifying the
plugin).

The former method is convenient but limited in
power as most of the plugins arguments must
be defaulted.  The latter method allows for
complete plugin argument specification.

[types.GeneratorOptions]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/generatoroptions.go
[types.SecretArgs]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/secretargs.go
[types.ConfigMapArgs]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/configmapargs.go
[config.FieldSpec]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/fieldspec.go
[types.ObjectMeta]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/objectmeta.go
[types.Selector]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/selector.go
[types.Replica]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/replica.go
[types.PatchStrategicMerge]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/patchstrategicmerge.go
[types.PatchTarget]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/patchtarget.go
[image.Image]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/image.go

## _AnnotationTransformer_

### Usage via `kustomization.yaml`

#### field name: `commonAnnotations`

Adds annotions (non-identifying metadata) to add
all resources. Like labels, these are key value
pairs.

```
commonAnnotations:
  oncallPager: 800-555-1212
```

### Usage via plugin

#### Arguments

> Annotations map\[string\]string
>
> FieldSpecs  \[\][config.FieldSpec]

#### Example
>
> ```
> apiVersion: builtin
> kind: AnnotationsTransformer
> metadata:
>   name: not-important-to-example
> annotations:
>   app: myApp
>   greeting/morning: a string with blanks
> fieldSpecs:
> - path: metadata/annotations
>   create: true
> ```

## _ConfigMapGenerator_

### Usage via `kustomization.yaml`

#### field name: `configMapGenerator`

Each entry in this list results in the creation of
one ConfigMap resource (it's a generator of n maps).

The example below creates three ConfigMaps. One with the names and contents of
the given files, one with key/value as data, and a third which sets an
annotation and label via `options` for that single ConfigMap.

Each configMapGenerator item accepts a parameter of
`behavior: [create|replace|merge]`.
This allows an overlay to modify or
replace an existing configMap from the parent.

Also, each entry has an `options` field, that has the
same subfields as the kustomization file's `generatorOptions` field.
  
This `options` field allows one to add labels and/or
annotations to the generated instance, or to individually
disable the name suffix hash for that instance.
Labels and annotations added here will not be overwritten
by the global options associated with the kustomization
file `generatorOptions` field.  However, due to how
booleans behave, if the global `generatorOptions` field
specifies `disableNameSuffixHash: true`, this will
trump any attempt to locally override it.

```
# These labels are added to all configmaps and secrets.
generatorOptions:
  labels:
    fruit: apple

configMapGenerator:
- name: my-java-server-props
  behavior: merge
  files:
  - application.properties
  - more.properties
- name: my-java-server-env-vars
  literals: 
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
  options:
    disableNameSuffixHash: true
    labels:
      pet: dog
- name: dashboards
  files:
  - mydashboard.json
  options:
    annotations:
      dashboard: "1"
    labels:
      app.kubernetes.io/name: "app1"
```

It is also possible to
[define a key](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#define-the-key-to-use-when-creating-a-configmap-from-a-file)
to set a name different than the filename.

The example below creates a ConfigMap
with the name of file as `myFileName.ini`
while the _actual_ filename from which the
configmap is created is `whatever.ini`.

```
configMapGenerator:
- name: app-whatever
  files:
  - myFileName.ini=whatever.ini
```

### Usage via plugin

#### Arguments

> [types.ConfigMapArgs]

#### Example
>
> ```
> apiVersion: builtin
> kind: ConfigMapGenerator
> metadata:
>   name: mymap
> envs:
> - devops.env
> - uxteam.env
> literals:
> - FRUIT=apple
> - VEGETABLE=carrot
> ```

## _ImageTagTransformer_

### Usage via `kustomization.yaml`

#### field name: `images`

Images modify the name, tags and/or digest for images
without creating patches.  E.g. Given this
kubernetes Deployment fragment:

```
containers:
- name: mypostgresdb
  image: postgres:8
- name: nginxapp
  image: nginx:1.7.9
- name: myapp
  image: my-demo-app:latest
- name: alpine-app
  image: alpine:3.7
```

one can change the `image` in the following ways:

- `postgres:8` to `my-registry/my-postgres:v1`,
- nginx tag `1.7.9` to `1.8.0`,
- image name `my-demo-app` to `my-app`,
- alpine's tag `3.7` to a digest value

all with the following *kustomization*:

```
images:
- name: postgres
  newName: my-registry/my-postgres
  newTag: v1
- name: nginx
  newTag: 1.8.0
- name: my-demo-app
  newName: my-app
- name: alpine
  digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
```

### Usage via plugin

#### Arguments

> ImageTag   [image.Image]
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```
> apiVersion: builtin
> kind: ImageTagTransformer
> metadata:
>   name: not-important-to-example
> imageTag:
>   name: nginx
>   newTag: v2
> ```

## _LabelTransformer_

### Usage via `kustomization.yaml`

#### field name: `commonLabels`

Adds labels to all resources and selectors

```
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

### Usage via plugin

#### Arguments

> Labels  map\[string\]string
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```
> apiVersion: builtin
> kind: LabelTransformer
> metadata:
>   name: not-important-to-example
> labels:
>   app: myApp
>   env: production
> fieldSpecs:
> - path: metadata/labels
>   create: true
> ```

## _NamespaceTransformer_

### Usage via `kustomization.yaml`

#### field name: `namespace`

Adds namespace to all resources

```
namespace: my-namespace
```

### Usage via plugin

#### Arguments

> [types.ObjectMeta]
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```
> apiVersion: builtin
>  kind: NamespaceTransformer
>  metadata:
>    name: not-important-to-example
>    namespace: test
>  fieldSpecs:
>  - path: metadata/namespace
>    create: true
>  - path: subjects
>    kind: RoleBinding
>    group: rbac.authorization.k8s.io
>  - path: subjects
>    kind: ClusterRoleBinding
>    group: rbac.authorization.k8s.io
> ```

## _PatchesJson6902_

### Usage via `kustomization.yaml`

#### field name: `patchesJson6902`

Each entry in this list should resolve to
a kubernetes object and a JSON patch that will be applied
to the object.
The JSON patch is documented at <https://tools.ietf.org/html/rfc6902>

target field points to a kubernetes object within the same kustomization
by the object's group, version, kind, name and namespace.
path field is a relative file path of a JSON patch file.
The content in this patch file can be either in JSON format as

```
 [
   {"op": "add", "path": "/some/new/path", "value": "value"},
   {"op": "replace", "path": "/some/existing/path", "value": "new value"}
 ]
 ```

or in YAML format as

```
- op: add
  path: /some/new/path
  value: value
- op: replace
  path: /some/existing/path
  value: new value
```

```
patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  path: add_init_container.yaml
- target:
    version: v1
    kind: Service
    name: my-service
  path: add_service_annotation.yaml
```

The patch content can be an inline string as well:

```
patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  patch: |-
    - op: add
      path: /some/new/path
      value: value
    - op: replace
      path: /some/existing/path
      value: "new value"
```

### Usage via plugin

#### Arguments

> Target [types.PatchTarget]
>
> Path   string
>
> JsonOp string

#### Example
>
> ```
> apiVersion: builtin
> kind: PatchJson6902Transformer
> metadata:
>   name: not-important-to-example
> target:
>   group: apps
>   version: v1
>   kind: Deployment
>   name: my-deploy
> path: jsonpatch.json
> ```

## _PatchesStrategicMerge_

### Usage via `kustomization.yaml`

#### field name: `patchesStrategicMerge`

Each entry in this list should be either a relative
file path or an inline content
resolving to a partial or complete resource
definition.

The names in these (possibly partial) resource
files must match names already loaded via the
`resources` field.  These entries are used to
_patch_ (modify) the known resources.

Small patches that do one thing are best, e.g. modify
a memory request/limit, change an env var in a
ConfigMap, etc.  Small patches are easy to review and
easy to mix together in overlays.

```
patchesStrategicMerge:
- service_port_8888.yaml
- deployment_increase_replicas.yaml
- deployment_increase_memory.yaml
```

The patch content can be a inline string as well.

```
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: nignx:latest
```

Note that kustomize does not support more than one patch
for the same object that contain a _delete_ directive. To remove
several fields / slice elements from an object create a single
patch that performs all the needed deletions.

### Usage via plugin

#### Arguments

> Paths \[\][types.PatchStrategicMerge]
>
> Patches string

#### Example
>
> ```
> apiVersion: builtin
> kind: PatchStrategicMergeTransformer
> metadata:
>   name: not-important-to-example
> paths:
> - patch.yaml
> ```

## _PatchTransformer_

### Usage via `kustomization.yaml`

#### field name: `patches`

Each entry in this list should resolve to an Patch
object, which includes a patch and a target selector.
The patch can be either a strategic merge patch or a
JSON patch. it can be either a patch file or an inline
string. The target selects
resources by group, version, kind, name, namespace,
labelSelector and annotationSelector. A resource
which matches all the specified fields is selected
to apply the patch.

```
patches:
- path: patch.yaml
  target:
    group: apps
    version: v1
    kind: Deployment
    name: deploy.*
    labelSelector: "env=dev"
    annotationSelector: "zone=west"
- patch: |-
    - op: replace
      path: /some/existing/path
      value: new value
  target:
    kind: MyKind
    labelSelector: "env=dev"
```

The `name` and `namespace` fields of the patch target selector are
automatically anchored regular expressions. This means that the value `myapp`
is equivalent to `^myapp$`.

### Usage via plugin

#### Arguments

> Path string
>
> Patch string
>
> Target \*[types.Selector]

#### Example
>
> ```
> apiVersion: builtin
> kind: PatchTransformer
> metadata:
>   name: not-important-to-example
> patch: '[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:latest"}]'
> target:
>   name: .*Deploy
>   kind: Deployment
> ```

## _PrefixSuffixTransformer_

### Usage via `kustomization.yaml`

#### field names: `namePrefix`, `nameSuffix`

Prepends or postfixes the value to the names
of all resources.

E.g. a deployment named `wordpress` could
become `alices-wordpress` or  `wordpress-v2`
or `alices-wordpress-v2`.

```
namePrefix: alices-
nameSuffix: -v2
```

The suffix is appended before the content hash if
the resource type is ConfigMap or Secret.

### Usage via plugin

#### Arguments

> Prefix     string
>
> Suffix     string
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```
> apiVersion: builtin
> kind: PrefixSuffixTransformer
> metadata:
>   name: not-important-to-example
> prefix: baked-
> suffix: -pie
> fieldSpecs:
>   - path: metadata/name
> ```

## _ReplicaCountTransformer_

### Usage via `kustomization.yaml`

#### field name: `replicas`

Replicas modified the number of replicas for a resource.

E.g. Given this kubernetes Deployment fragment:

```
kind: Deployment
metadata:
  name: deployment-name
spec:
  replicas: 3
```

one can change the number of replicas to 5
by adding the following to your kustomization:

```
replicas:
- name: deployment-name
  count: 5
```

This field accepts a list, so many resources can
be modified at the same time.

As this declaration does not take in a `kind:` nor a `group:`
it will match any `group` and `kind` that has a matching name and
that is one of:

- `Deployment`
- `ReplicationController`
- `ReplicaSet`
- `StatefulSet`

For more complex use cases, revert to using a patch.

### Usage via plugin

#### Arguments

> Replica [types.Replica]
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```
> apiVersion: builtin
> kind: ReplicaCountTransformer
> metadata:
>   name: not-important-to-example
> replica:
>   name: myapp
>   count: 23
> fieldSpecs:
> - path: spec/replicas
>   create: true
>   kind: Deployment
> - path: spec/replicas
>   create: true
>   kind: ReplicationController
> ```

## _SecretGenerator_

### Usage via `kustomization.yaml`

#### field name: `secretGenerator`

Each entry in the argument list
results in the creation of
one Secret resource
(it's a generator of n secrets).

This works like the `configMapGenerator` field
described above.

```
secretGenerator:
- name: app-tls
  files:
  - secret/tls.cert
  - secret/tls.key
  type: "kubernetes.io/tls"
- name: app-tls-namespaced
  # you can define a namespace to generate
  # a secret in, defaults to: "default"
  namespace: apps
  files:
  - tls.crt=catsecret/tls.cert
  - tls.key=secret/tls.key
  type: "kubernetes.io/tls"
- name: env_file_secret
  envs:
  - env.txt
  type: Opaque
- name: secret-with-annotation
  files:
  - app-config.yaml
  type: Opaque
  options:
    annotations:
      app_config: "true"
    labels:
      app.kubernetes.io/name: "app2"
```

### Usage via plugin

#### Arguments

> [types.ObjectMeta]
>
> [types.SecretArgs]

#### Example

> ```
> apiVersion: builtin
> kind: SecretGenerator
> metadata:
>   name: my-secret
>   namespace: whatever
> behavior: merge
> envs:
> - a.env
> - b.env
> files:
> - obscure=longsecret.txt
> literals:
> - FRUIT=apple
> - VEGETABLE=carrot
> ```
