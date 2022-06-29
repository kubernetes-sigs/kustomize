// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"strconv"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
)

// define the input API schema as a struct
type API struct {
	Metadata struct {
		// Name is the Deployment Resource and Container name
		Name string `yaml:"name"`
	} `yaml:"metadata"`

	Spec struct {
		// Replicas is the number of Deployment replicas
		// Defaults to the REPLICAS env var, or 1
		Replicas *int `yaml:"replicas"`
	} `yaml:"spec"`
}

func main() {
	api := new(API)

	// create the template
	fn := framework.TemplateProcessor{
		// Templates input
		TemplateData: api,
		// Templates
		ResourceTemplates: []framework.ResourceTemplate{
			{
				Templates: parser.TemplateStrings(serviceTemplate + "---\n" + deploymentTemplate),
			},
		},
	}

	cmd := command.Build(fn, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initAPI(api *API) error {
	// Default functionConfig values from environment variables if they are not set
	// in the functionConfig
	r := os.Getenv("REPLICAS")
	if r != "" && api.Spec.Replicas == nil {
		replicas, err := strconv.Atoi(r)
		if err != nil {
			return errors.Wrap(err)
		}
		api.Spec.Replicas = &replicas
	}
	if api.Spec.Replicas == nil {
		r := 1
		api.Spec.Replicas = &r
	}
	if api.Metadata.Name == "" {
		return errors.Errorf("must specify metadata.name\n")
	}

	return nil
}

var serviceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{ .Metadata.Name }}
  labels:
    app: nginx
    instance: {{ .Metadata.Name }}
spec:
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: nginx
    instance: {{ .Metadata.Name }}
`

var deploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Metadata.Name }}
  labels:
    app: nginx
    instance: {{ .Metadata.Name }}
spec:
  replicas: {{ .Spec.Replicas }}
  selector:
    matchLabels:
      app: nginx
      instance: {{ .Metadata.Name }}
  template:
    metadata:
      labels:
        app: nginx
        instance: {{ .Metadata.Name }}
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
`
