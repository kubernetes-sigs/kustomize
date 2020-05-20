// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runtimeutil

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	FunctionAnnotationKey    = "config.kubernetes.io/function"
	oldFunctionAnnotationKey = "config.k8s.io/function"
)

var functionAnnotationKeys = []string{FunctionAnnotationKey, oldFunctionAnnotationKey}

// FunctionSpec defines a spec for running a function
type FunctionSpec struct {
	// Network is the name of the network to use from a container
	Network string `json:"network,omitempty" yaml:"network,omitempty"`

	DeferFailure bool `json:"deferFailure,omitempty" yaml:"deferFailure,omitempty"`

	// Container is the spec for running a function as a container
	Container ContainerSpec `json:"container,omitempty" yaml:"container,omitempty"`

	// Starlark is the spec for running a function as a starlark script
	Starlark StarlarkSpec `json:"starlark,omitempty" yaml:"starlark,omitempty"`

	// ExecSpec is the spec for running a function as an executable
	Exec ExecSpec `json:"exec,omitempty" yaml:"exec,omitempty"`

	// Mounts are the storage or directories to mount into the container
	StorageMounts []StorageMount `json:"mounts,omitempty" yaml:"mounts,omitempty"`
}

type ExecSpec struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// ContainerSpec defines a spec for running a function as a container
type ContainerSpec struct {
	// Image is the container image to run
	Image string `json:"image,omitempty" yaml:"image,omitempty"`

	// Network defines network specific configuration
	Network ContainerNetwork `json:"network,omitempty" yaml:"network,omitempty"`

	// Mounts are the storage or directories to mount into the container
	StorageMounts []StorageMount `json:"mounts,omitempty" yaml:"mounts,omitempty"`
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

	// URL specifies a url containing a starlark script
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}

// StorageMount represents a container's mounted storage option(s)
type StorageMount struct {
	// Type of mount e.g. bind mount, local volume, etc.
	MountType string `json:"type,omitempty" yaml:"type,omitempty"`

	// Source for the storage to be mounted.
	// For named volumes, this is the name of the volume.
	// For anonymous volumes, this field is omitted (empty string).
	// For bind mounts, this is the path to the file or directory on the host.
	Src string `json:"src,omitempty" yaml:"src,omitempty"`

	// The path where the file or directory is mounted in the container.
	DstPath string `json:"dst,omitempty" yaml:"dst,omitempty"`
}

func (s *StorageMount) String() string {
	return fmt.Sprintf("type=%s,src=%s,dst=%s:ro", s.MountType, s.Src, s.DstPath)
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

	if fn := getFunctionSpecFromAnnotation(n, meta); fn != nil {
		fn.Network = ""
		fn.StorageMounts = []StorageMount{}
		return fn
	}

	// legacy function specification for backwards compatibility
	container := meta.Annotations["config.kubernetes.io/container"]
	if container != "" {
		return &FunctionSpec{Container: ContainerSpec{Image: container}}
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
			err := yaml.Unmarshal([]byte(fn), &fs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			return &fs
		}
	}
	n, err := n.Pipe(yaml.Lookup("metadata", "configFn"))
	if err != nil || yaml.IsEmpty(n) {
		return nil
	}
	s, err := n.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	err = yaml.Unmarshal([]byte(s), &fs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	return &fs
}

func StringToStorageMount(s string) StorageMount {
	m := make(map[string]string)
	options := strings.Split(s, ",")
	for _, option := range options {
		keyVal := strings.SplitN(option, "=", 2)
		m[keyVal[0]] = keyVal[1]
	}
	var sm StorageMount
	for key, value := range m {
		switch {
		case key == "type":
			sm.MountType = value
		case key == "src":
			sm.Src = value
		case key == "dst":
			sm.DstPath = value
		}
	}
	return sm
}

// IsReconcilerFilter filters Resources based on whether or not they are Reconciler Resource.
// Resources with an apiVersion starting with '*.gcr.io', 'gcr.io' or 'docker.io' are considered
// Reconciler Resources.
type IsReconcilerFilter struct {
	// ExcludeReconcilers if set to true, then Reconcilers will be excluded -- e.g.
	// Resources with a reconcile container through the apiVersion (gcr.io prefix) or
	// through the annotations
	ExcludeReconcilers bool `yaml:"excludeReconcilers,omitempty"`

	// IncludeNonReconcilers if set to true, the NonReconciler will be included.
	IncludeNonReconcilers bool `yaml:"includeNonReconcilers,omitempty"`
}

// Filter implements kio.Filter
func (c *IsReconcilerFilter) Filter(inputs []*yaml.RNode) ([]*yaml.RNode, error) {
	var out []*yaml.RNode
	for i := range inputs {
		isFnResource := GetFunctionSpec(inputs[i]) != nil
		if isFnResource && !c.ExcludeReconcilers {
			out = append(out, inputs[i])
		}
		if !isFnResource && c.IncludeNonReconcilers {
			out = append(out, inputs[i])
		}
	}
	return out, nil
}
