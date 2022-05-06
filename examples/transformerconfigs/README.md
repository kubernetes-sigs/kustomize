# Transformer Configurations

Kustomize creates new resources by applying a series of transformations to an original
set of resources. Kustomize provides the following default transformers:

- annotations
- images
- labels
- name reference
- namespace
- prefix/suffix
- variable reference

A `fieldSpec` list, in a transformer's configuration, determines which resource types and which fields
within those types the transformer can modify.

## FieldSpec

FieldSpec is a type that represents a path to a field in one kind of resource.

```yaml
group: some-group
version: some-version
kind: some-kind
path: path/to/the/field
create: false
```

If `create` is set to `true`, the transformer creates the path to the field in the resource if the path is not already found. This is most useful for label and annotation transformers, where the path for labels or annotations may not be set before the transformation.

## Images transformer

The default images transformer updates the specified image key values found in paths that include
`containers` and `initcontainers` sub-paths.
If found, the `image` key value is customized by the values set in the `newName`, `newTag`, and `digest` fields.
The `name` field should match the `image` key value in a resource.

Example kustomization.yaml:

```yaml
images:
  - name: postgres
    newName: my-registry/my-postgres
    newTag: v1
  - name: nginx
    newTag: 1.8.0
  - name: my-demo-app
    newName: my-app
  - name: alpine
    digest: sha256:25a0d4
```

Image transformer configurations can be customized by creating a list of `images` containing the `path` and `kind` fields.
The images transformation tutorial shows how to specify the default images transformer and customize the [images transformer configuration](images/README.md).

## Prefix/suffix transformer

The prefix/suffix transformer adds a prefix/suffix to the `metadata/name` field for all resources. Here is the default prefix transformer configuration:

```yaml
namePrefix:
- path: metadata/name

nameSuffix:
- path: metadata/name
```

Example kustomization.yaml:

```yaml

namePrefix:
  alices-

nameSuffix:
  -v2
```

## Labels transformer

The labels transformer adds labels to the `metadata/labels` field for all resources. It also adds labels to the `spec/selector` field in all Service resources as well as the `spec/selector/matchLabels` field in all Deployment resources.

Example:

```yaml
commonLabels:
- path: metadata/labels
  create: true

- path: spec/selector
  create: true
  version: v1
  kind: Service

- path: spec/selector/matchLabels
  create: true
  kind: Deployment
```

Example kustomization.yaml:

```yaml
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

## Annotations transformer

The annotations transformer adds annotations to the `metadata/annotations` field for all resources.
Annotations are also added to `spec/template/metadata/annotations` for Deployment,
ReplicaSet, DaemonSet, StatefulSet, Job, and CronJob resources, and `spec/jobTemplate/spec/template/metadata/annotations`
for CronJob resources.

Example kustomization.yaml

```yaml
commonAnnotations:
  oncallPager: 800-555-1212
```

## Name reference transformer

`nameReference` Transformer is used to tie a target resource's name to a list of other resources' referrers' names.
Once tied, the referrers' names will change alongside the target name via transformers like `namePrefix` and `nameSuffix`

### Usage
  - The syntax `nameReference` should be written in the `configurations` files, not directly in `kustomization.yaml`
  - kustomize has a set of builtin nameReference, and you don't need to write additional configs to use those nameReference. Check the full list [here](#builtin-namereference).
  - The referrer's `name` has to match the target's `name`. Otherwise, the referrer's name won't be changed. This name can be the target's current name, or, if the target has been through multiple name transformations, can be any of the target's previous names.
  - `nameReference` should be used together with other name-changing transformers. Using it alone won't make any changes.

### Example

kustomization.yaml
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  # resources.yaml contains a Gorilla and a AnimalPark
  - resources.yaml

# Add name prefix "sample-" to the Gorilla's and AnimalPark's name
namePrefix: sample-

configurations:
  # Tie target Gorilla to AnimalPark's referrer spec.gorillaRef.
  # Once Gorilla name is changed, the AnimalPark referrerd Gorilla name will be changed as well.
  - nameReference.yaml
```
nameReference.yaml
```yaml
nameReference:
  - kind: Gorilla
    fieldSpecs:
      - kind: AnimalPark
        path: spec/gorillaRef/name
```

resources.yaml
```yaml
apiVersion: animal/v1
kind: Gorilla
metadata:
  name: gg
---
apiVersion: animal/v1
kind: AnimalPark
metadata:
  name: ap
spec:
  gorillaRef:
    name: gg
    kind: Gorilla
    apiVersion: animal/v1
```
Output of `kustomize build`
```yaml
apiVersion: animal/v1
kind: AnimalPark
metadata:
  name: sample-ap # changed by `namePrefix`
spec:
  gorillaRef:
    apiVersion: animal/v1
    kind: Gorilla
    name: sample-gg # changed by `nameReference`
---
apiVersion: animal/v1
kind: Gorilla
metadata:
  name: sample-gg # changed by `namePrefix`
```

### builtin NameReference

```yaml
nameReference:
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

- kind: StatefulSet
  fieldSpecs:
    - path: spec/scaleTargetRef/name
      kind: HorizontalPodAutoscaler

- kind: ConfigMap
  version: v1
  fieldSpecs:
    - path: spec/volumes/configMap/name
      version: v1
      kind: Pod
    - path: spec/containers/env/valueFrom/configMapKeyRef/name
      version: v1
      kind: Pod
    - path: spec/initContainers/env/valueFrom/configMapKeyRef/name
      version: v1
      kind: Pod
    - path: spec/containers/envFrom/configMapRef/name
      version: v1
      kind: Pod
    - path: spec/initContainers/envFrom/configMapRef/name
      version: v1
      kind: Pod
    - path: spec/volumes/projected/sources/configMap/name
      version: v1
      kind: Pod
    - path: template/spec/volumes/configMap/name
      kind: PodTemplate
    - path: spec/template/spec/volumes/configMap/name
      kind: Deployment
    - path: spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
      kind: Deployment
    - path: spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
      kind: Deployment
    - path: spec/template/spec/containers/envFrom/configMapRef/name
      kind: Deployment
    - path: spec/template/spec/initContainers/envFrom/configMapRef/name
      kind: Deployment
    - path: spec/template/spec/volumes/projected/sources/configMap/name
      kind: Deployment
    - path: spec/template/spec/volumes/configMap/name
      kind: ReplicaSet
    - path: spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
      kind: ReplicaSet
    - path: spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
      kind: ReplicaSet
    - path: spec/template/spec/containers/envFrom/configMapRef/name
      kind: ReplicaSet
    - path: spec/template/spec/initContainers/envFrom/configMapRef/name
      kind: ReplicaSet
    - path: spec/template/spec/volumes/projected/sources/configMap/name
      kind: ReplicaSet
    - path: spec/template/spec/volumes/configMap/name
      kind: DaemonSet
    - path: spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
      kind: DaemonSet
    - path: spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
      kind: DaemonSet
    - path: spec/template/spec/containers/envFrom/configMapRef/name
      kind: DaemonSet
    - path: spec/template/spec/initContainers/envFrom/configMapRef/name
      kind: DaemonSet
    - path: spec/template/spec/volumes/projected/sources/configMap/name
      kind: DaemonSet
    - path: spec/template/spec/volumes/configMap/name
      kind: StatefulSet
    - path: spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
      kind: StatefulSet
    - path: spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
      kind: StatefulSet
    - path: spec/template/spec/containers/envFrom/configMapRef/name
      kind: StatefulSet
    - path: spec/template/spec/initContainers/envFrom/configMapRef/name
      kind: StatefulSet
    - path: spec/template/spec/volumes/projected/sources/configMap/name
      kind: StatefulSet
    - path: spec/template/spec/volumes/configMap/name
      kind: Job
    - path: spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
      kind: Job
    - path: spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
      kind: Job
    - path: spec/template/spec/containers/envFrom/configMapRef/name
      kind: Job
    - path: spec/template/spec/initContainers/envFrom/configMapRef/name
      kind: Job
    - path: spec/template/spec/volumes/projected/sources/configMap/name
      kind: Job
    - path: spec/jobTemplate/spec/template/spec/volumes/configMap/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/volumes/projected/sources/configMap/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/containers/env/valueFrom/configMapKeyRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/initContainers/env/valueFrom/configMapKeyRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/containers/envFrom/configMapRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/initContainers/envFrom/configMapRef/name
      kind: CronJob
    - path: spec/configSource/configMap
      kind: Node
    - path: rules/resourceNames
      kind: Role
    - path: rules/resourceNames
      kind: ClusterRole
    - path: metadata/annotations/nginx.ingress.kubernetes.io\/fastcgi-params-configmap
      kind: Ingress

- kind: Secret
  version: v1
  fieldSpecs:
    - path: spec/volumes/secret/secretName
      version: v1
      kind: Pod
    - path: spec/containers/env/valueFrom/secretKeyRef/name
      version: v1
      kind: Pod
    - path: spec/initContainers/env/valueFrom/secretKeyRef/name
      version: v1
      kind: Pod
    - path: spec/containers/envFrom/secretRef/name
      version: v1
      kind: Pod
    - path: spec/initContainers/envFrom/secretRef/name
      version: v1
      kind: Pod
    - path: spec/imagePullSecrets/name
      version: v1
      kind: Pod
    - path: spec/volumes/projected/sources/secret/name
      version: v1
      kind: Pod
    - path: spec/template/spec/volumes/secret/secretName
      kind: Deployment
    - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: Deployment
    - path: spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
      kind: Deployment
    - path: spec/template/spec/containers/envFrom/secretRef/name
      kind: Deployment
    - path: spec/template/spec/initContainers/envFrom/secretRef/name
      kind: Deployment
    - path: spec/template/spec/imagePullSecrets/name
      kind: Deployment
    - path: spec/template/spec/volumes/projected/sources/secret/name
      kind: Deployment
    - path: spec/template/spec/volumes/secret/secretName
      kind: ReplicaSet
    - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: ReplicaSet
    - path: spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
      kind: ReplicaSet
    - path: spec/template/spec/containers/envFrom/secretRef/name
      kind: ReplicaSet
    - path: spec/template/spec/initContainers/envFrom/secretRef/name
      kind: ReplicaSet
    - path: spec/template/spec/imagePullSecrets/name
      kind: ReplicaSet
    - path: spec/template/spec/volumes/projected/sources/secret/name
      kind: ReplicaSet
    - path: spec/template/spec/volumes/secret/secretName
      kind: DaemonSet
    - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: DaemonSet
    - path: spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
      kind: DaemonSet
    - path: spec/template/spec/containers/envFrom/secretRef/name
      kind: DaemonSet
    - path: spec/template/spec/initContainers/envFrom/secretRef/name
      kind: DaemonSet
    - path: spec/template/spec/imagePullSecrets/name
      kind: DaemonSet
    - path: spec/template/spec/volumes/projected/sources/secret/name
      kind: DaemonSet
    - path: spec/template/spec/volumes/secret/secretName
      kind: StatefulSet
    - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: StatefulSet
    - path: spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
      kind: StatefulSet
    - path: spec/template/spec/containers/envFrom/secretRef/name
      kind: StatefulSet
    - path: spec/template/spec/initContainers/envFrom/secretRef/name
      kind: StatefulSet
    - path: spec/template/spec/imagePullSecrets/name
      kind: StatefulSet
    - path: spec/template/spec/volumes/projected/sources/secret/name
      kind: StatefulSet
    - path: spec/template/spec/volumes/secret/secretName
      kind: Job
    - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: Job
    - path: spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
      kind: Job
    - path: spec/template/spec/containers/envFrom/secretRef/name
      kind: Job
    - path: spec/template/spec/initContainers/envFrom/secretRef/name
      kind: Job
    - path: spec/template/spec/imagePullSecrets/name
      kind: Job
    - path: spec/template/spec/volumes/projected/sources/secret/name
      kind: Job
    - path: spec/jobTemplate/spec/template/spec/volumes/secret/secretName
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/volumes/projected/sources/secret/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/initContainers/env/valueFrom/secretKeyRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/containers/envFrom/secretRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/initContainers/envFrom/secretRef/name
      kind: CronJob
    - path: spec/jobTemplate/spec/template/spec/imagePullSecrets/name
      kind: CronJob
    - path: spec/tls/secretName
      kind: Ingress
    - path: metadata/annotations/ingress.kubernetes.io\/auth-secret
      kind: Ingress
    - path: metadata/annotations/nginx.ingress.kubernetes.io\/auth-secret
      kind: Ingress
    - path: metadata/annotations/nginx.ingress.kubernetes.io\/auth-tls-secret
      kind: Ingress
    - path: spec/tls/secretName
      kind: Ingress
    - path: imagePullSecrets/name
      kind: ServiceAccount
    - path: parameters/secretName
      kind: StorageClass
    - path: parameters/adminSecretName
      kind: StorageClass
    - path: parameters/userSecretName
      kind: StorageClass
    - path: parameters/secretRef
      kind: StorageClass
    - path: rules/resourceNames
      kind: Role
    - path: rules/resourceNames
      kind: ClusterRole
    - path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
      kind: Service
      group: serving.knative.dev
      version: v1
    - path: spec/azureFile/secretName
      kind: PersistentVolume

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
    - path: spec/rules/http/paths/backend/service/name
      kind: Ingress
    - path: spec/defaultBackend/service/name
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

- kind: PriorityClass
  version: v1
  group: scheduling.k8s.io
  fieldSpecs:
    - path: spec/priorityClassName
      kind: Pod
    - path: spec/template/spec/priorityClassName
      kind: StatefulSet
    - path: spec/template/spec/priorityClassName
      kind: Deployment
    - path: spec/template/spec/priorityClassName
      kind: ReplicationController
    - path: spec/jobTemplate/spec/template/spec/priorityClassName
      kind: CronJob
    - path: spec/template/spec/priorityClassName
      kind: Job
    - path: spec/template/spec/priorityClassName
      kind: DaemonSet

- kind: IngressClass
  version: v1
  group: networking.k8s.io/v1
  fieldSpecs:
    - path: spec/ingressClassName
      kind: Ingress
```

## Customizing transformer configurations

In addition to the default transformers, you can create custom transformer configurations.

This tutorial shows how to create custom transformer configurations:

- [support a CRD type](crd/README.md)
- add extra fields for variable substitution
- add extra fields for name reference


## Supporting escape characters in CRD path

```yaml
metadata:
  annotations:
    foo.k8s.io/bar: baz
```
Kustomize supports escaping special characters in path, e.g `metadata/annotations/foo.k8s.io\/bar`
