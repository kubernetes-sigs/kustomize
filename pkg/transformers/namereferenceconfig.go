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
var defaultNameReferencePathConfigs = []ReferencePathConfig{
	{
		referencedGVK: schema.GroupVersionKind{
			Kind: "Deployment",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "HorizontalPodAutoscaler",
				},
				Path:               []string{"spec", "scaleTargetRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Kind: "ReplicationController",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "HorizontalPodAutoscaler",
				},
				Path:               []string{"spec", "scaleTargetRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Kind: "ReplicaSet",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "HorizontalPodAutoscaler",
				},
				Path:               []string{"spec", "scaleTargetRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
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
				Path:               []string{"spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
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
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "initContainers", "envFrom", "configMapRef", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
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
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "projected", "sources", "configMap", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
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
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
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
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
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
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "projected", "sources", "configMap", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
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
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
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
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
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
				Path:               []string{"spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
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
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "imagePullSecrets", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
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
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
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
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
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
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
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
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
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
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
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
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
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
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Ingress",
				},
				Path:               []string{"spec", "tls", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Ingress",
				},
				Path:               []string{"metadata", "annotations", "ingress.kubernetes.io/auth-secret"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Ingress",
				},
				Path:               []string{"metadata", "annotations", "nginx.ingress.kubernetes.io/auth-secret"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ServiceAccount",
				},
				Path:               []string{"imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		// StatefulSet references headless service, so need to update the references.
		referencedGVK: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Service",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Group: "apps",
					Kind:  "StatefulSet",
				},
				Path:               []string{"spec", "serviceName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Ingress",
				},
				Path:               []string{"spec", "rules", "http", "paths", "backend", "serviceName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Ingress",
				},
				Path:               []string{"spec", "backend", "serviceName"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Group: "rbac.authorization.k8s.io",
			Kind:  "Role",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Group: "rbac.authorization.k8s.io",
					Kind:  "RoleBinding",
				},
				Path:               []string{"roleRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Group: "rbac.authorization.k8s.io",
			Kind:  "ClusterRole",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Group: "rbac.authorization.k8s.io",
					Kind:  "RoleBinding",
				},
				Path:               []string{"roleRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Group: "rbac.authorization.k8s.io",
					Kind:  "ClusterRoleBinding",
				},
				Path:               []string{"roleRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "ServiceAccount",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Group: "rbac.authorization.k8s.io",
					Kind:  "RoleBinding",
				},
				Path:               []string{"subjects", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Group: "rbac.authorization.k8s.io",
					Kind:  "ClusterRoleBinding",
				},
				Path:               []string{"subjects", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Pod",
				},
				Path:               []string{"spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicationController",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "PersistentVolumeClaim",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Pod",
				},
				Path:               []string{"spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "ReplicationController",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "PersistentVolume",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{
					Kind: "PersistentVolumeClaim",
				},
				Path:               []string{"spec", "volumeName"},
				CreateIfNotPresent: false,
			},
		},
	},
}

// AddNameReferencePathConfigs adds extra reference path configs to the default one
func AddNameReferencePathConfigs(r []ReferencePathConfig) {
	for _, p := range r {
		defaultNameReferencePathConfigs = MergeNameReferencePathConfigs(defaultNameReferencePathConfigs, p)
	}
}

// MergeNameReferencePathConfigs merges one ReferencePathConfig into a slice of ReferencePathConfig
func MergeNameReferencePathConfigs(configs []ReferencePathConfig, config ReferencePathConfig) []ReferencePathConfig {
	var result []ReferencePathConfig
	found := false
	for _, c := range configs {
		if c.referencedGVK == config.referencedGVK {
			c.pathConfigs = append(c.pathConfigs, config.pathConfigs...)
			found = true
		}
		result = append(result, c)
	}

	if !found {
		result = append(result, config)
	}
	return result
}
