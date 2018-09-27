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
	"sigs.k8s.io/kustomize/pkg/gvk"
)

// defaultNameReferencePathConfigs is the default configuration for updating
// the fields reference the name of other resources.
var defaultNameReferencePathConfigs = []ReferencePathConfig{
	{
		referencedGVK: gvk.Gvk{
			Kind: "Deployment",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "HorizontalPodAutoscaler",
				},
				Path:               []string{"spec", "scaleTargetRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Kind: "ReplicationController",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "HorizontalPodAutoscaler",
				},
				Path:               []string{"spec", "scaleTargetRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Kind: "ReplicaSet",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "HorizontalPodAutoscaler",
				},
				Path:               []string{"spec", "scaleTargetRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Version: "v1",
			Kind:    "ConfigMap",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "projected", "sources", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "projected", "sources", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "volumes", "configMap", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "env", "valueFrom", "configMapKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "envFrom", "configMapRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Version: "v1",
			Kind:    "Secret",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Version: "v1",
					Kind:    "Pod",
				},
				Path:               []string{"spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicaSet",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "volumes", "secret", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "env", "valueFrom", "secretKeyRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "containers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "initContainers", "envFrom", "secretRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Ingress",
				},
				Path:               []string{"spec", "tls", "secretName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Ingress",
				},
				Path:               []string{"metadata", "annotations", "ingress.kubernetes.io/auth-secret"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Ingress",
				},
				Path:               []string{"metadata", "annotations", "nginx.ingress.kubernetes.io/auth-secret"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ServiceAccount",
				},
				Path:               []string{"imagePullSecrets", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StorageClass",
				},
				Path:               []string{"parameters", "secretName"}, // This is for Glusterfs,
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StorageClass",
				},
				Path:               []string{"parameters", "adminSecretName"}, // This is for Quobyte, CephRBD, StorageOS
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StorageClass",
				},
				Path:               []string{"parameters", "userSecretName"}, // This is for CephRBD
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StorageClass",
				},
				Path:               []string{"parameters", "secretRef"}, // This is for ScaleIO
				CreateIfNotPresent: false,
			},
		},
	},
	{
		// StatefulSet references headless service, so need to update the references.
		referencedGVK: gvk.Gvk{
			Version: "v1",
			Kind:    "Service",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Group: "apps",
					Kind:  "StatefulSet",
				},
				Path:               []string{"spec", "serviceName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Ingress",
				},
				Path:               []string{"spec", "rules", "http", "paths", "backend", "serviceName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Ingress",
				},
				Path:               []string{"spec", "backend", "serviceName"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Group: "rbac.authorization.k8s.io",
			Kind:  "Role",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Group: "rbac.authorization.k8s.io",
					Kind:  "RoleBinding",
				},
				Path:               []string{"roleRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Group: "rbac.authorization.k8s.io",
			Kind:  "ClusterRole",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Group: "rbac.authorization.k8s.io",
					Kind:  "RoleBinding",
				},
				Path:               []string{"roleRef", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Group: "rbac.authorization.k8s.io",
					Kind:  "ClusterRoleBinding",
				},
				Path:               []string{"roleRef", "name"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Version: "v1",
			Kind:    "ServiceAccount",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Group: "rbac.authorization.k8s.io",
					Kind:  "RoleBinding",
				},
				Path:               []string{"subjects", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Group: "rbac.authorization.k8s.io",
					Kind:  "ClusterRoleBinding",
				},
				Path:               []string{"subjects", "name"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Pod",
				},
				Path:               []string{"spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicationController",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "serviceAccountName"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Version: "v1",
			Kind:    "PersistentVolumeClaim",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Pod",
				},
				Path:               []string{"spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "StatefulSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Deployment",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "ReplicationController",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "CronJob",
				},
				Path:               []string{"spec", "jobTemplate", "spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "Job",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "DaemonSet",
				},
				Path:               []string{"spec", "template", "spec", "volumes", "persistentVolumeClaim", "claimName"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		referencedGVK: gvk.Gvk{
			Version: "v1",
			Kind:    "PersistentVolume",
		},
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &gvk.Gvk{
					Kind: "PersistentVolumeClaim",
				},
				Path:               []string{"spec", "volumeName"},
				CreateIfNotPresent: false,
			},
		},
	},
}
