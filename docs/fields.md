# Kustomization File Fields

[field-name-namespace]: plugins/builtins.md#field-name-namespace
[field-name-images]: plugins/builtins.md#field-name-images
[field-name-namePrefix]: plugins/builtins.md#field-name-prefix
[field-name-nameSuffix]: plugins/builtins.md#field-name-prefix
[field-name-patches]: plugins/builtins.md#field-name-patches
[field-name-patchesStrategicMerge]: plugins/builtins.md#field-name-patchesStrategicMerge
[field-name-patchesJson6902]: plugins/builtins.md#field-name-patchesJson6902
[field-name-replicas]: plugins/builtins.md#field-name-replicas
[field-name-secretGenerator]: plugins/builtins.md#field-name-secretGenerator
[field-name-commonLabels]: plugins/builtins.md#field-name-commonLabels
[field-name-commonAnnotations]: plugins/builtins.md#field-name-commonAnnotations
[field-name-configMapGenerator]: plugins/builtins.md#field-name-configMapGenerator


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

_The `bases` field was deprecated in v2.1.0._

Move entries into the [resources](#resources)
field.  This allows bases - which are still a
[central concept](glossary.md#base) - to be
ordered relative to other input resources.

### commonLabels
See [field-name-commonLabels].

### commonAnnotations
See [field-name-commonAnnotations].

### configMapGenerator
See [field-name-configMapGenerator].

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

See [field-name-images].

### inventory

See [inventory object](inventory_object.md).

### kind

If missing, this field's value defaults to

```
kind: Kustomization
```

### namespace

See [field-name-namespace].

### namePrefix

See [field-name-namePrefix].

### nameSuffix

See [field-name-nameSuffix].

### patches

See [field-name-patches].

### patchesStrategicMerge

See [field-name-patchesStrategicMerge].

### patchesJson6902

See [field-name-patchesJson6902].

### replicas

See [field-name-replicas].

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

See [field-name-secretGenerator].

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


