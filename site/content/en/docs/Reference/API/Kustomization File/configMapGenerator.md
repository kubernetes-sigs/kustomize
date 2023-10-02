---
title: "configMapGenerator"
linkTitle: "configMapGenerator"
type: docs
weight: 6
description: >
    Generate ConfigMap resources.
---

Each entry in this list results in the creation of
one ConfigMap resource (it's a generator of n maps).

The example below creates four ConfigMaps:

- first, with the names and contents of the given files
- second, with key/value as data using key/value pairs from files
- third, also with key/value as data, directly specified using `literals`
- and a fourth, which sets an annotation and label via `options` for that single ConfigMap

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

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

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
- name: my-java-server-env-file-vars
  envs:
  - my-server-env.properties
  - more-server-props.env
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

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

configMapGenerator:
- name: app-whatever
  files:
  - myFileName.ini=whatever.ini
```

## ConfigMap `from File`

ConfigMap Resources may be generated from files - such as a java `.properties` file.  To generate a ConfigMap
Resource for a file, add an entry to `configMapGenerator` with the filename.

**Example:** Generate a ConfigMap with a data item containing the contents of a file.

The ConfigMaps will have data values populated from the file contents.  The contents of each file will
appear as a single data item in the ConfigMap keyed by the filename.

The example illustrates how you can create ConfigMaps from File using Generators.

### File Input

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: my-application-properties
  files:
  - application.properties

```

```yaml
# application.properties
FOO=Bar
```

### Build Output

```yaml
apiVersion: v1
data:
  application.properties: |-
    FOO=Bar
kind: ConfigMap
metadata:
  name: my-application-properties-f7mm6mhf59
```

## ConfigMap `from Literals`

ConfigMap Resources may be generated from literal key-value pairs - such as `JAVA_HOME=/opt/java/jdk`.
To generate a ConfigMap Resource from literal key-value pairs, add an entry to `configMapGenerator` with a
list of `literals`.

{{< alert color="success" title="Literal Syntax" >}}
- The key/value are separated by a `=` sign (left side is the key)
- The value of each literal will appear as a data item in the ConfigMap keyed by its key.
{{< /alert >}}

**Example:** Create a ConfigMap with 2 data items generated from literals.

The example illustrates how you can create ConfigMaps from Literals using Generators.

### File Input

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: my-java-server-env-vars
  literals:
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof

```

### Build Output

```yaml
apiVersion: v1
data:
  JAVA_HOME: /opt/java/jdk
  JAVA_TOOL_OPTIONS: -agentlib:hprof
kind: ConfigMap
metadata:
  name: my-java-server-env-vars-44k658k8gk
```

## ConfigMap `from env file`

ConfigMap Resources may be generated from key-value pairs much the same as using the literals option
but taking the key-value pairs from an environment file. These generally end in `.env`.
To generate a ConfigMap Resource from an environment file, add an entry to `configMapGenerator` with a
single `envs` entry, e.g. `envs: [ 'config.env' ]`.

{{< alert color="success" title="Environment File Syntax" >}}
- The key/value pairs inside of the environment file are separated by a `=` sign (left side is the key)
- The value of each line will appear as a data item in the ConfigMap keyed by its key.
- Pairs may span a single line only.
{{< /alert >}}

**Example:** Create a ConfigMap with 3 data items generated from an environment file.

### File Input

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: tracing-options
  envs:
  - tracing.env
```

```bash
# tracing.env
ENABLE_TRACING=true
SAMPLER_TYPE=probabilistic
SAMPLER_PARAMETERS=0.1
```

### Build Output

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  # The name has had a suffix applied
  name: tracing-options-6bh8gkdf7k
# The data has been populated from each literal pair
data:
  ENABLE_TRACING: "true"
  SAMPLER_TYPE: "probabilistic"
  SAMPLER_PARAMETERS: "0.1"
```

## Overriding Base ConfigMap Values

ConfigMap values from bases may be overridden by adding another generator for the ConfigMap
in the overlay and specifying the `behavior` field.  `behavior` may be
one of:
* `create` (default value): used to create a new ConfigMap. A name conflict error will be thrown if a ConfigMap with the same name and namespace already exists.
* `replace`: replace an existing ConfigMap from the base.
* `merge`: add or update the values in an existing ConfigMap from the base.

When updating an existing ConfigMap with the `merge` or `replace` strategies, you must ensure that both the name and namespace match the ConfigMap you're targeting. For example, if the namespace is unspecified in the base, you should not specify it in the overlay. Conversely, if it is specified in the base, you must specify it in the overlay as well. This is true even if the overlay Kustomization includes a namespace, because configMapGenerator runs before the namespace transformer.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: my-new-namespace

resources:
- ../base

configMapGenerator:
  - name: existing-name
    namespace: existing-ns # needs to match target ConfigMap from base
    behavior: replace
    literals:
      - ENV=dev
```

{{< alert color="warning" title="Name suffixing with overlay configMapGenerator" >}}
When using configMapGenerator to override values of an existing ConfigMap, the overlay configMapGenerator does not cause suffixing of the existing ConfigMap's name to occur. To take advantage of name suffixing, use configMapGenerator in the base, and the overlay generator will correctly update the suffix based on the new content. 
{{< /alert >}}

## Propagating the Name Suffix

Workloads that reference the ConfigMap or Secret will need to know the name of the generated Resource,
including the suffix. Kustomize takes care of this automatically by identifying
references to generated ConfigMaps and Secrets, and updating them.

In the following example, the generated ConfigMap name will be `my-java-server-env-vars` with a suffix unique to its contents.
Changes to the contents will change the name suffix, resulting in the creation of a new ConfigMap,
which Kustomize will transform Workloads to point to.

The PodTemplate volume references the ConfigMap by the name specified in the generator (excluding the suffix).
Kustomize will update the name to include the suffix applied to the ConfigMap name.

**Input:** The kustomization.yaml and deployment.yaml files

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: my-java-server-env-vars
  literals:
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
resources:
- deployment.yaml
```

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app: test
spec:
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: container
        image: registry.k8s.io/busybox
        command: [ "/bin/sh", "-c", "ls /etc/config/" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
      volumes:
      - name: config-volume
        configMap:
          name: my-java-server-env-vars
```

**Result:** The output of the Kustomize build.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  # The name has been updated to include the suffix
  name: my-java-server-env-vars-k44mhd6h5f
data:
  JAVA_HOME: /opt/java/jdk
  JAVA_TOOL_OPTIONS: -agentlib:hprof
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test
  name: test-deployment
spec:
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - ls /etc/config/
        image: registry.k8s.io/busybox
        name: container
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      volumes:
      - configMap:
          # The name has been updated to include the
          # suffix matching the ConfigMap
          name: my-java-server-env-vars-k44mhd6h5f
        name: config-volume
```
