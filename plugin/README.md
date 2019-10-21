This directory and the directories below it
are not _importable_ Go packages.

Each directory contains a kustomize plugin,
which is either 

* some random non-Go based executable, 
  e.g. a bash script that runs java
  in a JVM, or python in a python VM, etc.

* the source code for a Go executable,
  i.e. a `main.go` in the unimportable
  `main` package,
  ideally declaring its dependencies
  with it's own `go.mod` file.

* the source code for a Go
  plugin, which is also an unimportable
  `main` package ideally with its
  own `go.mod` file, formulated to be
  a Go plugin.  If in the `builtin`
  sub-directory, these plugins are converted
  to statically linkable code.

To read more about plugins, see
sigs.k8s.io/kustomize/api/plugins/doc.go

