---
title: "内置插件"
linkTitle: "内置插件"
type: docs
weight: 1
description: >
    内置插件
---

内置插件包括生成器和转化器。

每个插件都可以通过如下两种方式触发：

* 通过 kustomization 文件的字段隐式触发插件，例如 `AnnotationsTransformer` 就是由 `commonAnnotations` 字段触发的。

* 通过 `generators` 或 `transformers` 字段显式触发插件（通过指定插件的配置文件）。

直接使用 `kustomization.yaml` 文件中的字段比如 `commonLables`、`commonAnnotation` 他们可以修改一些默认的字段，如果用户想要添加或减少能被 `commonLabel` 所修改的字段，则不是很容易做到；而使用 `transformers` 的话，用户可以指定哪些字段能被修改，而不受默认值的影响。

[types.GeneratorOptions]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/generatoroptions.go
[types.SecretArgs]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/secretargs.go
[types.ConfigMapArgs]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/configmapargs.go
[config.FieldSpec]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/fieldspec.go
[types.ObjectMeta]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/objectmeta.go
[types.Selector]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/selector.go
[types.Replica]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/replica.go
[types.PatchStrategicMerge]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/patchstrategicmerge.go
[types.PatchTarget]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/patchtarget.go
[image.Image]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/types/image.go

## _AnnotationTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`commonAnnotations`

为所有资源添加注释，和标签一样以 key: value 的形式。

```yaml
commonAnnotations:
  oncallPager: 800-555-1212
```

### 使用插件

#### Arguments

> Annotations map\[string\]string
>
> FieldSpecs  \[\][config.FieldSpec]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: AnnotationsTransformer
> metadata:
>   name: not-important-to-example
> annotations:
>   app: myApp
>   greeting/morning: a string with blanks
> fieldSpecs:
> - path: metadata/annotations
>   create: true
> ```

## _ConfigMapGenerator_

### 使用 `kustomization.yaml`

#### 字段名称：`configMapGenerator`

列表中的每个条目都将生成一个 ConfigMap （合计可以生成 n 个 ConfigMap）。

下面的示例会生成 3 个ConfigMap：第一个带有给定文件的名称和内容，第二个将在 data 中添加 key/value，第三个通过 `options` 为单个 ConfigMap 设置注释和标签。

每个 configMapGenerator 项均接受的参数 `behavior: [create|replace|merge]`，这个参数允许修改或替换父级现有的 configMap。

此外，每个条目都有一个 `options` 字段，该字段具有与 kustomization 文件的 `generatorOptions` 字段相同的子字段。

`options` 字段允许用户为生成的实例添加标签和（或）注释，或者分别禁用该实例名称的哈希后缀。此处添加的标签和注释不会被 kustomization 文件 `generatorOptions` 字段关联的全局选项覆盖。但是如果全局 `generatorOptions` 字段指定 `disableNameSuffixHash: true`，其他 `options` 的设置将无法将其覆盖。

```yaml
# These labels are added to all configmaps and secrets.
generatorOptions:
  labels:
    fruit: apple

configMapGenerator:
- name: my-java-server-props
  behavior: merge
  files:
  - application.properties
  - more.properties
- name: my-java-server-env-vars
  literals: 
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
  options:
    disableNameSuffixHash: true
    labels:
      pet: dog
- name: dashboards
  files:
  - mydashboard.json
  options:
    annotations:
      dashboard: "1"
    labels:
      app.kubernetes.io/name: "app1"
```

这里也可以[定义一个 key](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#define-the-key-to-use-when-creating-a-configmap-from-a-file) 来为文件设置不同名称。

下面这个示例会创建一个 ConfigMap，并将 `whatever.ini` 重命名为 `myFileName.ini`：

```yaml
configMapGenerator:
- name: app-whatever
  files:
  - myFileName.ini=whatever.ini
```

### 使用插件

#### Arguments

> [types.ConfigMapArgs]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: ConfigMapGenerator
> metadata:
>   name: mymap
> envs:
> - devops.env
> - uxteam.env
> literals:
> - FRUIT=apple
> - VEGETABLE=carrot
> ```

## _ImageTagTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`images`

修改镜像的名称、tag 或 image digest ，而无需使用 patches 。例如，对于这种 kubernetes Deployment 片段：

```yaml
containers:
- name: mypostgresdb
  image: postgres:8
- name: nginxapp
  image: nginx:1.7.9
- name: myapp
  image: my-demo-app:latest
- name: alpine-app
  image: alpine:3.7
```

想要将 `image` 做如下更改：

- 将 `postgres:8` 改为 `my-registry/my-postgres:v1`
- 将 nginx tag 从 `1.7.9` 改为 `1.8.0`
- 将镜像名称 `my-demo-app` 改为 `my-app`
- 将 alpine 的 tag `3.7` 改为 digest 值

只需在 *kustomization* 中添加以下内容：

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
  digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
```

### 使用插件

#### Arguments

> ImageTag   [image.Image]
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: ImageTagTransformer
> metadata:
>   name: not-important-to-example
> imageTag:
>   name: nginx
>   newTag: v2
> ```

## _LabelTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`commonLabels`

为所有资源和 selectors 增加标签。

```yaml
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

### 使用插件

#### Arguments

> Labels  map\[string\]string
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: LabelTransformer
> metadata:
>   name: not-important-to-example
> labels:
>   app: myApp
>   env: production
> fieldSpecs:
> - path: metadata/labels
>   create: true
> ```

## _NamespaceTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`namespace`

为所有资源添加 namespace。

```yaml
namespace: my-namespace
```

### 使用插件

#### Arguments

> [types.ObjectMeta]
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```yaml
> apiVersion: builtin
>  kind: NamespaceTransformer
>  metadata:
>    name: not-important-to-example
>    namespace: test
>  fieldSpecs:
>  - path: metadata/namespace
>    create: true
>  - path: subjects
>    kind: RoleBinding
>    group: rbac.authorization.k8s.io
>  - path: subjects
>    kind: ClusterRoleBinding
>    group: rbac.authorization.k8s.io
> ```

## _PatchesJson6902_

### 使用 `kustomization.yaml`

#### 字段名称：`patchesJson6902`

patchesJson6902 列表中的每个条目都应可以解析为 kubernetes 对象和将应用于该对象的 [JSON patch](https://tools.ietf.org/html/rfc6902)。

目标字段指向的 kubernetes 对象的 group、 version、 kind、 name 和 namespace 在同一 kustomization 内 path 字段内容是 JSON patch 文件的相对路径。

patch 文件中的内容可以如下这种 JSON 格式：

```json
 [
   {"op": "add", "path": "/some/new/path", "value": "value"},
   {"op": "replace", "path": "/some/existing/path", "value": "new value"}
 ]
 ```

也可以使用 YAML 格式表示：

```yaml
- op: add
  path: /some/new/path
  value: value
- op: replace
  path: /some/existing/path
  value: new value
```

```yaml
patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  path: add_init_container.yaml
- target:
    version: v1
    kind: Service
    name: my-service
  path: add_service_annotation.yaml
```

patch 内容也可以是一个inline string：

```yaml
patchesJson6902:
- target:
    version: v1
    kind: Deployment
    name: my-deployment
  patch: |-
    - op: add
      path: /some/new/path
      value: value
    - op: replace
      path: /some/existing/path
      value: "new value"
```

### 使用插件

#### Arguments

> Target [types.PatchTarget]
>
> Path   string
>
> JsonOp string

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: PatchJson6902Transformer
> metadata:
>   name: not-important-to-example
> target:
>   group: apps
>   version: v1
>   kind: Deployment
>   name: my-deploy
> path: jsonpatch.json
> ```

## _PatchesStrategicMerge_

### 使用 `kustomization.yaml`

#### 字段名称：`patchesStrategicMerge`

此列表中的每个条目都应可以解析为 [StrategicMergePatch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md).

这些（也可能是部分的）资源文件中的 name 必须与已经通过 `resources` 加载的 name 字段匹配，或者通过 `bases` 中的 name 字段匹配。这些条目将用于 _patch_（修改）已知资源。

推荐使用小的 patches，例如：修改内存的 request/limit，更改 ConfigMap 中的 env 变量等。小的 patches 易于维护和查看，并且易于在 overlays 中混合使用。

```yaml
patchesStrategicMerge:
- service_port_8888.yaml
- deployment_increase_replicas.yaml
- deployment_increase_memory.yaml
```

patch 内容也可以是一个inline string：

```yaml
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: nignx:latest
```

请注意，kustomize 不支持同一个 patch 对象中包含多个 _删除_ 指令。要从一个对象中删除多个字段或切片元素，需要创建一个单独的 patch，以执行所有需要的删除。

### 使用插件

#### Arguments

> Paths \[\][types.PatchStrategicMerge]
>
> Patches string

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: PatchStrategicMergeTransformer
> metadata:
>   name: not-important-to-example
> paths:
> - patch.yaml
> ```

## _PatchTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`patches`

这个列表中的每个条目应该解析到一个 Patch 对象，其中包括一个 patch 和一个目标选择器。patch 可以是 Strategic Merge Patch 或 JSON patch，也可以是 patch 文件或 inline string。目标选择器可以通过 group、version、kind、name、namespace、标签选择器和注释选择器来选择资源，选择一个或多个匹配所有指定字段的资源来应用 patch。

```yaml
patches:
- path: patch.yaml
  target:
    group: apps
    version: v1
    kind: Deployment
    name: deploy.*
    labelSelector: "env=dev"
    annotationSelector: "zone=west"
- patch: |-
    - op: replace
      path: /some/existing/path
      value: new value
  target:
    kind: MyKind
    labelSelector: "env=dev"
```

The `name` and `namespace` fields of the patch target selector are
automatically anchored regular expressions. This means that the value `myapp`
is equivalent to `^myapp$`。

### 使用插件

#### Arguments

> Path string
>
> Patch string
>
> Target \*[types.Selector]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: PatchTransformer
> metadata:
>   name: not-important-to-example
> patch: '[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:latest"}]'
> target:
>   name: .*Deploy
>   kind: Deployment
> ```

## _PrefixSuffixTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`namePrefix`, `nameSuffix`

为所有资源的名称添加前缀或后缀。

例如：将 deployment 名称从 `wordpress` 变为 `alices-wordpress` 或  `wordpress-v2` 或 `alices-wordpress-v2`。

```yaml
namePrefix: alices-
nameSuffix: -v2
```

如果资源类型是 ConfigMap 或 Secret，则在哈希值之前添加后缀。

### 使用插件

#### Arguments

> Prefix     string
>
> Suffix     string
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: PrefixSuffixTransformer
> metadata:
>   name: not-important-to-example
> prefix: baked-
> suffix: -pie
> fieldSpecs:
>   - path: metadata/name
> ```

## _ReplicaCountTransformer_

### 使用 `kustomization.yaml`

#### 字段名称：`replicas`

修改资源的副本数。

例如：对于如下 kubernetes Deployment 片段：

```yaml
kind: Deployment
metadata:
  name: deployment-name
spec:
  replicas: 3
```

在 kustomization 中添加以下内容，将副本数更改为 5：

```yaml
replicas:
- name: deployment-name
  count: 5
```

该字段内容为列表，所以可以同时修改许多资源。

由于这个声明无法设置 `kind:` 或 `group:`，所以他只能匹配如下资源中的一种：

- `Deployment`
- `ReplicationController`
- `ReplicaSet`
- `StatefulSet`

对于更复杂的用例，请使用 patch 。

### 使用插件

#### Arguments

> Replica [types.Replica]
>
> FieldSpecs \[\][config.FieldSpec]

#### Example
>
> ```yaml
> apiVersion: builtin
> kind: ReplicaCountTransformer
> metadata:
>   name: not-important-to-example
> replica:
>   name: myapp
>   count: 23
> fieldSpecs:
> - path: spec/replicas
>   create: true
>   kind: Deployment
> - path: spec/replicas
>   create: true
>   kind: ReplicationController
> ```

## _SecretGenerator_

### 使用 `kustomization.yaml`

#### 字段名称：`secretGenerator`

列表中的每个条目都将生成一个 Secret（合计可以生成 n 个 Secrets）。

功能与之前描述的 `configMapGenerator` 字段类似。

```yaml
secretGenerator:
- name: app-tls
  files:
  - secret/tls.cert
  - secret/tls.key
  type: "kubernetes.io/tls"
- name: app-tls-namespaced
  # you can define a namespace to generate
  # a secret in, defaults to: "default"
  namespace: apps
  files:
  - tls.crt=catsecret/tls.cert
  - tls.key=secret/tls.key
  type: "kubernetes.io/tls"
- name: env_file_secret
  envs:
  - env.txt
  type: Opaque
- name: secret-with-annotation
  files:
  - app-config.yaml
  type: Opaque
  options:
    annotations:
      app_config: "true"
    labels:
      app.kubernetes.io/name: "app2"
```

### 使用插件

#### Arguments

> [types.ObjectMeta]
>
> [types.SecretArgs]

#### Example

> ```yaml
> apiVersion: builtin
> kind: SecretGenerator
> metadata:
>   name: my-secret
>   namespace: whatever
> behavior: merge
> envs:
> - a.env
> - b.env
> files:
> - obscure=longsecret.txt
> literals:
> - FRUIT=apple
> - VEGETABLE=carrot
> ```
