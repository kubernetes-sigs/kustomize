# Configuration IO API Semantics

Resource Configuration may be read / written from / to sources such as directories,
stdin|out or network. Tools may be composed using pipes such that the tools writing
Resource Configuration may be a different tool from the one that read the configuration.
In order for tools to be composed in this way, while preserving origin information --
such as the original file, index, etc.:

Tools **SHOULD** insert the following annotations when reading from sources,
and **SHOULD** delete the annotations when writing to sinks.

### `config.kubernetes.io/path`

Records the slash-delimited, OS-agnostic, relative file path to a Resource.

This annotation **SHOULD** be set when reading Resources from files.
It **SHOULD** be unset when writing Resources to files.
When writing Resources to a directory, the Resource **SHOULD** be written to the corresponding
path relative to that directory.

Example:

```yaml
metadata:
  annotations:
    config.kubernetes.io/path: "relative/file/path.yaml"
```

### `config.kubernetes.io/index`

Records the index of a Resource in file. In a multi-object YAML file, Resources are separated
by three dashes (`---`), and the index represents the positon of the Resource starting from zero.

This annotation **SHOULD** be set when reading Resources from files.
It **SHOULD** be unset when writing Resources to files.
When writing multiple Resources to the same file, the Resource **SHOULD** be written in the
relative order matching the index.

When this annotation is not specified, it implies a value of `0`.

Example:

```yaml
metadata:
  annotations:
    config.kubernetes.io/path: "relative/file/path.yaml"
    config.kubernetes.io/index: 2
```

This represents the third Resource in the file.

### `config.kubernetes.io/local-config`

`config.kubernetes.io/local-config` declares that the configuration is to local tools
rather than a remote Resource. e.g. The `Kustomization` config in a `kustomization.yaml`
**SHOULD** contain this annotation so that tools know it is not intended to be sent to
the Kubernetes api server.

Example:

```yaml
metadata:
  annotations:
    config.kubernetes.io/local-config: "true"
```
