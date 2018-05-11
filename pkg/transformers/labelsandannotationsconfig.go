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
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "Job"},
		Path:               []string{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "Job"},
		Path:               []string{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "spec", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	{
		GroupVersionKind:   &schema.GroupVersionKind{Group: "batch", Kind: "CronJob"},
		Path:               []string{"spec", "jobTemplate", "spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
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
