---
title: "FieldSpec"
linkTitle: "FieldSpec"
weight: 1
date: 2023-07-28
description: >
  FieldSpec specifies a field for Kustomize to target.
---

* **group** (string)

  Kubernetes group that this FieldSpec applies to.
  If empty, this FieldSpec applies to all groups.
  Currently, there is no way to specify only the core group, which is also represented by the empty string.

* **version** (string)

  Kubernetes version that this FieldSpec applies to.
  If empty, this FieldSpec applies to all versions.

* **kind** (string)

  Kubernetes kind that this FieldSpec applies to.
  If empty, this FieldSpec applies to all kinds.

* **path** (string)

  Path to target field. Fields in path are delimited by forward slashes "/".

* **create** (bool)

  If true, creates fields in **path** not already present.