---
title: "generatorOptions"
linkTitle: "generatorOptions"
type: docs
weight: 8
description: >
    Modify behavior of generators.
---

`apiVersion: kustomize.config.k8s.io/v1beta1`

### generatorOptions
GeneratorOptions modifies resource generation behavior.

---

* **labels** (map[string]string), optional

  Labels to add to all generated resources.

* **annotations** (map[string]string), optional

  Annotations to add to all generated resources.

* **disableNameSuffixHash** (bool), optional

  DisableNameSuffixHash if true disables the default behavior of adding a suffix to the names of generated resources that is a hash of the resource contents.

* **immutable** (bool), optional

  Immutable if true add to all generated resources.
