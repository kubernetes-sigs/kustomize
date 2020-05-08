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
	resourceList := framework.ResourceList{}
	cmd := framework.Command(&resourceList, func() error {
		for i := range resourceList.Items {
			// set the annotation on each resource item
			err := resourceList.Items[i].PipeE(yaml.SetAnnotation("value", value))
			if err != nil {
				return err
			}
		}
		return nil
	})
	cmd.Flags().StringVar(&value, "value", "", "annotation value")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
