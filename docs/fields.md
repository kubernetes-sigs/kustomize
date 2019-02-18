# Kustomize Fields

- [Operators](#operators)
- [Operands](#operands)
- [Generators](#generators)

You can find examples of how to use Kustomize [here](https://github.com/kubernetes-sigs/kustomize/tree/master/examples).

## Operators

For modifying operands, e.g. namePrefix, nameSuffix, commonLabels, patches, etc.

### Namespace

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

The value is appended to the names of all resources.
Ex. A deplou,ent names "wordpress" would become "wordpress-v2"
The suffix is appended before content has if resource type is ConfigMap or Secret
```
nameSuffix: -v2
```

### commonLabels

Adds labels to all resources and selectors
```
commonLabels:
    someName: someValue
    owner: alice
    app: bingo
```

### commonAnnotations

Adds annotions (non-identifying metadata) to add all resources. Like labls, these are key value pairs.

```
commonAnnotations:
    oncallPager: 800-555-1212
```

### vars

Vars are used to capture text from one resource's field
and insert that text elsewhere.

For example, suppose one specify the name of a k8s Service
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

### images

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

## Operands

[resources](#resources) - completely specified k8s API objects, e.g. deployment.yaml, configmap.yaml, etc.

[bases](#bases) - paths or github URLs specifying directories containing a kustomization. These bases may be subjected to more customization, or merely included in the output.

[CRDs](#crds) - custom resource definition files, to allow use of custom resources in the resources list. Not an actual operand - but allows the use of new operands.

### resources

Each entry in this list must resolve to an existing
resource definition in YAML.  These are the resource
files that kustomize reads, modifies and emits as a
YAML string, with resources separated by document
markers ("---").

```
resource:
- some-service.yaml
- sub-dir/some-deployment.yaml
```

### bases

Each entry in this list should resolve to a directory
containing a kustomization file, else the
customization fails.

The entry could be a relative path pointing to a local directory
or a url pointing to a directory in a remote repo.
The url should follow hashicorp/go-getter URL format
https://github.com/hashicorp/go-getter#url-format

The presence of this field means this file (the file
you a reading) is an _overlay_ that further
customizes information coming from these _bases_.

Typical use case: a dev, staging and production
environment that are mostly identical but differing
crucial ways (image tags, a few server arguments,
etc. that differ from the common base).
```
bases:
- ../../base
- github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
- github.com/Liujingfang1/mysql
- github.com/Liujingfang1/kustomize//examples/helloWorld?ref=test-branch
```

### crds

Each entry in this list should be a relative path to
a file for custom resource definition(CRD).

The presence of this field is to allow kustomize be
aware of CRDs and apply proper
transformation for any objects in those types.

Typical use case: A CRD object refers to a ConfigMap object.
In kustomization, the ConfigMap object name may change by adding namePrefix, nameSuffix, or hashing
The name reference for this ConfigMap object in CRD object need to be
updated with namePrefix, nameSuffix, or hashing in the same way.

```
crds:
- crds/typeA.yaml
- crds/typeB.yaml
```

### patchesStrategicMerge

Each entry in this list should resolve to
a partial or complete resource definition file.

The names in these (possibly partial) resource files
must match names already loaded via the `resources`
field or via `resources` loaded transitively via the
`bases` entries.  These entries are used to _patch_
(modify) the known resources.

Small patches that do one thing are best, e.g. modify
a memory request/limit, change an env var in a
ConfigMap, etc.  Small patches are easy to review and
easy to mix together in overlays.

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

- op: add
  path: /some/new/path
  value: value
- op:replace
  path: /some/existing/path
  value: new value

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

## Generators

Generators, for creating more resources (configmaps and secrets) which can then be customized.

### configMapGenerator

Each entry in this list results in the creation of
one ConfigMap resource (it's a generator of n maps).
The example below creates two ConfigMaps. One with the
names and contents of the given files, the other with
key/value as data.

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
```

env is a path to a file to read lines of key=val
you can only specify one env file per secret.

```
env: env.txt
type: Opaque
```

### generatorOptions
generatorOptions modify behavior of all ConfigMap and Secret generators

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