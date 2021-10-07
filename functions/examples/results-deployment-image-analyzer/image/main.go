// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements a validator function run by `kustomize fn run`
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	p := framework.SimpleProcessor{
		Filter: kio.FilterFunc(func(in []*yaml.RNode) ([]*yaml.RNode, error) {
			// An analyzer would make no changes to the RNodes so it echoes the same input out.
			return in, imageTagAnalyzer(in)
		}),
	}
	rw := &kio.ByteReadWriter{
		WrappingKind:       kio.ResourceListKind,
		WrappingAPIVersion: kio.ResourceListAPIVersion,
	}
	if err := framework.Execute(p, rw); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func imageTagAnalyzer(in []*yaml.RNode) error {
	deployments, err := (&framework.Selector{
		Kinds:       []string{"Deployment"},
		APIVersions: []string{"apps/v1"},
	}).Filter(in)
	if err != nil {
		return err
	}

	var results framework.Results
	for _, resource := range deployments {
		r, err := validate(resource)
		if err != nil {
			return err
		}
		results = append(results, r...)
	}
	return results
}

func validate(r *yaml.RNode) (framework.Results, error) {
	meta, err := r.GetMeta()
	if err != nil {
		return nil, err
	}

	// lookup the containers field in the Resource
	containersNode, err := r.Pipe(yaml.Lookup("spec", "template", "spec", "containers"))
	if err != nil {
		s, _ := r.String()
		return nil, fmt.Errorf("%v: %s", err, s)
	}
	if containersNode == nil {
		// doesn't have containers, ignore it
		return nil, nil
	}

	containers, _ := containersNode.Elements()
	var out framework.Results
	for i, node := range containers {
		fileIndex, err := strconv.Atoi(meta.Annotations[kioutil.IndexAnnotation])
		if err != nil {
			return nil, err
		}

		imageField := node.Field("image")
		if imageField == nil || imageField.Value == nil {
			continue
		}
		image := yaml.GetValue(imageField.Value)
		if strings.Contains(image, ":") {
			continue
		}

		out = append(out, &framework.Result{
			Message:  "missing image version",
			Severity: framework.Error,
			ResourceRef: &yaml.ResourceIdentifier{
				TypeMeta: meta.TypeMeta,
				NameMeta: meta.NameMeta,
			},
			Field: &framework.Field{
				Path:          fmt.Sprintf("spec.template.spec.containers[%d].image", i),
				CurrentValue:  image,
				ProposedValue: image + ":<VERSION>",
			},
			File: &framework.File{
				Path:  meta.Annotations[kioutil.PathAnnotation],
				Index: fileIndex,
			},
		})
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}
