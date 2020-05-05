// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	var stringValue string
	var intValue int
	var boolValue bool
	cmd := framework.Command(nil, func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			// set the annotation on each resource item
			if err := items[i].PipeE(yaml.SetAnnotation("b-string-value", stringValue)); err != nil {
				return nil, err
			}

			if err := items[i].PipeE(yaml.SetAnnotation("b-int-value", fmt.Sprintf("%v", intValue))); err != nil {
				return nil, err
			}

			if err := items[i].PipeE(yaml.SetAnnotation("b-bool-value", fmt.Sprintf("%v", boolValue))); err != nil {
				return nil, err
			}
		}
		return items, nil
	})
	cmd.Flags().StringVar(&stringValue, "string-value", "", "annotation value")
	cmd.Flags().IntVar(&intValue, "int-value", 0, "annotation value")
	cmd.Flags().BoolVar(&boolValue, "bool-value", false, "annotation value")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
