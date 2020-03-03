// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements adding an Application CR to a group of resources and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/application/api/v1beta1"
	yaml2 "sigs.k8s.io/yaml"

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

		Descriptor v1beta1.Descriptor `yaml:"descriptor,omitempty"`
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

func addApplicationCR(api API, groupKinds []metav1.GroupKind) (*yaml.RNode, error) {
	app := v1beta1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "app.k8s.io/v1beta1",
			Kind:       "Application",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   api.Spec.Namespace,
			Name:        api.Spec.Name,
			Labels:      map[string]string{"app.kubernetes.io/name": api.Spec.Name},
			Annotations: map[string]string{"app.kubernetes.io/managed-by": api.Spec.ManagedBy},
		},
		Spec: v1beta1.ApplicationSpec{
			ComponentGroupKinds: groupKinds,
			Descriptor:          api.Spec.Descriptor,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": api.Spec.Name},
			},
		},
	}

	data, err := yaml2.Marshal(app)
	if err != nil {
		return nil, err
	}
	return yaml.Parse(string(data))
}

func addApplicationLabel(name string, in []*yaml.RNode) error {
	for _, r := range in {
		if _, err := r.Pipe(yaml.SetLabel("app.kubernetes.io/name", name)); err != nil {
			return err
		}
	}
	return nil
}
