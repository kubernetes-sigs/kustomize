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

// defaultNameReferencePathConfigs is the default configuration for updating
// the fields reference the name of other resources.
var defaultNameReferencePathConfigs = []referencePathConfig{
	{
		referencedGVK: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "ConfigMap",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Secret",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
}
