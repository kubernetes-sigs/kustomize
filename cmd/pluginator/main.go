/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// pluginator is a code generator that converts
// kustomize generator (G) and/or transformer (T)
// Go plugins to statically linkable code.
//
// Arises from following requirements:
//
// * extension
//
//     kustomize does two things - generate or
//     transform k8s resources.  Plugins let
//     users write their own G&T's without
//     having to fork kustomize and learn its
//     internals.
//
// * dogfooding
//
//     A G&T extension framework one can trust
//     should be used by its authors to deliver
//     builtin G&T's.
//
// * distribution
//
//     kustomize should be distributable via
//     `go get` and should run where Go
//     programs are expected to run.
//
// The extension requirement led to the creation
// of a framework that accommodates writing a
// G or T as either
//
//   * an 'exec' plugin (any executable file
//     runnable as a kustomize subprocess), or
//
//   * as a Go plugin - see
//     https://golang.org/pkg/plugin.
//
// The dogfooding (and an implicit performance
// requirement) requires a 'builtin' G or T to
// be written as a Go plugin.
//
// The distribution ('go get') requirement demands
// conversion of Go plugins to statically linked
// code, hence this program.
//
//
// HOW PLUGINS RUN
//
// Assume a file 'secGen.yaml' containing
//
//   apiVersion: someteam.example.com/v1
//   kind: SecretGenerator
//   metadata:
//     name: makesecrets
//   name: mySecret
//   behavior: merge
//   envs:
//   - db.env
//   - fruit.env
//
// If this file were referenced by a kustomization
// file in its 'generators' field, kustomize would
//
// * Read 'secGen.yaml'.
//
// * Use the value of $XGD_CONFIG_HOME and
//   'apiversion' and to find an executable
//   named 'SecretGenerator' to use as
//   an exec plugin, or failing that,
//
// * use the same info to load a Go plugin
//   object file called 'SecretGenerator.so'.
//
// * Send either the file name 'secGen.yaml' as
//   the first arg to the exec plugin, or send its
//   contents to the go plugin's Config method.
//
// * Use the plugin to generate and/or transform.
//
//
// GO PLUGIN CONVENTIONS
//
// A .go file can be a Go plugin if it declares
// 'main' as it's package, and exports a symbol to
// which useful functions are attached. It can
// further be used as a _kustomize_ plugin if
// those functions implement the Configurable,
// Generator and Transformer interfaces.
//
// Converting the plugin file to a normal .go package
// file is a matter string substitution permitted
// by the following conventions.
//
// * Configuration of builtin plugins:
//
//     Config file looks like
//
//     ---------------------------------------------
//     apiVersion: builtin
//     kind: SecretGenerator
//     metadata:
//       name: whatever
//     otherFields: whatever
//     ---------------------------------------------
//
//     The apiVersion must be 'builtin'.
//
//     For non-builtins the apiVersion can be any legal
//     apiVersion value, e.g. 'someteam.example.com/v1beta1'
//
//     The builtin source must be at:
//
//        ${repo}/plugin/${apiVersion}/${kind}.go
//
//     where repo=$GOPATH/src/sigs.k8s.io/kustomize
//
//     k8s wants 'kind' values to follow CamelCase,
//     while Go style wants (but doesn't demand)
//     lowercase file names.
//
//     kustomize will accept either idiom, but the Go file
//     name must be ${kind}.go (CamelCase allowed).
//
// * Source follows this pattern
//
//     ---------------------------------------------
//     // +build plugin
//
//     //go:generate go run sigs.k8s.io/kustomize/cmd/pluginator
//     package main
//     import ...
//     type plugin struct{...}
//     var KustomizePlugin plugin
//     func (p *plugin) Config(
//        ldr ifc.Loader, rf *resmap.Factory,
//        k ifc.Kunstructured) error {...}
//     func (p *plugin) Generate(
//        ) (resmap.ResMap, error) {...}
//     func (p *plugin) Transform(
//        m resmap.ResMap) error {...}
//     ---------------------------------------------
//
//     - The 2nd line must be empty.
//     - One may `go fmt` this file.
//     - There's no mention of 'SecretGenerator'
//       in this file; that binding is done by
//       the plugin loader or pluginator.
//
// * To compile this for loading as a Go plugin:
//
//     repo=$GOPATH/src/sigs.k8s.io/kustomize
//     dir=$repo/plugin/builtin
//     go build -buildmode plugin -tags=plugin \
//         -o $dir/SecretGenerator.so \
//         $dir/SecretGenerator.go
//
// * To generate code:
//
//     repo=$GOPATH/src/sigs.k8s.io/kustomize
//     cd $repo/plugin/builtin
//     go generate --tags plugin .
//
//   This creates
//
//     $repo/plugin/builtingen/SecretGenerator.go
//
//   etc.
//
// * Generated plugins are used in kustomize via
//
//     ---------------------------------------------
//     package whatever
//     import "sigs.k8s.io/kustomize/plugin/builtingen
//     ...
//     g := builtingen.NewSecretGenerator()
//     g.Config(l, rf, k)
//     resources, err := g.Generate()
//     err = g.Transform(resources)
//     // Eventually emit resources.
//     ---------------------------------------------
//
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/plugins"
)

func main() {
	root := fileRoot()
	file, err := os.Open(root + ".go")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	processBoilerPlate(scanner, file.Name())

	w := NewWriter(root)
	defer w.close()

	// This particular phrasing is required.
	w.write(
		fmt.Sprintf(
			"// Code generated by pluginator on %s; DO NOT EDIT.",
			root))
	w.write("package builtingen")

	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "//go:generate") {
			continue
		}
		if l == "var "+plugins.PluginSymbol+" plugin" {
			w.write("func New" + root + "Plugin() *" + root + "Plugin {")
			w.write("  return &" + root + "Plugin{}")
			w.write("}")
			continue
		}
		w.write(l)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func fileRoot() string {
	n := os.Getenv("GOFILE")
	if !strings.HasSuffix(n, ".go") {
		log.Fatalf("expecting .go suffix on %s", n)
	}
	return n[:len(n)-len(".go")]
}

func processBoilerPlate(s *bufio.Scanner, f string) {
	if !s.Scan() {
		log.Fatalf("1: %s not long enough", f)
	}
	first := s.Text()
	if !s.Scan() {
		log.Fatalf("2: %s not long enough", f)
	}
	next := s.Text()
	if !hasPluginTag(first, next) {
		log.Fatalf("%s lacks plugin tag", f)
	}
	gotMain := false
	for !gotMain && s.Scan() {
		next = s.Text()
		gotMain = strings.HasPrefix(next, "package main")
	}
	if !gotMain {
		log.Fatalf("%s missing package main", f)
	}
}

func hasPluginTag(first, next string) bool {
	return strings.HasPrefix(first, "// +build plugin") &&
		len(next) == 0
}

type writer struct {
	root string
	f    *os.File
}

func NewWriter(r string) *writer {
	n := makeSrcFileName(r)
	f, err := os.Create(n)
	if err != nil {
		log.Fatalf("unable to create `%s`; %v", n, err)
	}
	return &writer{root: r, f: f}
}

func makeSrcFileName(root string) string {
	return filepath.Join(
		os.Getenv("GOPATH"),
		"src",
		pgmconfig.DomainName,
		pgmconfig.ProgramName,
		pgmconfig.PluginRoot,
		"builtingen",
		root+".go")
}

func (w *writer) close() { w.f.Close() }

func (w *writer) write(line string) {
	_, err := w.f.WriteString(w.filter(line) + "\n")
	if err != nil {
		log.Printf("Trouble writing: %s", line)
		log.Fatal(err)
	}
}

func (w *writer) filter(in string) string {
	if ok, newer := w.replace(in, "type plugin struct"); ok {
		return newer
	}
	if ok, newer := w.replace(in, "*plugin)"); ok {
		return newer
	}
	return in
}

// replace 'plugin' with 'FooPlugin' in context
// sensitive manner.
func (w *writer) replace(in, target string) (bool, string) {
	if !strings.Contains(in, target) {
		return false, ""
	}
	newer := strings.Replace(
		target, "plugin", w.root+"Plugin", 1)
	return true, strings.Replace(in, target, newer, 1)
}
