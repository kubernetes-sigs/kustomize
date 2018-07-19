/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package transformers

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// defaultLabelsPathConfigs is the default configuration for mutating labels and
// selector fields for native k8s APIs.
var defaultLabelsPathConfigs = []PathConfig{
	{
		Path:               []string{"metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Version: "v1", Kind: "Service"},
		Path:               []string{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Version: "v1", Kind: "ReplicationController"},
		Path:               []string{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Version: "v1", Kind: "ReplicationController"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "Deployment"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "Deployment"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "Deployment"},
		Path: []string{"spec", "template", "spec", "affinity", "podAffinity",
			"preferredDuringSchedulingIgnoredDuringExecution",
			"podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "Deployment"},
		Path: []string{"spec", "template", "spec", "affinity", "podAffinity",
			"requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "Deployment"},
		Path: []string{"spec", "template", "spec", "affinity", "podAntiAffinity",
			"preferredDuringSchedulingIgnoredDuringExecution",
			"podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "Deployment"},
		Path: []string{"spec", "template", "spec", "affinity", "podAntiAffinity",
			"requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "ReplicaSet"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "ReplicaSet"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "DaemonSet"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "DaemonSet"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path: []string{"spec", "template", "spec", "affinity", "podAffinity",
			"preferredDuringSchedulingIgnoredDuringExecution",
			"podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path: []string{"spec", "template", "spec", "affinity", "podAffinity",
			"requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path: []string{"spec", "template", "spec", "affinity", "podAntiAffinity",
			"preferredDuringSchedulingIgnoredDuringExecution",
			"podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path: []string{"spec", "template", "spec", "affinity", "podAntiAffinity",
			"requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "Job"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "Job"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "policy", Kind: "PodDisruptionBudget"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "NetworkPolicy"},
		Path:               []string{"spec", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "NetworkPolicy"},
		Path:               []string{"spec", "ingress", "from", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "NetworkPolicy"},
		Path:               []string{"spec", "egress", "to", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
}

// defaultLabelsPathConfigs is the default configuration for mutating annotations
// fields for native k8s APIs.
var defaultAnnotationsPathConfigs = []PathConfig{
	{
		Path:               []string{"metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Version: "v1", Kind: "ReplicationController"},
		Path:               []string{"spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "Deployment"},
		Path:               []string{"spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "ReplicaSet"},
		Path:               []string{"spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Kind: "DaemonSet"},
		Path:               []string{"spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "apps", Kind: "StatefulSet"},
		Path:               []string{"spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "Job"},
		Path:               []string{"spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "spec", "template", "metadata", "annotations"},
		CreateIfNotPresent: true,
	},
}

// AddLabelsPathConfigs adds extra path configs to the default one
func AddLabelsPathConfigs(pathConfigs ...PathConfig) {
	defaultLabelsPathConfigs = append(defaultLabelsPathConfigs, pathConfigs...)
}

// AddAnnotationsPathConfigs adds extra path configs to the default one
func AddAnnotationsPathConfigs(pathConfigs ...PathConfig) {
	defaultAnnotationsPathConfigs = append(defaultAnnotationsPathConfigs, pathConfigs...)
}
