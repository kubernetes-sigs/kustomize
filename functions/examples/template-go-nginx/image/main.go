// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	rw := &kio.ByteReadWriter{
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		KeepReaderAnnotations: true,
	}

	err := kio.Pipeline{
		Inputs: []kio.Reader{rw},
		Filters: []kio.Filter{
			&filter{rw: rw},       // generate the Resources from the template
			filters.MergeFilter{}, // merge the generated template
			// set Resource filenames
			&filters.FileSetter{FilenamePattern: filepath.Join("config", "%n.yaml")},
			filters.FormatFilter{}, // format the output
		},
		Outputs: []kio.Writer{rw},
	}.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

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

// filter implements kio.Filter
type filter struct {
	rw *kio.ByteReadWriter
}

// Filter checks each input and ensures that all containers have cpu and memory
// reservations set, otherwise it returns an error.
func (f *filter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	api := f.parseAPI()

	// execute the service template
	buff := &bytes.Buffer{}
	t := template.Must(template.New("nginx-service").Parse(serviceTemplate))
	if err := t.Execute(buff, api); err != nil {
		return nil, err
	}
	s, err := yaml.Parse(buff.String())
	if err != nil {
		return nil, err
	}

	// execute the deployment template
	buff = &bytes.Buffer{}
	t = template.Must(template.New("nginx-deployment").Parse(deploymentTemplate))
	if err := t.Execute(buff, api); err != nil {
		return nil, err
	}
	d, err := yaml.Parse(buff.String())
	if err != nil {
		return nil, err
	}

	// add the template generated Resources to the output -- these will get merged by the next
	// filter
	in = append(in, s, d)
	return in, nil
}

// parseAPI parses the functionConfig into an API struct, and validates the input
func (f *filter) parseAPI() API {
	// parse the input function config -- TODO: simplify this
	var api API
	if err := yaml.Unmarshal([]byte(f.rw.FunctionConfig.MustString()), &api); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Default functionConfig values from environment variables if they are not set
	// in the functionConfig
	r := os.Getenv("REPLICAS")
	if r != "" && api.Spec.Replicas == nil {
		replicas, err := strconv.Atoi(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		api.Spec.Replicas = &replicas
	}
	if api.Spec.Replicas == nil {
		r := 1
		api.Spec.Replicas = &r
	}
	if api.Metadata.Name == "" {
		fmt.Fprintf(os.Stderr, "must specify metadata.name\n")
		os.Exit(1)
	}

	return api
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
