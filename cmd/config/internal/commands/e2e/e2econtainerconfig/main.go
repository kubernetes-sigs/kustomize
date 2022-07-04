// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	config := &corev1.ConfigMap{}

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		var newNodes []*yaml.RNode
		for i := range items {
			// set the annotation on each resource item
			if err := items[i].PipeE(yaml.SetAnnotation("a-string-value", config.Data["stringValue"])); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
			intValue, _ := strconv.Atoi(config.Data["intValue"])
			if err := items[i].PipeE(yaml.SetAnnotation("a-int-value", strconv.Itoa(intValue))); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
			boolValue, _ := strconv.ParseBool(config.Data["boolValue"])
			if err := items[i].PipeE(yaml.SetAnnotation("a-bool-value", strconv.FormatBool(boolValue))); err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			newNodes = append(newNodes, items[i])
		}
		return newNodes, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
