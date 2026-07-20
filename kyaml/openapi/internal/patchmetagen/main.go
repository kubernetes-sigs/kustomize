// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Command patchmetagen generates zz_generated_patchmeta.go from one or more
// Kubernetes OpenAPI v2 documents (swagger.pb protobuf or swagger.json).
//
// The generated table is the union, across all input documents, of the
// strategic-merge-patch metadata the kyaml merge walker consumes:
// x-kubernetes-patch-strategy, x-kubernetes-patch-merge-key and
// x-kubernetes-list-map-keys, plus the minimal definition spine needed to
// descend from a resource root to each annotated field. Everything else in
// the schema (descriptions, validation, unannotated subtrees) is dropped —
// dropping them is behavior-preserving for merges because the walker treats
// "schema present without patch extensions" identically to "no schema"
// (see walkAssociativeSequence).
//
// Usage: go run ./openapi/internal/patchmetagen <swagger.pb|swagger.json>... > zz_generated_patchmeta.go
package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"google.golang.org/protobuf/proto"
	"k8s.io/kube-openapi/pkg/validation/spec"
	k8syaml "sigs.k8s.io/yaml"
)

type field struct {
	Strategy   string
	MergeKey   string
	MergeKeys  []string
	Ref        string // object field or map-value definition ref
	ElementRef string // array element definition ref
	IsArray    bool
	IsMap      bool
}

type gvk struct{ APIVersion, Kind string }

func refName(r spec.Ref) string {
	const p = "#/definitions/"
	s := r.String()
	if strings.HasPrefix(s, p) {
		return strings.TrimPrefix(s, p)
	}
	return ""
}

func load(path string) (spec.Definitions, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var swagger spec.Swagger
	if filepath.Ext(path) == ".pb" {
		doc := &openapi_v2.Document{}
		if err := proto.Unmarshal(b, doc); err != nil {
			return nil, err
		}
		if _, err := swagger.FromGnostic(doc); err != nil {
			return nil, err
		}
	} else {
		if len(b) > 0 && b[0] != '{' {
			if b, err = k8syaml.YAMLToJSON(b); err != nil {
				return nil, err
			}
		}
		if err := swagger.UnmarshalJSON(b); err != nil {
			return nil, err
		}
	}
	return swagger.Definitions, nil
}

func main() {
	k8sVersion := flag.String("k8s-version", "",
		"kubernetes/kubernetes tag the input swagger documents were taken from (required)")
	flag.Parse()
	if flag.NArg() < 1 || *k8sVersion == "" {
		fmt.Fprintln(os.Stderr, "usage: patchmetagen -k8s-version <tag> <swagger.pb|swagger.json>...")
		os.Exit(1)
	}

	defs := map[string]map[string]field{} // union across inputs
	gvks := map[gvk]string{}
	edges := map[string]map[string]bool{}

	for _, path := range flag.Args() {
		definitions, err := load(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loading %s: %v\n", path, err)
			os.Exit(1)
		}
		for name, d := range definitions {
			if exts, ok := d.Extensions["x-kubernetes-group-version-kind"]; ok {
				if list, ok := exts.([]interface{}); ok {
					for _, e := range list {
						m, ok := e.(map[string]interface{})
						if !ok {
							continue
						}
						g, _ := m["group"].(string)
						v, _ := m["version"].(string)
						k, _ := m["kind"].(string)
						av := v
						if g != "" {
							av = g + "/" + v
						}
						// first input wins per GVK (pass newest schema first)
						if _, seen := gvks[gvk{av, k}]; !seen {
							gvks[gvk{av, k}] = name
						}
					}
				}
			}
			if defs[name] == nil {
				defs[name] = map[string]field{}
				edges[name] = map[string]bool{}
			}
			for fname, p := range d.Properties {
				f := field{}
				// Property-level extensions survive the runtime Field()
				// lookup ONLY when the property carries no $ref: the legacy
				// resolve loop replaces the property schema (extensions and
				// all) with the referenced definition. Mirror that so the
				// table is bug-for-bug equivalent to the parsed schema.
				if refName(p.Ref) == "" {
					f.Strategy, _ = p.Extensions.GetString("x-kubernetes-patch-strategy")
					f.MergeKey, _ = p.Extensions.GetString("x-kubernetes-patch-merge-key")
					if raw, ok := p.Extensions["x-kubernetes-list-map-keys"]; ok {
						if list, ok := raw.([]interface{}); ok {
							for _, v := range list {
								if s, ok := v.(string); ok {
									f.MergeKeys = append(f.MergeKeys, s)
								}
							}
						}
					}
				}
				switch {
				case p.Items != nil && p.Items.Schema != nil:
					f.IsArray = true
					f.ElementRef = refName(p.Items.Schema.Ref)
				case p.AdditionalProperties != nil && p.AdditionalProperties.Schema != nil:
					f.IsMap = true
					f.Ref = refName(p.AdditionalProperties.Schema.Ref)
				default:
					f.Ref = refName(p.Ref)
				}
				// union: keep an existing entry with extensions over a bare one
				if old, ok := defs[name][fname]; !ok || (old.Strategy == "" && f.Strategy != "") {
					defs[name][fname] = f
				}
				for _, r := range []string{f.Ref, f.ElementRef} {
					if r != "" {
						edges[name][r] = true
					}
				}
			}
		}
	}

	// reverse reachability: keep only definitions that can reach a field
	// carrying a patch strategy
	keep := map[string]bool{}
	for name, fields := range defs {
		for _, f := range fields {
			if f.Strategy != "" {
				keep[name] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for name, kids := range edges {
			if keep[name] {
				continue
			}
			for k := range kids {
				if keep[k] {
					keep[name] = true
					changed = true
					break
				}
			}
		}
	}

	// prune: per kept definition, keep fields that carry a strategy or lead
	// into a kept definition
	pruned := map[string]map[string]field{}
	for name, fields := range defs {
		if !keep[name] {
			continue
		}
		out := map[string]field{}
		for fname, f := range fields {
			lead := (f.Ref != "" && keep[f.Ref]) || (f.ElementRef != "" && keep[f.ElementRef])
			if f.Strategy == "" && !lead {
				continue
			}
			// drop refs pointing outside the kept set
			if f.Ref != "" && !keep[f.Ref] {
				f.Ref = ""
			}
			if f.ElementRef != "" && !keep[f.ElementRef] {
				f.ElementRef = ""
			}
			out[fname] = f
		}
		pruned[name] = out
	}

	var b strings.Builder
	b.WriteString("// Copyright 2026 The Kubernetes Authors.\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Code generated by openapi/internal/patchmetagen; DO NOT EDIT.\n\n")
	b.WriteString("package openapi\n\n")
	b.WriteString("import \"sigs.k8s.io/kustomize/kyaml/yaml\"\n\n")
	fmt.Fprintf(&b, "// precomputedPatchMetaK8sVersion is the kubernetes/kubernetes tag whose\n")
	fmt.Fprintf(&b, "// openapi-spec/swagger.json this table was generated from.\n")
	fmt.Fprintf(&b, "const precomputedPatchMetaK8sVersion = %q\n\n", *k8sVersion)

	b.WriteString("// precomputedGVKToDef indexes every resource root in the builtin schema.\n")
	b.WriteString("// Types mapping to a definition absent from precomputedPatchDefs carry no\n")
	b.WriteString("// strategic-merge metadata anywhere in their tree.\n")
	b.WriteString("var precomputedGVKToDef = map[yaml.TypeMeta]string{\n")
	gvkKeys := make([]gvk, 0, len(gvks))
	for k := range gvks {
		gvkKeys = append(gvkKeys, k)
	}
	sort.Slice(gvkKeys, func(i, j int) bool {
		if gvkKeys[i].APIVersion != gvkKeys[j].APIVersion {
			return gvkKeys[i].APIVersion < gvkKeys[j].APIVersion
		}
		return gvkKeys[i].Kind < gvkKeys[j].Kind
	})
	for _, k := range gvkKeys {
		def := gvks[k]
		if !keep[def] {
			def = "" // sentinel: known type, no patch metadata anywhere below
		}
		fmt.Fprintf(&b, "\t{APIVersion: %q, Kind: %q}: %q,\n", k.APIVersion, k.Kind, def)
	}
	b.WriteString("}\n\n")

	b.WriteString("var precomputedPatchDefs = map[string]map[string]pmField{\n")
	defNames := make([]string, 0, len(pruned))
	for name := range pruned {
		defNames = append(defNames, name)
	}
	sort.Strings(defNames)
	for _, name := range defNames {
		fmt.Fprintf(&b, "\t%q: {\n", name)
		fields := pruned[name]
		fnames := make([]string, 0, len(fields))
		for fname := range fields {
			fnames = append(fnames, fname)
		}
		sort.Strings(fnames)
		for _, fname := range fnames {
			f := fields[fname]
			var parts []string
			if f.Strategy != "" {
				parts = append(parts, fmt.Sprintf("Strategy: %q", f.Strategy))
			}
			if f.MergeKey != "" {
				parts = append(parts, fmt.Sprintf("MergeKey: %q", f.MergeKey))
			}
			if len(f.MergeKeys) > 0 {
				parts = append(parts, fmt.Sprintf("MergeKeys: %#v", f.MergeKeys))
			}
			if f.Ref != "" {
				parts = append(parts, fmt.Sprintf("Ref: %q", f.Ref))
			}
			if f.ElementRef != "" {
				parts = append(parts, fmt.Sprintf("ElementRef: %q", f.ElementRef))
			}
			if f.IsArray {
				parts = append(parts, "IsArray: true")
			}
			if f.IsMap {
				parts = append(parts, "IsMap: true")
			}
			fmt.Fprintf(&b, "\t\t%q: {%s},\n", fname, strings.Join(parts, ", "))
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "formatting generated source: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(formatted)
}
