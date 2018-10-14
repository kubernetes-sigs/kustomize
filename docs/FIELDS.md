# List of fields

## Generators

### `configMapGenerator`

Each entry in this list results in the creation of
one ConfigMap resource (it's a generator of n maps).
The example below creates two ConfigMaps. One with the
names and contents of the given files, the other with
key/value as data.

Usage:

```yaml
configMapGenerator:
- name: myJavaServerEnvVars
  literals:	
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
- name: myJavaServerProps
  files:
  - application.properties
  - more.properties
```

### `secretGenerator`

Each entry in this list results in the creation of
one Secret resource (it's a generator of n secrets).
A command can do anything to get a secret,
e.g. prompt the user directly, start a webserver to
initate an oauth dance, etc.

Usage:

```yaml
secretGenerator:
- name: app-tls
  commands:
    tls.crt: "cat secret/tls.cert"
    tls.key: "cat secret/tls.key"
  type: "kubernetes.io/tls"
- name: downloaded_secret
  # timeoutSeconds specifies the number of seconds to
  # wait for the commands below. It defaults to 5 seconds.
  timeoutSeconds: 30
  commands:
    username: "curl -s https://path/to/secrets/username.yaml"
    password: "curl -s https://path/to/secrets/password.yaml"
  type: Opaque
- name: env_file_secret
  # envCommand is similar to command but outputs lines of key=val pairs
  # i.e. a Docker .env file or a .ini file.
  # you can only specify one envCommand per secret.
  envCommand: printf "DB_USERNAME=admin\nDB_PASSWORD=somepw"
  type: Opaque
```

## Input resources

### `resources`

Each entry in this list must resolve to an existing
resource definition in YAML.  These are the resource
files that kustomize reads, modifies and emits as a
YAML string, with resources separated by document
markers ("---").


Usage:

```yaml
resources:
- some-service.yaml
- ../some-dir/some-deployment.yaml
```

### `bases`

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

Usage:

```yaml
bases:
- ../../base
- github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
- github.com/Liujingfang1/mysql
- github.com/Liujingfang1/kustomize//examples/helloWorld?ref=test-branch
```

### `crds`

Each entry in this list should be a relative path to
a file for custom resource definition(CRD).

The presence of this field is to allow kustomize be
aware of CRDs and apply proper
transformation for any objects in those types.

Typical use case: A CRD object refers to a ConfigMap object.
In kustomization, the ConfigMap object name may change by adding namePrefix or hashing
The name reference for this ConfigMap object in CRD object need to be
updated with namePrefix or hashing in the same way.

Usage:

```yaml
crds:
- crds/typeA.yaml
- crds/typeB.yaml
```

## Patching

### `namespace`

Adds a namespace to all resources.

Usage: 

```yaml
namespace: my-namespace
```

### `namePrefix`

Value of this field is prepended to the
names of all resources, e.g. a deployment named
"wordpress" becomes "alices-wordpress".

Usage:

```yaml
namePrefix: alices-
```

### `commonLabels`

Labels to add to all resources and selectors.

Usage:

```yaml
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

### `commonAnnotations`

Annotations (non-identifying metadata)
to add to all resources.  Like labels,
these are key value pairs.


Usage:

```yaml
commonAnnotations:
  oncallPager: 800-555-1212
```

### `patches`

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

Usage:

```yaml
patches:
- service_port_8888.yaml
- deployment_increase_replicas.yaml
- deployment_increase_memory.yaml
```

### `patchesJson6902`

Each entry in this list should resolve to
a kubernetes object and a JSON patch that will be applied
to the object.
The JSON patch is documented at https://tools.ietf.org/html/rfc6902

target field points to a kubernetes object within the same kustomization
by the object's group, version, kind, name and namespace.
path field is a relative file path of a JSON patch file.
The content in this patch file can be either in JSON format as
```json
[
  {
    "op": "add", 
    "path": "/some/new/path",
    "value": "value"
  }, {
    "op": "replace",
    "path": "/some/existing/path",
    "value": "new value"
  }
]
```
 or in YAML format as
```yaml
- op: add
  path: /some/new/path
  value: value
- op:replace
  path: /some/existing/path
  value: new value
```

Usage:

```yaml
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

### `vars`

Vars are used to insert values from resources that cannot be referenced
otherwise. For example if you need to pass a Service's name to the arguments
or environment variables of a program but without hard coding the actual name
of the Service you'd insert `$(MY_SERVICE_NAME)` into the value field of the
env var or into the command or args of the container as shown here:
```
  containers:
    - image: myimage
      command: ["start", "--host", "$(MY_SERVICE_NAME)"]
      env:
        - name: SECRET_TOKEN
          value: $(SOME_SECRET_NAME)
```

Then you'll add an entry to `vars:` like shown below with the same name
and a reference to the resource from which to pull the field's value.
The actual field's path is optional and by default it will use
`metadata.name`. Currently only string type fields are supported, no integers
or booleans, etc. Also array access is currently not possible. For example getting
the image field of container number 2 inside of a pod can currently not be done.

Not every location of a variable is supported. To see a complete list of locations
see the file [refvars.go](https://github.com/kubernetes-sigs/kustomize/blob/master/pkg/transformers/refvars.go#L20).

An example of a situation where you'd not use vars is when you'd like to set a
pod's `serviceAccountName`. In that case you would just reference the ServiceAccount
by name and Kustomize will resolve it to the eventual name while building the manifests.

Usage:

```yaml
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

### `imageTags`

ImageTags modify the tags for images without creating patches.
E.g. Given this fragment of a Deployment:
```
containers:
  - name: myapp
    image: mycontainerregistry/myimage:v0
  - name: nginxapp
    image: nginx:1.7.9
```

It also supports digests. If digest is present newTag is ignored.

One can change the tag of myimage to v1 and the tag of nginx to 1.8.0 with the following:

Usage:

```yaml
imageTags:
  - name: mycontainerregistry/myimage
    newTag: v1
  - name: nginx
    newTag: 1.8.0
  - name: alpine
    digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
```
