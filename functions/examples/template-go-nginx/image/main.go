// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"strconv"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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
	functionConfig := &API{}
	resourceList := &framework.ResourceList{FunctionConfig: functionConfig}

	cmd := framework.Command(resourceList, func() error {
		// initialize API defaults
		if err := initAPI(functionConfig); err != nil {
			return err
		}

		// execute the service template
		buff := &bytes.Buffer{}
		t := template.Must(template.New("nginx-service").Parse(serviceTemplate))
		if err := t.Execute(buff, functionConfig); err != nil {
			return err
		}
		s, err := yaml.Parse(buff.String())
		if err != nil {
			return err
		}

		// execute the deployment template
		buff = &bytes.Buffer{}
		t = template.Must(template.New("nginx-deployment").Parse(deploymentTemplate))
		if err := t.Execute(buff, functionConfig); err != nil {
			return err
		}
		d, err := yaml.Parse(buff.String())
		if err != nil {
			return err
		}

		// add the template generated Resources to the output -- these will get merged by the next
		// filter
		resourceList.Items = append(resourceList.Items, s, d)

		// merge the new copies with the old copies of each resource
		resourceList.Items, err = filters.MergeFilter{}.Filter(resourceList.Items)
		if err != nil {
			return err
		}

		// apply formatting
		resourceList.Items, err = filters.FormatFilter{}.Filter(resourceList.Items)
		if err != nil {
			return err
		}

		return nil
	})
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
