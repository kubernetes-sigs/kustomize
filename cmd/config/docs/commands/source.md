## source

[Alpha] Implement a Source by reading a local directory.

### Synopsis

[Alpha] Implement a Source by reading a local directory.

    kustomize config source DIR

  DIR:
    Path to local directory.

`source` emits configuration to act as input to a function

### Examples

    # emity configuration directory as input source to a function
    kustomize config source DIR/

    kustomize config source DIR/ | your-function | kustomize config sink DIR/
