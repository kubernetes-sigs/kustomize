# Kustomization File Fields

An explanation of the fields in a [kustomization.yaml](glossary.md#kustomization) file.


## Resources

What existing things should be customized.

| Field  | Type  | Explanation |
|---|---|---|
|[resources](#resources) |  list  |Files containing k8s API objects, or directories containing other kustomizations. |
|[CRDs](#crds)| list |Custom resource definition files, to allow specification of the custom resources in the resources list. |

## Generators

What things should be created (and optionally subsequently customized)?

| Field  | Type  | Explanation |
|---|---|---|
|[configMapGenerator](#configmapgenerator)| list  |Each entry in this list results in the creation of one ConfigMap resource (it's a generator of n maps).|
|[secretGenerator](#secretgenerator)| list  |Each entry in this list results in the creation of one Secret resource (it's a generator of n secrets)|
|[generatorOptions](#generatoroptions)|string|generatorOptions modify behavior of all ConfigMap and Secret generators|
|[generators](#generators)|list|[plugin](plugins) configuration files|


## Transformers

What transformations (customizations) should be applied?

| Field  | Type  | Explanation |
|---|---|---|
| [commonLabels](#commonlabels) | string | Adds labels and some corresponding label selectors to all resources. |
| [commonAnnotations](#commonannotations) | string | Adds annotions (non-identifying metadata) to add all resources. |
| [images](#images) | list | Images modify the name, tags and/or digest for images without creating patches. |
| [inventory](#inventory) | struct | Specify an object who's annotations will contain a build result summary. |
| [namespace](#namespace)   | string | Adds namespace to all resources |
| [namePrefix](#nameprefix) | string | Prepends value to the names of all resources |
| [nameSuffix](#namesuffix) | string | The value is appended to the names of all resources. |
| [replicas](#replicas) | list | Replicas modifies the number of replicas of a resource. |
| [patches](#patches) | list | Each entry should resolve to a patch that can be applied to multiple targets. |
|[patchesStrategicMerge](#patchesstrategicmerge)| list |Each entry in this list should resolve to a partial or complete resource definition file.|
|[patchesJson6902](#patchesjson6902)| list  |Each entry in this list should resolve to a kubernetes object and a JSON patch that will be applied to the object.|
|[transformers](#transformers)|list|[plugin](plugins) configuration files|


## Meta

[k8s metadata]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

|Field|Type|Explanation|
|---|---|---|
| [vars](#vars)     | string | Vars capture text from one resource's field and insert that text elsewhere. |
| [apiVersion](#apiversion)     | string | [k8s metadata] field. |
| [kind](#kind)     | string | [k8s metadata] field. |

----

### apiVersion

If missing, this field's value defaults to
```
apiVersion: kustomize.config.k8s.io/v1beta1
```

### bases

The `bases` field was deprecated in v2.1.0.

Move entries into the [resources](#resources)
field.  This allows bases - which are still a
[central concept](glossary.md#base) - to be
ordered relative to other input resources.

### commonLabels

Adds labels to all resources and selectors
```
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

### commonAnnotations

Adds annotions (non-identifying metadata) to add
all resources. Like labels, these are key value
pairs.

```
commonAnnotations:
  oncallPager: 800-555-1212
```

### configMapGenerator

Each entry in this list results in the creation of
one ConfigMap resource (it's a generator of n maps).

The example below creates two ConfigMaps. One with the
names and contents of the given files, the other with
key/value as data.

Each configMapGenerator item accepts a parameter of
`behavior: [create|replace|merge]`.
This allows an overlay to modify or
replace an existing configMap from the parent.

```
configMapGenerator:
- name: myJavaServerProps
  files:
  - application.properties
  - more.properties
- name: myJavaServerEnvVars
  literals:	
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
```

### crds

Each entry in this list should be a relative path to
a file for custom resource definition (CRD).

The presence of this field is to allow kustomize be
aware of CRDs and apply proper
transformation for any objects in those types.

Typical use case: A CRD object refers to a
ConfigMap object.  In a kustomization, the ConfigMap
object name may change by adding namePrefix,
nameSuffix, or hashing. The name reference for this
ConfigMap object in CRD object need to be updated
with namePrefix, nameSuffix, or hashing in the
same way.

The annotations can be put into openAPI definitions are:
 -  "x-kubernetes-annotation": ""
 -  "x-kubernetes-label-selector": ""
 -  "x-kubernetes-identity": ""
 -  "x-kubernetes-object-ref-api-version": "v1",
 -  "x-kubernetes-object-ref-kind": "Secret",
 -  "x-kubernetes-object-ref-name-key": "name",


```

crds:
- crds/typeA.yaml
- crds/typeB.yaml
```


### generatorOptions

Modifies behavior of all [ConfigMap](#configmapgenerator)
and [Secret](#secretgenerator) generators.

```
generatorOptions:
  # labels to add to all generated resources
  labels:
    kustomize.generated.resources: somevalue
  # annotations to add to all generated resources
  annotations:
    kustomize.generated.resource: somevalue
  # disableNameSuffixHash is true disables the default behavior of adding a
  # suffix to the names of generated resources that is a hash of
  # the resource contents.
  disableNameSuffixHash: true
```

### generators

A list of generator [plugin](plugins) configuration files.

```
generators:
- mySecretGeneratorPlugin.yaml
- myAppGeneratorPlugin.yaml
```

### images

Images modify the name, tags and/or digest for images without creating patches.
E.g. Given this kubernetes Deployment fragment:

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

### inventory

See [inventory object](inventory_object.md).

### kind

If missing, this field's value defaults to

```
kind: Kustomization
```


### namespace

Adds namespace to all resources

```
namespace: my-namespace
```

### namePrefix

Prepends value to the names of all resources
Ex. a deployment named `wordpress` would become `alices-wordpress`

```
namePrefix: alices-
```

### nameSuffix

The value is appended to the names of all
resources.  Ex. A deployment named `wordpress`
would become `wordpress-v2`.

The suffix is appended before content has if
resource type is ConfigMap or Secret.

```
nameSuffix: -v2
```

### patches

Each entry in this list should resolve to an Patch object,
which includes a patch and a target selector. 
The patch can be either a strategic merge patch or a JSON patch.
it can be either a patch file or an inline string.
The target selects
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

### patchesStrategicMerge

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

### patchesJson6902

Each entry in this list should resolve to
a kubernetes object and a JSON patch that will be applied
to the object.
The JSON patch is documented at https://tools.ietf.org/html/rfc6902

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

### replicas

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


#### Limitation

As this declaration does not take in a `kind:` nor a `group:`
it will match any `group` and `kind` that has a matching name and
that is one of:
- `Deployment`
- `ReplicationController`
- `ReplicaSet`
- `StatefulSet`

For more complex use cases, revert to using a patch.


### resources

Each entry in this list must be a path to a
_file_, or a path (or URL) refering to another
kustomization _directory_, e.g.

```
resource:
- myNamespace.yaml
- sub-dir/some-deployment.yaml
- ../../commonbase
- github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
- deployment.yaml
- github.com/kubernets-sigs/kustomize//examples/helloWorld?ref=test-branch
```

Resources will be read and processed in
depth-first order.

Files should contain k8s resources in YAML form.
A file may contain multiple resources separated by
the document marker `---`.  File paths should be
specified _relative_ to the directory holding the
kustomization file containing the `resources`
field.

[hashicorp URL]: https://github.com/hashicorp/go-getter#url-format

Directory specification can be relative, absolute,
or part of a URL.  URL specifications should
follow the [hashicorp URL] format.  The directory
must contain a `kustomization.yaml` file.


### secretGenerator

Each entry in this list results in the creation of
one Secret resource (it's a generator of n secrets).

```
secretGenerator:
- name: app-tls
  files:
  - secret/tls.cert
  - secret/tls.key
  type: "kubernetes.io/tls"
- name: app-tls-namespaced
  # you can define a namespace to generate secret in, defaults to: "default"
  namespace: apps
  files:
  - tls.crt=catsecret/tls.cert
  - tls.key=secret/tls.key
  type: "kubernetes.io/tls"
- name: env_file_secret
  envs:
  - env.txt
  type: Opaque
```

### vars

Vars are used to capture text from one resource's field
and insert that text elsewhere - a reflection feature.

For example, suppose one specifies the name of a k8s Service
object in a container's command line, and the name of a
k8s Secret object in a container's environment variable,
so that the following would work:

```
containers:
  - image: myimage
    command: ["start", "--host", "$(MY_SERVICE_NAME)"]
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
```

To do so, add an entry to `vars:` as follows:

```
vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
- name: MY_SERVICE_NAME
  objref:
    kind: Service
    name: my-service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: ANOTHER_DEPLOYMENTS_POD_RESTART_POLICY
  objref:
    kind: Deployment
    name: my-deployment
    apiVersion: apps/v1
  fieldref:
    fieldpath: spec.template.spec.restartPolicy
```

A var is a tuple of variable name, object
reference and field reference within that object.
That's where the text is found.

The field reference is optional; it defaults to
`metadata.name`, a normal default, since kustomize
is used to generate or modify the names of
resources.

At time of writing, only string type fields are
supported.  No ints, bools, arrays etc.  It's not
possible to, say, extract the name of the image in
container number 2 of some pod template.

A variable reference, i.e. the string '$(FOO)',
can only be placed in particular fields of
particular objects as specified by kustomize's
configuration data.

The default config data for vars is at
https://github.com/kubernetes-sigs/kustomize/blob/master/pkg/transformers/config/defaultconfig/varreference.go
Long story short, the default targets are all
container command args and env value fields.

Vars should _not_ be used for inserting names in
places where kustomize is already handling that
job.  E.g., a Deployment may reference a ConfigMap
by name, and if kustomize changes the name of a
ConfigMap, it knows to change the name reference
in the Deployment.


