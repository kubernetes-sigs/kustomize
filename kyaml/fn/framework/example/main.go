// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	var value string
	cmd := framework.Command(nil, func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			// set the annotation on each resource item
			if err := items[i].PipeE(yaml.SetAnnotation("value", value)); err != nil {
				return nil, err
			}
		}
		return items, nil
	})
	cmd.Flags().StringVar(&value, "value", "", "annotation value")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
