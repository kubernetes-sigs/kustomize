// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

/*

Read docs/plugins.md first for an overview of kustomize plugins.


BUILTIN PLUGIN CONFIGURATION

There are two kinds of plugins, Go plugins (shared
object library) and exec plugins (independent binary).
For performance and standardized testing reasons, all
builtin plugins are Go plugins (not exec plugins).

Using "SecretGenerator" as an example in what
follows.

The plugin config file looks like

  apiVersion: builtin
  kind: SecretGenerator
  metadata:
    name: whatever
  otherField1: whatever
  otherField2: whatever
  ...

The apiVersion must be 'builtin'.

The kind is the CamelCase name of the plugin.

The source for a builtin plugin must be at:

  repo=$GOPATH/src/sigs.k8s.io/kustomize
  ${repo}/plugin/builtin/LOWERCASE(${kind})/${kind}

k8s wants 'kind' values to follow CamelCase, while
Go style doesn't like but does allow such names.

The lowercased value of kind is used as the name of the
directory holding the plugin, its test, and any
optional associated files (possibly a go.mod file).


BUILTIN PLUGIN GENERATION

The `pluginator` program is a code generator that
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

The extension requirement led to building
a framework that accommodates writing a
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

See scripts/kyaml-pre-commit.sh for canonical way
to execute the above.

This creates

  $repo/api/plugins/builtins/SecretGenerator.go

etc.

Generated plugins are used in kustomize via

  package whatever
  import sigs.k8s.io/kustomize/api/plugins/builtins
  ...
  g := builtin.NewSecretGenerator()
  g.Config(h, k)
  resources, err := g.Generate()
  err = g.Transform(resources)
  // Eventually emit resources.

*/
package plugins
