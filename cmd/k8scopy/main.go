// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// k8scopy is a code reuse mechanism for k8s.io code.
//
// kustomize, kyaml, cmd/config are kustomize repo modules that want to reuse
// some small bits of k8s.io code. These modules cannot depend k8s.io via
// normal Go import or vendoring because kubectl will depend on these things
// eventually (see kubernetes-sigs/kustomize/issues/1500), and kubectl's code
// reuse is tricky. While kubectl remains in the k/k repo, it depends on local
// relative symlinked paths to a 'staging' version of k8s.io code.  No code
// imported by kubectl can refer to any other version of k8s.io code, not by
// Go importing, not by Go vendoring.
//
// This main exists to allow "go generate" to copy select k8s.io packages into
// the kustomize repo at well defined tags in reproducible fashion.  It's
// a form of vendoring reuse that avoids the problems created by k8s staging.
// The copied code is labelled as generated and is not otherwise edited.
//
// When/if kubectl is finally extracted from k/k to its own repo, it can
// depend on k8s.io code via normal imports, and then so can kustomize,
// so this technique can be dropped.
//
// Until then, if a bug is found in a particular instance of copied k8s.io
// (highly unlikely, since only old stable versions are copied), just update
// the version being copied, re-generate, and if need be adjust call points.
package main

import (
	"log"
	"os"

	"sigs.k8s.io/kustomize/cmd/k8scopy/internal"
)

const pgmName = "k8scopy"

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Need name of yaml file containing module specs and prefix.")
	}
	spec := internal.ReadSpec(os.Args[1])
	c := internal.NewCopier(spec, os.Args[2], pgmName)
	internal.RunNoOutputCommand("go", "get", spec.Name())
	for _, p := range spec.Packages {
		for _, n := range p.Files {
			if err := c.CopyFile(p.Name, n); err != nil {
				log.Fatal(err)
			}
		}
	}
	internal.RunNoOutputCommand(
		"go", "mod", "edit", "-droprequire="+spec.Module)
	internal.RunNoOutputCommand("go", "mod", "tidy")
	internal.RunGetOutputCommand("go", "fmt", "./...")
}
