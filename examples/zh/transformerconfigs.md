# 自定义 transformer 配置

Kustomize 通过对原始资源集进行一系列转换来创建新资源。Kustomize 提供以下默认的 transformers ：

- annotations
- images
- labels
- name reference
- namespace
- prefix/suffix
- variable reference

transformer 配置的 `fieldSpec` 列表，用来确定哪些资源类型和这些类型的 transformer 可以修改哪些字段。

## FieldSpec

FieldSpec 是一种表示资源中字段路径的类型。

```yaml
group: some-group
version: some-version
kind: some-kind
path: path/to/the/field
create: false
```

如果 `create` 设置为 `true`，表示如果尚未找到该路径，则 transformer 将在资源中创建该路径。这对于 label 和 annotation 转换器最有用，因为在转换之前可能未设置 label 或 annotation 的路径。

## Images transformer

默认的 images transformer 会更新包含 `containers` 和 `initcontainers` 子路径的路径中找到的指定镜像的键值 。如果找到，则更新 `image` 的 `newName`，`newTag` 和 `digest` 等字段。该 `name` 字段应与 `image` 资源中的键值匹配。

kustomization.yaml 示例：

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
可以通过创建 `images` 包含 `path` 和 `kind` 字段的列表来自定义镜像 transformer 配置。[镜像 transformer 教程](image.md) 展示了如何指定默认镜像 transformer 和自定义镜像 transformer 配置。

## Prefix/suffix transformer

prefix/suffix transformer 为所有资源的 `metadata/name` 字段添加前缀/后缀。默认的 prefix transformer 配置如下：

```yaml
namePrefix:
- path: metadata/name
```

kustomization.yaml 示例：

```yaml

namePrefix:
  alices-

nameSuffix:
  -v2
```

## Labels transformer

labels transformer 将 labels 添加到所有资源的 `metadata/labels` 字段。它还将 labels 添加到 `spec/selector` 和 `spec/selector/matchLabels` 字段以及所有 Deployment 资源中的字段。

示例：

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

kustomization.yaml 示例:

```yaml
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

## Annotations transformer

annotations transformer 可以：

- 将 annotations 添加到所有资源的 `metadata/annotations` 字段
- 将 annotations 添加到Deployment，ReplicaSet，DaemonSet，StatefulSet，Job 和 CronJob 等资源的 `spec/template/metadata/annotations`
- 将 annotations 添加到CronJob 资源的 `spec/jobTemplate/spec/template/metadata/annotations`字段。

kustomization.yaml 示例：

```yaml
commonAnnotations:
  oncallPager: 800-555-1212
```

## Name reference transformer

Name reference transformer 的配置不同于其他所有的 transformer。`nameReferences` 列表代表一种可以用作其他类型资源中的引用的所有可能字段。一个 `nameReference` 包含一个类型如 ConfigMap 以及 `fieldSpecs` 列表，其中 `ConfigMap` 其他资源被引用。下面是一个例子：

```yaml
kind: ConfigMap
version: v1
fieldSpecs:
- kind: Pod
  version: v1
  path: spec/volumes/configMap/name
- kind: Deployment
  path: spec/template/spec/volumes/configMap/name
- kind: Job
  path: spec/template/spec/volumes/configMap/name
```

Name reference transformer 的配置为 `nameReferences` 列表包含 ConfigMap，Secret，Service，Role和ServiceAccount等资源。下面是一个示例配置：

```yaml
nameReference:
- kind: ConfigMap
  version: v1
  fieldSpecs:
  - path: spec/volumes/configMap/name
    version: v1
    kind: Pod
  - path: spec/containers/env/valueFrom/configMapKeyRef/name
    version: v1
    kind: Pod
  # ...
- kind: Secret
  version: v1
  fieldSpecs:
  - path: spec/volumes/secret/secretName
    version: v1
    kind: Pod
  - path: spec/containers/env/valueFrom/secretKeyRef/name
    version: v1
    kind: Pod
```

## Customizing transformer configurations

除默认 transformers 外，您还可以创建自定义的 transformers 配置。通过调用将默认的 transformers 配置保存到本地目录`kustomize config save -d`，然后修改和使用这些配置。本教程显示了如何创建自定义 transformers 配置：

- [support a CRD type](../transformerconfigs/crd/README.md)
- 添加额外的字段以进行变量替换
- 添加额外的字段以供名称参考

## Supporting escape characters in CRD path

```yaml
metadata:
  annotations:
    foo.k8s.io/bar: baz
```
Kustomize 支持在路径中转义特殊字符，例如： `metadata/annotations/foo.k8s.io\/bar`
