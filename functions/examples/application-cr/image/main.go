// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements adding an Application CR to a group of resources and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func appCR() error {
	rw := &kio.ByteReadWriter{
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		OmitReaderAnnotations: true,
		KeepReaderAnnotations: true,
	}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw}, // read the inputs into a slice
		Filters: []kio.Filter{appCRFilter{rw: rw}},
		Outputs: []kio.Writer{rw}, // copy the inputs to the output
	}
	if err := p.Execute(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := appCR(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// appCRFilter implements kio.Filter
type appCRFilter struct {
	rw *kio.ByteReadWriter
}

// define the input API schema as a struct
type API struct {
	Spec struct {
		ManagedBy string `yaml:"managedBy"`

		Name string `yaml:"name"`

		Namespace string `yaml:"namespace"`
	} `yaml:"spec"`
}

// Filter checks each resource for validity, otherwise returning an error.
func (f appCRFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	api := f.parseAPI()

	groupKinds, err := getGroupKinds(in)
	if err != nil {
		return nil, err
	}

	app, err := addApplicationCR(api, groupKinds)
	if err != nil {
		return nil, err
	}
	err = addApplicationLabel(api.Spec.Name, in)
	if err != nil {
		return nil, err
	}
	if app != nil {
		return append(in, app), nil
	}
	return in, nil

}

// parseAPI parses the functionConfig into an API struct.
func (f *appCRFilter) parseAPI() API {
	// parse the input function config -- TODO: simplify this
	var api API
	if err := yaml.Unmarshal([]byte(f.rw.FunctionConfig.MustString()), &api); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	return api
}

func getGroupKinds(in []*yaml.RNode) ([]metav1.GroupKind, error) {
	var groupKinds []metav1.GroupKind
	for _, r := range in {
		meta, err := r.GetMeta()
		if err != nil {
			return nil, err
		}
		gvk := schema.FromAPIVersionAndKind(meta.APIVersion, meta.Kind)

		found := false
		for _, gk := range groupKinds {
			if gk.Group == gvk.Group && gk.Kind == gvk.Kind {
				break
			}
		}
		if !found {
			groupKinds = append(groupKinds, metav1.GroupKind{
				Group: gvk.Group,
				Kind:  gvk.Kind,
			})
		}
	}
	return groupKinds, nil
}

var applicationTemplate = `apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app.kubernetes.io/name: {{.Name}}
  annotations:
    app.kubernetes.io/managed-by: {{.ManagedBy}}
spec:
  selector:
    app.kubernetes.io/name: {{.Name}}
{{if .ComponentKinds }}
  componentKinds:
{{range $kind := .ComponentKinds }}
  - group: {{$kind.Group}}
    kind: {{$kind.Kind}}
{{end}}
{{end}}

`

func addApplicationCR(api API, groupKinds []metav1.GroupKind) (*yaml.RNode, error) {
	data := struct {
		Name           string
		Namespace      string
		ManagedBy      string
		ComponentKinds []metav1.GroupKind
	}{
		Name:           api.Spec.Name,
		Namespace:      api.Spec.Namespace,
		ManagedBy:      api.Spec.ManagedBy,
		ComponentKinds: groupKinds,
	}
	// execute the deployment template
	buff := &bytes.Buffer{}
	t := template.Must(template.New("application").Parse(applicationTemplate))
	if err := t.Execute(buff, data); err != nil {
		return nil, err
	}
	return yaml.Parse(buff.String())
}

func addApplicationLabel(name string, in []*yaml.RNode) error {
	for _, r := range in {
		if _, err := r.Pipe(yaml.SetLabel("app.kubernetes.io/name", name)); err != nil {
			return err
		}
	}
	return nil
}
