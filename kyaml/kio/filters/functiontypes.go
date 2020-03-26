// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	FunctionAnnotationKey    = "config.kubernetes.io/function"
	oldFunctionAnnotationKey = "config.k8s.io/function"
)

var functionAnnotationKeys = []string{FunctionAnnotationKey, oldFunctionAnnotationKey}

// FunctionSpec defines a spec for running a function
type FunctionSpec struct {
	// Path defines the path for scoped functions
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Network is the name of the network to use from a container
	Network string `json:"network,omitempty" yaml:"network,omitempty"`

	// Container is the spec for running a function as a container
	Container ContainerSpec `json:"container,omitempty" yaml:"container,omitempty"`

	// Starlark is the spec for running a function as a starlark script
	Starlark StarlarkSpec `json:"starlark,omitempty" yaml:"starlark,omitempty"`
}

// ContainerSpec defines a spec for running a function as a container
type ContainerSpec struct {
	// Image is the container image to run
	Image string `json:"image,omitempty" yaml:"image,omitempty"`

	// Network defines network specific configuration
	Network ContainerNetwork `json:"network,omitempty" yaml:"network,omitempty"`
}

// ContainerNetwork
type ContainerNetwork struct {
	// Required specifies that function requires a network
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`
}

// StarlarkSpec defines how to run a function as a starlark program
type StarlarkSpec struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Path specifies a path to a starlark script
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// GetFunctionSpec returns the FunctionSpec for a resource.  Returns
// nil if the resource does not have a FunctionSpec.
//
// The FunctionSpec is read from the resource metadata.annotation
// "config.kubernetes.io/function"
func GetFunctionSpec(n *yaml.RNode) *FunctionSpec {
	meta, err := n.GetMeta()
	if err != nil {
		return nil
	}

	// path to the function, this will be mounted into the container
	path := meta.Annotations[kioutil.PathAnnotation]
	if fn := getFunctionSpecFromAnnotation(n, meta); fn != nil {
		fn.Network = ""
		fn.Path = path
		return fn
	}

	// legacy function specification for backwards compatibility
	container := meta.Annotations["config.kubernetes.io/container"]
	if container != "" {
		return &FunctionSpec{
			Path: path, Container: ContainerSpec{Image: container}}
	}
	return nil
}

// getFunctionSpecFromAnnotation parses the config function from an annotation
// if it is found
func getFunctionSpecFromAnnotation(n *yaml.RNode, meta yaml.ResourceMeta) *FunctionSpec {
	var fs FunctionSpec
	for _, s := range functionAnnotationKeys {
		fn := meta.Annotations[s]
		if fn != "" {
			_ = yaml.Unmarshal([]byte(fn), &fs)
			return &fs
		}
	}
	n, err := n.Pipe(yaml.Lookup("metadata", "configFn"))
	if err != nil || yaml.IsEmpty(n) {
		return nil
	}
	s, err := n.String()
	if err != nil {
		return nil
	}
	_ = yaml.Unmarshal([]byte(s), &fs)
	return &fs
}
