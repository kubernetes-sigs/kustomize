// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package plugin contains builtin and example
plugins, tests and test libraries, and a code
generator for converting a plugin to statically
loadable code (see pluginator).


HOW PLUGINS RUN

Assume a file 'secGen.yaml' containing

   apiVersion: someteam.example.com/v1
   kind: SecretGenerator
   metadata:
     name: makesecrets
   name: mySecret
   behavior: merge
   envs:
   - db.env
   - fruit.env

If this file were referenced by a kustomization
file in its 'generators' field, kustomize would

 * Read 'secGen.yaml'.

 * Use the value of $XGD_CONFIG_HOME and
   'apiversion' and to find an executable
   named 'SecretGenerator' to use as
   an exec plugin, or failing that,

 * use the same info to load a Go plugin
   object file called 'SecretGenerator.so'.

 * Send either the file name 'secGen.yaml' as
   the first arg to the exec plugin, or send its
   contents to the go plugin's Config method.

 * Use the plugin to generate and/or transform.


GO PLUGINS

A .go file can be a Go plugin if it declares
'main' as it's package, and exports a symbol to
which useful functions are attached.

It can further be used as a _kustomize_ plugin if
the symbol is named 'KustomizePlugin' and the
attached functions implement the Configurable,
Generator and Transformer interfaces.

A plugin won't load into some program foo/main.go
if there is any package version mismatch in the
dependencies of the plugin and the dependencies of
foo/main.go.  Control this with matching
declarations in go.mod files.  The versions of the
builtin packages "fmt", "io", "os" (not normally
listed in go.mod) etc have the same version as the
compiler.


ONE PLUGIN PER DIRECTORY

For kustomize (and perhaps anyone), it's simplest
to put each plugin into its own directory.

Go plugins must be in package `main`, and so
having more than one plugin in a directory means
their loading symbols have to differ, which makes
it hard to standardize around how they get loaded,
or it means one must use build tags to suppress
full directory compilation - which creates
difficulties using IDEs, the `go mod` tool, `go
test ./...`, etc.

A one plugin per directory policy makes it easy to
define the plugin as a module, with its own
`go.mod` file - which is vital for resolving
package version dependency mismatches at load
time.  It also makes it easy to create a plugin
tarball (source, test, go.mod, plugin data files,
etc.) for distribution.


BUILTIN PLUGIN CONFIGURATION

For performance reasons, all builting plugins are
Go plugins (not exec plugins).

Using "SecretGenerator" as an example in what
follows.

The plugin config file looks like

  ---------------------------------------------
  apiVersion: builtin
  kind: SecretGenerator
  metadata:
    name: whatever
  otherField1: whatever
  otherField2: whatever
  ...
  ---------------------------------------------

The apiVersion must be 'builtin'.  The kind is the
CamelCase name of the plugin.

For non-builtins the apiVersion can be any legal
apiVersion value, e.g. 'someteam.example.com/v1beta1'

The builtin source must be at:

  repo=$GOPATH/src/sigs.k8s.io/kustomize
  ${repo}/plugin/${apiVersion}/LOWERCASE(${kind})/${kind}.go

(dropping the ".go" for exec plugins).

k8s wants 'kind' values to follow CamelCase, while
Go style doesn't like but does allow such names.

The lowercased value of kind is used as the name of the
directory holding the plugin, its test, and any
optional associated files (possibly a go.mod file).


PLUGIN SOURCE

 * Pattern

   secretgenerator.go
   ---------------------------------------------
   //go:generate go run sigs.k8s.io/kustomize/cmd/pluginator
   package main
   import ...
   type plugin struct{...}
   var KustomizePlugin plugin
   func (p *plugin) Config(
      ldr ifc.Loader,
      rf *resmap.Factory,
      c []byte) error {...}
   func (p *plugin) Generate(
      ) (resmap.ResMap, error) {...}
   func (p *plugin) Transform(
      m resmap.ResMap) error {...}
   ---------------------------------------------

   The plugin name doesn't appear in the file itself.

 * Compilation

     repo=$GOPATH/src/sigs.k8s.io/kustomize
     dir=$repo/plugin/builtin
     go build -buildmode plugin \
         -o $dir/secretgenerator.so \
         $dir/secretgenerator.go



BUILTIN PLUGIN GENERATION

The pluginator program is a code generator that
converts kustomize generator (G) and/or
transformer (T) Go plugins to statically linkable
code.

It arises from following requirements:

 * extension

     kustomize does two things - generate or
     transform k8s resources.  Plugins let
     users write their own G&T's without
     having to fork kustomize and learn its
     internals.

 * dogfooding

     A G&T extension framework one can trust
     should be used by its authors to deliver
     builtin G&T's.

 * distribution

     kustomize should be distributable via
     `go get` and should run where Go
     programs are expected to run.

 The extension requirement led to the creation
 of a framework that accommodates writing a
 G or T as either

   * an 'exec' plugin (any executable file
     runnable as a kustomize subprocess), or

   * as a Go plugin - see
     https://golang.org/pkg/plugin.

 The dogfooding (and an implicit performance
 requirement) requires a 'builtin' G or T to
 be written as a Go plugin.

 The distribution ('go get') requirement demands
 conversion of Go plugins to statically linked
 code, hence this program.


TO GENERATE CODE

  repo=$GOPATH/src/sigs.k8s.io/kustomize
  cd $repo/plugin/builtin
  go generate ./...

This creates

  $repo/plugin/builtin/SecretGenerator.go

etc.

Generated plugins are used in kustomize via

  ---------------------------------------------
  package whatever
  import "sigs.k8s.io/kustomize/plugin/builtin
  ...
  g := builtin.NewSecretGenerator()
  g.Config(l, rf, k)
  resources, err := g.Generate()
  err = g.Transform(resources)
  // Eventually emit resources.
  ---------------------------------------------

*/
package plugin
