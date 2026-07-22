---
title: "ReplacementTransformer"
linkTitle: "ReplacementTransformer"
weight: 6
date: 2024-02-12
description: >
  ReplacementTransformer substitutes values in target resources with values from source resources or static values.
---

See [Transformers]({{< relref "../Transformers" >}}) for common required fields.

* **apiVersion**: builtin
* **kind**: ReplacementTransformer
* **metadata** ([ObjectMeta](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta))

  Standard object's metadata.

* **replacements** ([]Replacement)

  List of replacements to execute. Each replacement can either be specified inline or loaded from a file path.

  * **path** (string)

    File path to external YAML/JSON configuration containing replacement specifications. Cannot be combined with inline `source` or `targets`.

  * **source** (SourceSelector)

    The source object and field path to extract value from.
    * **kind**, **group**, **version**, **name**, **namespace** (string): GVK and metadata to select the source resource.
    * **fieldPath** (string): Dot-separated path to the field value (e.g. `metadata.name`, `spec.template.spec.containers.[name=nginx].image`).
    * **options** (FieldOptions): Refines field extraction (e.g. `delimiter`, `index`).

  * **sourceValue** (string)

    Optional static string value to use as the source instead of extracting from a resource.

  * **targets** ([]TargetSelector)

    List of target resources and field paths to update with the source value.
    * **select** (Selector): Selects target resources matching `kind`, `group`, `version`, `name`, `namespace`.
    * **reject** ([]Selector): Optional list of selectors to exclude from matching.
    * **fieldPaths** ([]string): List of target field paths to write the value to.
    * **options** (FieldOptions): Refines field insertion (e.g., `delimiter`, `index`, `create` to auto-create missing target fields).
