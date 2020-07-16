---
title: "configMapGenerator"
linkTitle: "configMapGenerator"
type: docs
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
