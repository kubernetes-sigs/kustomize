---
title: "NamespaceTransformer"
linkTitle: "NamespaceTransformer"
weight: 3
date: 2024-02-12
description: >
  NamespaceTransformer sets the Namespace of user-input namespaced resources.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

* **apiVersion**: builtin
* **kind**: NamespaceTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **fieldSpecs** (\[\][FieldSpec]({{< relref "../Common%20Definitions/FieldSpec.md" >}})), optional

  fieldSpecs allows the user to re-define the field path of the Kubernetes Namespace field from `metadata/namespace` for different resources.

  If not specified, NamespaceTransformer applies the namespace to the `metadata/namespace` field of all resources.

* **unsetOnly** (bool), optional

  UnsetOnly indicates whether the NamespaceTransformer will only set namespace fields that are currently unset. Defaults to false.

* **setRoleBindingSubjects** (RoleBindingSubjectMode), optional

  SetRoleBindingSubjects determines which subject fields in RoleBinding and ClusterRoleBinding objects will have their namespace fields set. Overrides field specs provided for these types.

  _RoleBindingSubjectMode specifies which subjects will be set. It can be one of three possible values:_

  - `defaultOnly` (default): namespace will be set only on subjects named "default".
  - `allServiceAccounts`: Namespace will be set on all subjects with `kind: ServiceAccount`.
  - `none`: All subjects will be skipped.
