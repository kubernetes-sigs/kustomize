// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldreference

import (
	"regexp"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/fieldspec"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ReferenceNameAnnotation = "kustomize.io/reference-name/"
	OriginalNameAnnotation  = "kustomize.io/original-name/"
	generateNamePattern     = `^(.+)-kust[a-z0-9]{6,16}$`
)

var namePatternRegex *regexp.Regexp

func GetNamePatternRegex() *regexp.Regexp {
	if namePatternRegex == nil {
		namePatternRegex = regexp.MustCompile(generateNamePattern)
	}
	return namePatternRegex
}

var nameFieldReferenceList *FieldReferenceList

func getNameReferenceFieldSpecs() *FieldReferenceList {
	if nameFieldReferenceList != nil {
		return nameFieldReferenceList
	}
	nameFieldReferenceList = &FieldReferenceList{}
	err := yaml.Unmarshal([]byte(nameReferenceFieldSpecsData), &nameFieldReferenceList)
	if err != nil {
		panic(err)
	}
	return nameFieldReferenceList
}

type FieldReference struct {
	fieldspec.Gvk `yaml:",inline,omitempty"`
	FieldSpecs    []fieldspec.FieldSpec `yaml:"fieldSpecs,omitempty"`
}

type FieldReferenceList struct {
	Items []FieldReference `yaml:"items,omitempty"`
}

type NameKey struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
	Name       string `yaml:"name,omitempty"`
}

type NameIndex struct {
	Index map[NameKey]string
}

const nameReferenceFieldSpecsData = `
items:
- kind: Deployment
  fieldSpecs:
  - path: spec/scaleTargetRef/name
    kind: HorizontalPodAutoscaler

- kind: ReplicationController
  fieldSpecs:
  - path: spec/scaleTargetRef/name
    kind: HorizontalPodAutoscaler

- kind: ReplicaSet
  fieldSpecs:
  - path: spec/scaleTargetRef/name
    kind: HorizontalPodAutoscaler

- kind: ConfigMap
  version: v1
  fieldSpecs:
  - path: spec/volumes/configMap/name # tested
  - path: spec/containers/env/valueFrom/configMapKeyRef/name
  - path: spec/initContainers/env/valueFrom/configMapKeyRef/name
  - path: spec/containers/envFrom/configMapRef/name
  - path: spec/initContainers/envFrom/configMapRef/name
  - path: spec/volumes/projected/sources/configMap/name
  - path: spec/template/spec/volumes/configMap/name # tested
  - path: spec/template/spec/containers/env/valueFrom/configMapKeyRef/name # tested
  - path: spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
  - path: spec/template/spec/containers/envFrom/configMapRef/name
  - path: spec/template/spec/initContainers/envFrom/configMapRef/name
  - path: spec/template/spec/volumes/projected/sources/configMap/name
  - path: spec/jobTemplate/spec/template/spec/volumes/configMap/name
  - path: spec/jobTemplate/spec/template/spec/volumes/projected/sources/configMap/name
  - path: spec/jobTemplate/spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
  - path: spec/jobTemplate/spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
  - path: spec/jobTemplate/spec/template/spec/containers/envFrom/configMapRef/name
  - path: spec/jobTemplate/spec/template/spec/initContainers/envFrom/configMapRef/name

- kind: Secret
  version: v1
  fieldSpecs:
  - path: spec/volumes/secret/secretName
  - path: spec/containers/env/valueFrom/secretKeyRef/name
  - path: spec/initContainers/env/valueFrom/secretKeyRef/name
  - path: spec/containers/envFrom/secretRef/name
  - path: spec/initContainers/envFrom/secretRef/name
  - path: spec/imagePullSecrets/name
  - path: spec/volumes/projected/sources/secret/name
  - path: spec/template/spec/volumes/secret/secretName
  - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
  - path: spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
  - path: spec/template/spec/containers/envFrom/secretRef/name
  - path: spec/template/spec/initContainers/envFrom/secretRef/name
  - path: spec/template/spec/imagePullSecrets/name
  - path: spec/template/spec/volumes/projected/sources/secret/name
  - path: spec/jobTemplate/spec/template/spec/volumes/secret/secretName
  - path: spec/jobTemplate/spec/template/spec/volumes/projected/sources/secret/name
  - path: spec/jobTemplate/spec/template/spec/containers/env/valueFrom/secretKeyRef/name
  - path: spec/jobTemplate/spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
  - path: spec/jobTemplate/spec/template/spec/containers/envFrom/secretRef/name
  - path: spec/jobTemplate/spec/template/spec/initContainers/envFrom/secretRef/name
  - path: spec/jobTemplate/spec/template/spec/imagePullSecrets/name

  - path: spec/tls/secretName
  - path: metadata/annotations/ingress.kubernetes.io\/auth-secret
  - path: metadata/annotations/nginx.ingress.kubernetes.io\/auth-secret
  - path: metadata/annotations/nginx.ingress.kubernetes.io\/auth-tls-secret
  - path: imagePullSecrets/name
  - path: parameters/secretName
  - path: parameters/adminSecretName
  - path: parameters/userSecretName
  - path: parameters/secretRef
  - path: rules/resourceNames
  - path: rules/resourceNames

- kind: Service
  version: v1
  fieldSpecs:
  - path: spec/serviceName
    kind: StatefulSet
    group: apps
  - path: spec/rules/http/paths/backend/serviceName
    kind: Ingress
  - path: spec/backend/serviceName
    kind: Ingress
  - path: spec/service/name
    kind: APIService
    group: apiregistration.k8s.io
  - path: webhooks/clientConfig/service
    kind: ValidatingWebhookConfiguration
    group: admissionregistration.k8s.io
  - path: webhooks/clientConfig/service
    kind: MutatingWebhookConfiguration
    group: admissionregistration.k8s.io

- kind: Role
  group: rbac.authorization.k8s.io
  fieldSpecs:
  - path: roleRef/name
    kind: RoleBinding
    group: rbac.authorization.k8s.io

- kind: ClusterRole
  group: rbac.authorization.k8s.io
  fieldSpecs:
  - path: roleRef/name
    kind: RoleBinding
    group: rbac.authorization.k8s.io
  - path: roleRef/name
    kind: ClusterRoleBinding
    group: rbac.authorization.k8s.io

- kind: ServiceAccount
  version: v1
  fieldSpecs:
  - path: subjects
    kind: RoleBinding
    group: rbac.authorization.k8s.io
  - path: subjects
    kind: ClusterRoleBinding
    group: rbac.authorization.k8s.io
  - path: spec/serviceAccountName
    kind: Pod
  - path: spec/template/spec/serviceAccountName
    kind: StatefulSet
  - path: spec/template/spec/serviceAccountName
    kind: Deployment
  - path: spec/template/spec/serviceAccountName
    kind: ReplicationController
  - path: spec/jobTemplate/spec/template/spec/serviceAccountName
    kind: CronJob
  - path: spec/template/spec/serviceAccountName
    kind: Job
  - path: spec/template/spec/serviceAccountName
    kind: DaemonSet

- kind: PersistentVolumeClaim
  version: v1
  fieldSpecs:
  - path: spec/volumes/persistentVolumeClaim/claimName
    kind: Pod
  - path: spec/template/spec/volumes/persistentVolumeClaim/claimName
    kind: StatefulSet
  - path: spec/template/spec/volumes/persistentVolumeClaim/claimName
    kind: Deployment
  - path: spec/template/spec/volumes/persistentVolumeClaim/claimName
    kind: ReplicationController
  - path: spec/jobTemplate/spec/template/spec/volumes/persistentVolumeClaim/claimName
    kind: CronJob
  - path: spec/template/spec/volumes/persistentVolumeClaim/claimName
    kind: Job
  - path: spec/template/spec/volumes/persistentVolumeClaim/claimName
    kind: DaemonSet

- kind: PersistentVolume
  version: v1
  fieldSpecs:
  - path: spec/volumeName
    kind: PersistentVolumeClaim
  - path: rules/resourceNames
    kind: ClusterRole

- kind: StorageClass
  version: v1
  group: storage.k8s.io
  fieldSpecs:
  - path: spec/storageClassName
    kind: PersistentVolume
  - path: spec/storageClassName
    kind: PersistentVolumeClaim
  - path: spec/volumeClaimTemplates/spec/storageClassName
    kind: StatefulSet
`
