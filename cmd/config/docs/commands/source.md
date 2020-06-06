## source

[Alpha] Implement a Source by reading a local directory.

### Synopsis

[Alpha] Implement a Source by reading a local directory.

    kustomize fn source DIR...

  DIR:
    One or more paths to local directories.  Contents from directories will be concatenated.
    If no directories are provided, source will read from stdin as if it were a single file.

`source` emits configuration to act as input to a function

### Examples

    # emity configuration directory as input source to a function
    kustomize fn source DIR/

    kustomize fn source DIR/ | your-function | kustomize fn sink DIR/
