---
title: "GeneratorArgs"
linkTitle: "GeneratorAgs"
weight: 2
date: 2023-11-15
description: >
  GeneratorArgs contains arguments common to generators.
---

* **name** (string)

    Name of the generated ConfigMap. A hash suffix will be added to the Name by default.

* **namespace** (string)

    Namespace of the generated ConfigMap.

* **behavior** (string)

    Behavior of generated resource, must be one of:

    * **create**: Create new ConfigMap
    * **replace**: Replace existing ConfigMap
    * **merge**: Merge with existing ConfigMap


* **literals** ([]string)

    List of string literal pair sources. Each literal source should be a key and literal value, e.g. `key=value`.

* **files** ([]string)

    List of files paths to use in creating a list of key value pairs. A source should be in the form [{key}=]{path}. If the `key=` argument is not provided, the key is the path's basename. If the `key=` argument is provided, it becomes the key. The value is the file contents. Specifying a directory will iterate each named file in the directory whose basename is a valid ConfigMap key.

* **envs** ([]string)

    List of file paths. The contents of each file should be one key=value pair per line. Additionally, npm `.env` and `.ini` files are supported.

* **env** (string)

    Env is the singular form of `envs`. This is merged with `env` on edits with `kustomize fix` for consistency with `literals` and `files`.

* **options** ([GeneratorOptions](#generatoroptions))

    Options override global `generatorOptions` field.
