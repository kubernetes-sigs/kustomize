// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type App struct {
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Port int `yaml:"port" json:"port"`
	} `yaml:"spec" json:"spec"`
}

func generateService(name string, sourcePort int, targetPort int) (*yaml.RNode, error) {
	serviceName := name + "-svc"
	svc, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  labels:
    app: %s
  name: %s
spec:
  selector:
    app: %s
  ports:
  - name: http
    port: %d
    protocol: TCP
    targetPort: %d
`, name, serviceName, name, sourcePort, targetPort))
	if err != nil {
		return nil, fmt.Errorf("failed to generate resource: %w", err)
	}
	return svc, nil
}

func main() {
	config := new(App)
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		var newNodes []*yaml.RNode
		resourceName := config.Metadata.Name
		for range items {
			// generate Service
			service, err := generateService(resourceName, config.Spec.Port, config.Spec.Port)
			if err != nil {
				return nil, err
			}
			newNodes = append(newNodes, service)
		}
		items = newNodes

		return items, nil
	}
	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
