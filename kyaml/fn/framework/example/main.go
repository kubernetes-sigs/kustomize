// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	var value string
	fn := func(rl *framework.ResourceList) error {
		for i := range rl.Items {
			// set the annotation on each resource item
			if err := rl.Items[i].PipeE(yaml.SetAnnotation("value", value)); err != nil {
				return err
			}
		}
		return nil
	}
	cmd := command.Build(framework.ResourceListProcessorFunc(fn), command.StandaloneEnabled, false)
	cmd.Flags().StringVar(&value, "value", "", "annotation value")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
