---
title: "vars"
linkTitle: "vars"
type: docs
description: >
    Substitute name references.
---

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
