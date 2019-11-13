# Configuration IO API Semantics

  Resource Configuration may be read / written from / to sources such as directories,
  stdin|out or network.  Tools may be composed using pipes such that the tools writing
  Resource Configuration may be a different tool from the one that read the configuration.
  In order for tools to be composed in this way, while preserving origin information --
  such as the original file, index, etc.

  Tools **SHOULD** write the following annotations when reading from sources,
  and **SHOULD** respect the annotations when writing to syncs.

### `config.kubernetes.io/path`

  `config.kubernetes.io/path` records a relative path on a Resource.  This annotation
  **SHOULD** be set when reading Resources from files.
  When writing Resources to a directory, the Resource **SHOULD** be written to the corresponding
  path relative to that directory.

  Example:

    metadata:
      annotations:
        config.kubernetes.io/path: "relative/file/path.yaml"

### `config.kubernetes.io/index`

  `config.kubernetes.io/index` records the index of a Resource into a file which may contain
  multiple Resource.  This annotation  **SHOULD** be set when reading Resources from files.
  When writing multiple Resources to the same file, the Resource **SHOULD** be written in the
  relative order matching the index.

  Example:

    metadata:
      annotations:
        config.kubernetes.io/index: "0"

### `config.kubernetes.io/local-config`

  `config.kubernetes.io/local-config` declares that the configuration is to local tools
  rather than a remote Resource.  e.g. The `Kustomization` config in a `kustomization.yaml`
  **SHOULD** contain this annotation so that tools know it is not intended to be sent to
  the Kubernetes api server.
  
  Example:
  
    metadata:
      annotations:
        config.kubernetes.io/local-config: "true"
