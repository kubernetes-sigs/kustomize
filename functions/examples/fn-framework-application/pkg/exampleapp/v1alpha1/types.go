// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// +groupName=platform.example.com
// +versionName=v1alpha1
// +kubebuilder:validation:Required

package v1alpha1

import (
	_ "embed"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate ./crd_gen.sh

//go:embed platform.example.com_exampleapps.yaml
var CRDString string

const Group = "platform.example.com"
const Version = "v1alpha1"
const Kind = "ExampleApp"

//nolint:gochecknoglobals
var GroupVersion = strings.Join([]string{Group, Version}, "/")

type ExampleApp struct {
	// Embedding these structs is required to use controller-gen to produce the CRD
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// +kubebuilder:validation:Enum=production;staging;development
	Env string `json:"env" yaml:"env"`

	// +optional
	AppImage string `json:"appImage" yaml:"appImage"`

	Workloads Workloads `json:"workloads" yaml:"workloads"`

	// +optional
	Datastores Datastores `json:"datastores,omitempty" yaml:"datastores,omitempty"`

	// +optional
	Overrides Overrides `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}

// +kubebuilder:validation:MinProperties=1
type Overrides struct {
	// +optional
	AdditionalResources []string `json:"additionalResources,omitempty" yaml:"additionalResources,omitempty"`

	// +optional
	ResourcePatches []string `json:"resourcePatches,omitempty" yaml:"resourcePatches,omitempty"`

	// +optional
	ContainerPatches []string `json:"containerPatches,omitempty" yaml:"containerPatches,omitempty"`
}

// +kubebuilder:validation:MinProperties=1
type Workloads struct {
	// +optional
	JobWorkers []JobWorker `json:"jobWorkers,omitempty" yaml:"jobWorkers,omitempty"`

	// +optional
	WebWorkers []WebWorker `json:"webWorkers,omitempty" yaml:"webWorkers,omitempty"`
}

// +kubebuilder:validation:Pattern="^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?(?:\\.[a-z0-9](?:[-a-z0-9]*[a-z0-9])?)*$"
type KubernetesMetaName string

type JobWorker struct {
	Name KubernetesMetaName `json:"name" yaml:"name"`

	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int `json:"replicas,omitempty" yaml:"replicas,omitempty"`

	// +optional
	Resources ResourceBinSize `json:"resources,omitempty" yaml:"resources,omitempty"`

	// +kubebuilder:validation:UniqueItems=true
	// +kubebuilder:validation:MinItems=1
	Queues []string `json:"queues" yaml:"queues"`
}

// +kubebuilder:validation:Enum=small;medium;large
type ResourceBinSize string

const ResourceBinSizeSmall ResourceBinSize = "small"

type WebWorker struct {
	Name KubernetesMetaName `json:"name" yaml:"name"`

	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int `json:"replicas,omitempty" yaml:"replicas,omitempty"`

	// +optional
	Resources ResourceBinSize `json:"resources,omitempty" yaml:"resources,omitempty"`

	// +kubebuilder:validation:UniqueItems=true
	// +kubebuilder:validation:MinItems=1
	Domains []string `json:"domains" yaml:"domains"`
}

type Datastores struct {
	// +optional
	PostgresInstance string `json:"postgresInstance,omitempty" yaml:"postgresInstance,omitempty"`
}
