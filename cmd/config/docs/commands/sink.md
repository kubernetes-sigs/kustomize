## sink

[Alpha] Implement a Sink by writing input to a local directory.

### Synopsis

[Alpha] Implement a Sink by writing input to a local directory.

    kustomize config sink [DIR]

  DIR:
    Path to local directory.  If unspecified, sink will write to stdout as if it were a single file.

`sink` writes its input to a directory

### Examples

    kustomize config source DIR/ | your-function | kustomize config sink DIR/
