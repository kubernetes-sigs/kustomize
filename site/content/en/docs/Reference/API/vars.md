---
title: "vars"
linkTitle: "vars"
type: docs
weight: 23
description: >
    Substitute name references.
---

[replacements]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/replacements/

{{% pageinfo color="warning" %}}
The `vars` field was deprecated in v5.0.0. This field will never be removed from the
kustomize.config.k8s.io/v1beta1 Kustomization API, but it will not be included
in the kustomize.config.k8s.io/v1 Kustomization API. When Kustomization v1 is available,
we will announce the deprecation of the v1beta1 version. There will be at least
two releases between deprecation and removal of Kustomization v1beta1 support from the
kustomize CLI, and removal itself will happen in a future major version bump.

Please try to migrate to the
the [replacements](/references/kustomize/kustomization/replacements) field. If you are
unable to restructure your configuration to use replacements instead of vars, please
ask for help in slack or file an issue for guidance.

We are experimentally attempting to
automatically convert `vars` to `replacements` with `kustomize edit fix --vars`. However,
converting vars to replacements in this way will potentially overwrite many resource files
and the resulting files may not produce the same output when `kustomize build` is run.
We recommend doing this in a clean git repository where the change is easy to undo.
{{% /pageinfo %}}

Vars are used to capture text from one resource's field
and insert that text elsewhere - a reflection feature.

For example, suppose one specifies the name of a k8s Service
object in a container's command line, and the name of a
k8s Secret object in a container's environment variable,
so that the following would work:

```yaml
containers:
  - image: myimage
    command: ["start", "--host", "$(MY_SERVICE_NAME)"]
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
```

To do so, add an entry to `vars:` as follows:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

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

The default config data for vars is at [/api/konfig/builtinpluginconsts/varreference.go](https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/varreference.go)
Long story short, the default targets are all
container command args and env value fields.

Vars should _not_ be used for inserting names in
places where kustomize is already handling that
job.  E.g., a Deployment may reference a ConfigMap
by name, and if kustomize changes the name of a
ConfigMap, it knows to change the name reference
in the Deployment.

### Convert vars to replacements

There are plans to deprecate vars, so we recommend migration to [replacements] as early as possible.

#### Simple migration example
Let's first take a simple example of how to manually do this conversion. Suppose we have a container
referencing secret (similar to the above example):

`pod.yaml`
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
```

and we are using vars as follows:

`kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml
- secret.yaml

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
```

In order to convert `vars` to `replacements`, we have to:
 1. Replace every instance of $(SOME_SECRET_NAME) with any arbitrary placeholder value.
 2. Convert the vars `objref` field to a [replacements] `source` field.
 3. Replace the vars `name` fied with a [replacements] `targets` field that points to
every instance of the placeholder value in step 1.

In our simple example here, this would look like the following:

`pod.yaml`
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_PLACEHOLDER_VALUE
```

`kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml
- secret.yaml

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - select:
      kind: Pod
      name: my-pod
    fieldPaths:
    - spec.containers.[name=hello].env.[name=SECRET_TOKEN].value
```

#### More complex migration example

Let's take a more complex usage of vars and convert it to [replacements]. We are going
to convert the vars in the [wordpress example](https://github.com/kubernetes-sigs/kustomize/tree/master/examples/wordpress)
to replacements.

The wordpress example has the following directory structure:

```
.
├── README.md
├── kustomization.yaml
├── mysql
│   ├── deployment.yaml
│   ├── kustomization.yaml
│   ├── secret.yaml
│   └── service.yaml
├── patch.yaml
└── wordpress
    ├── deployment.yaml
    ├── kustomization.yaml
    └── service.yaml
```

where `patch.yaml` has the following contents:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wordpress
spec:
  template:
    spec:
      initContainers:
      - name: init-command
        image: debian
        command: ["/bin/sh"]
        args: ["-c", "echo $(WORDPRESS_SERVICE); echo $(MYSQL_SERVICE)"]
      containers:
      - name: wordpress
        env:
        - name: WORDPRESS_DB_HOST
          value: $(MYSQL_SERVICE)
        - name: WORDPRESS_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-pass
              key: password
 ```

and the top level `kustomization.yaml` has the following contents:

 ```
 resources:
 - wordpress
 - mysql
 patchesStrategicMerge:
 - patch.yaml
 namePrefix: demo-

 vars:
 - name: WORDPRESS_SERVICE
   objref:
     kind: Service
     name: wordpress
     apiVersion: v1
 - name: MYSQL_SERVICE
   objref:
     kind: Service
     name: mysql
     apiVersion: v1
 ```

In this example, the patch is used to:
- Add an initial container to show the mysql service name
- Add environment variable that allow wordpress to find the mysql database

We can convert vars to replacements in this more complex case too, by taking the same steps as
the previous example. To do this, we can change the contents of `patch.yaml` to:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wordpress
spec:
  template:
    spec:
      initContainers:
      - name: init-command
        image: debian
        command: ["/bin/sh"]
        args: ["-c", "echo", "WORDPRESS_SERVICE", ";", "echo", "MYSQL_SERVICE"]
      containers:
      - name: wordpress
        env:
        - name: WORDPRESS_DB_HOST
          value: MYSQL_SERVICE
        - name: WORDPRESS_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-pass
              key: password

 ```

Then, in our kustomization, we can have our replacements:

`kustomization.yaml`

```yaml
resources:
- wordpress
- mysql
patchesStrategicMerge:
- patch.yaml
namePrefix: demo-

replacements:
- source:
    name: demo-wordpress
    kind: Service
    version: v1
  targets:
  - select:
      kind: Deployment
      name: demo-wordpress
    fieldPaths:
    - spec.template.spec.initContainers.[name=init-command].args.2
- source:
    name: demo-mysql
    kind: Service
    version: v1
  targets:
  - select:
      kind: Deployment
      name: demo-wordpress
    fieldPaths:
    - spec.template.spec.initContainers.[name=init-command].args.5
    - spec.template.spec.containers.[name=wordpress].env.[name=WORDPRESS_DB_HOST].value
```
