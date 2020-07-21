---
title: "crds"
linkTitle: "crds"
type: docs
description: >
    Adding CRD support
---

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

- "x-kubernetes-annotation": ""
- "x-kubernetes-label-selector": ""
- "x-kubernetes-identity": ""
- "x-kubernetes-object-ref-api-version": "v1",
- "x-kubernetes-object-ref-kind": "Secret",
- "x-kubernetes-object-ref-name-key": "name",

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

crds:
- crds/typeA.yaml
- crds/typeB.yaml
```
