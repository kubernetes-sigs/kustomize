---
title: "patches"
linkTitle: "patches"
type: docs
description: >
    Patch resources
---

Each entry in this list should resolve to patch object, which includes a patch and a target selector. 
The patch can be either a strategic merge patch or a JSON patch.  It can be either a patch file, or an inline
string. The target selects resources by group, version, kind, name, namespace, labelSelector and
annotationSelector. Any resource which matches all the specified fields has the patch applied.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

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