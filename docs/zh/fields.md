# Kustomization 文件字段

介绍 [kustomization](../glossary.md#kustomization) 文件中各字段的含义。

## Resources

现有可定制对象。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
|[resources](#resources) | list | 包含 k8s API 对象的文件，或其他包含 kustomizations 文件的目录。 |
|[CRDs](#crds)| list | CDR 文件，以允许在资源列表中指定自定义资源。 |

## Generators

生成可定制的对象。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
|[configMapGenerator](#configmapgenerator)| list  | 列表中的每个条目都将创建一个 ConfigMap （它是n个 ConfigMap 的生成器）。 |
|[secretGenerator](#secretgenerator)| list  | 此列表中的每个条目都将创建一个 Secret 资源（它是n个 secrets 的生成器）。 |
|[generatorOptions](#generatoroptions)| string | generatorOptions 可以修改所有 ConfigMapGenerator 和 SecretGenerator 的行为。 |
|[generators](#generators)| list | [插件](../plugins)配置文件。 |

## Transformers

可用的转换。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| [commonLabels](#commonlabels) | string | 为所有资源和 selectors 增加 Labels 。 |
| [commonAnnotations](#commonannotations) | string | 为所有资源增加 Annotations 。 |
| [images](#images) | list | 修改镜像的名称、tag 或 image digest ，而无需使用 patches 。 |
| [inventory](#inventory) | struct | 用于生成一个包含清单信息的对象。 |
| [namespace](#namespace)   | string | 为所有 resources 添加 namespace 。 |
| [namePrefix](#nameprefix) | string | 该字段的值将添加在所有资源的名称之前。 |
| [nameSuffix](#namesuffix) | string | 该字段的值将添加在所有资源的名称后面。 |
| [replicas](#replicas) | list | 修改资源的副本数。 |
| [patchesStrategicMerge](#patchesstrategicmerge) | list | 此列表中的每个条目都应可以解析为部分或完整的资源定义文件。 |
| [patchesJson6902](#patchesjson6902) | list | 列表中的每个条目都应可以解析为 kubernetes 对象和将应用于该对象的 JSON patch 。 |
| [transformers](#transformers) | list | [插件](../plugins)配置文件。 |

## Meta

[k8s metadata]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| [vars](#vars) | string | 获取一个对象中的字段并插入到另外的对象中。 |
| [apiVersion](#apiversion) | string | [k8s metadata] 字段。 |
| [kind](#kind) | string | [k8s metadata] 字段。 |

----

### apiVersion

该字段默认值为：
```
apiVersion: kustomize.config.k8s.io/v1beta1
```

### bases

`bases` 字段在 v2.1.0 中已被弃用。

该条目已被移动到 [resources](#resources) 字段中。

### commonLabels

为所有资源和 selectors 增加 Labels

```
commonLabels:
  someName: someValue
  owner: alice
  app: bingo
```

### commonAnnotations

为所有资源增加 Annotations ，和 labels 一样是 key:value 的键值对。

```
commonAnnotations:
  oncallPager: 800-555-1212
```

### configMapGenerator

列表中的每个条目都将创建一个 ConfigMap （它是n个 ConfigMap 的生成器）。

下面的示例创建了两个 ConfigMaps：

- 一个具有给定文件的名称和内容
- 另一个包含 key/value 键值对数据

每个 configMapGenerator 项都可以使用 `behavior: [create|replace|merge]` 参数。

允许 overlay 从父级修改或替换现有的 configMap。

```
configMapGenerator:
- name: myJavaServerProps
  files:
  - application.properties
  - more.properties
- name: myJavaServerEnvVars
  literals:	
  - JAVA_HOME=/opt/java/jdk
  - JAVA_TOOL_OPTIONS=-agentlib:hprof
```

### crds

此列表中的每个条目都应该是自定义资源定义（CRD）文件的相对路径。

该字段的存在是为了让 kustomize 知道用户自定义的 CRD ，并对这些类型中的对象应用适当的转换。

典型用例：CRD 引用 ConfigMap 对象

在 kustomization 中，ConfigMap 对象名称可能会通过 namePrefix 、nameSuffix 或 hashing 来更改 CRD 对象中此 ConfigMap 对象的名称，
引用时需要以相同的方式使用 namePrefix 、 nameSuffix 或 hashing 来进行更新。

Annotations 可以放入 openAPI 的定义中：

-  "x-kubernetes-annotation": ""
-  "x-kubernetes-label-selector": ""
-  "x-kubernetes-identity": ""
-  "x-kubernetes-object-ref-api-version": "v1",
-  "x-kubernetes-object-ref-kind": "Secret",
-  "x-kubernetes-object-ref-name-key": "name",

```

crds:
- crds/typeA.yaml
- crds/typeB.yaml
```


### generatorOptions

generatorOptions 修改所有 [ConfigMapGenerator](#configmapgenerator) 和 [SecretGenerator](#secretgenerator) 的行为。

```
generatorOptions:
  # 为所有生成的资源添加 labels
  labels:
    kustomize.generated.resources: somevalue
  # 为所有生成的资源添加 annotations
  annotations:
    kustomize.generated.resource: somevalue
  # disableNameSuffixHash 为 true 时将禁止默认的在名称后添加哈希值后缀的行为
  disableNameSuffixHash: true
```

### generators

[插件](../plugins)生成器配置文件列表。

```
generators:
- mySecretGeneratorPlugin.yaml
- myAppGeneratorPlugin.yaml
```

### images

修改镜像的名称、tag 或 image digest ，而无需使用 patches 。例如，对于这种 kubernetes Deployment 片段：

```
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

可以通过以下方式更改 `image` ：
 
 - `postgres:8` to `my-registry/my-postgres:v1`,
 - nginx tag `1.7.9` to `1.8.0`,
 - image name `my-demo-app` to `my-app`,
 - alpine's tag `3.7` to a digest value

可以在 *kustomization* 中添加以下内容：

```
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

### inventory

详见 [inventory object](inventory_object.md)。

### kind

该字段默认值为：

```
kind: Kustomization
```


### namespace

为所有 resources 添加 namespace 。

```
namespace: my-namespace
```

### namePrefix

该字段的值将添加在所有资源的名称之前，例如 将资源名称 `wordpress` 变为 `alices-wordpress` 。

```
namePrefix: alices-
```

### nameSuffix

该字段的值将添加在所有资源的名称后面，例如 将资源名称 `wordpress` 变为 `wordpress-v2` 。

如果资源类型为 ConfigMap 或 Secret ，则在哈希值之前附加后缀。

```
nameSuffix: -v2
```

### patchesStrategicMerge

此列表中的每个条目都应可以解析为部分或完整的资源定义文件。

这些（也可能是部分的）资源文件中的 name 必须与已经通过 `resources` 加载的 name 字段匹配，或者通过 `bases` 中的 name 字段匹配。这些条目将用于 _patch_（修改）已知资源。

推荐使用小的 patches，例如：修改内存的 request/limit，更改 ConfigMap 中的 env 变量等小的 patches 易于维护和查看，并且易于在 overlays 中混合使用。

```
patchesStrategicMerge:
- service_port_8888.yaml
- deployment_increase_replicas.yaml
- deployment_increase_memory.yaml
```

### patchesJson6902

patchesJson6902 列表中的每个条目都应可以解析为 kubernetes 对象和将应用于该对象的 JSON patch

JSON patch 的文档地址：https://tools.ietf.org/html/rfc6902

目标字段指向的 kubernetes 对象的 group、 version、 kind、 name 和 namespace 在同一 kustomization 内 path 字段内容是 JSON patch 文件的相对路径。

patch 文件中的内容可以如下这种 JSON 格式：

```
 [
   {"op": "add", "path": "/some/new/path", "value": "value"},
   {"op": "replace", "path": "/some/existing/path", "value": "new value"}
 ]
 ```

也可以使用 YAML 格式表示：

```
- op: add
  path: /some/new/path
  value: value
- op: replace
  path: /some/existing/path
  value: new value
```

```
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

### replicas

修改资源的副本数。

例如：对于如下 kubernetes Deployment 片段：

```
kind: Deployment
metadata:
  name: deployment-name
spec:
  replicas: 3
```

在 kustomization 中添加以下内容，将副本数更改为5：

```
replicas:
- name: deployment-name
  count: 5
```

该字段内容为列表，所以可以同时修改许多资源。

#### Limitation

由于这个声明无法设置 `kind:` 或 `group:` 它将匹配任何可以匹配名称的 `group` 和 `kind` ，并且它是以下之一：
- `Deployment`
- `ReplicationController`
- `ReplicaSet`
- `StatefulSet`

对于更复杂的用例，请使用 patch 。

### resources

该条目可以是指向本地目录的相对路径，也可以是指向远程仓库中的目录的 URL，例如：

```
resource:
- myNamespace.yaml
- sub-dir/some-deployment.yaml
- ../../commonbase
- github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
- deployment.yaml
- github.com/kubernets-sigs/kustomize//examples/helloWorld?ref=test-branch
```

将以深度优先的顺序读取和处理资源。


文件应包含 YAML 格式的 k8s 资源。一个资源描述文件可以含有多个由（“---”）分隔的资源。
应该包含 `resources` 字段的 kustomization 文件的指定文件目录的相对路径。

[hashicorp URL]: https://github.com/hashicorp/go-getter#url-format

目录规范可以是相对、绝对或部分的 URL。URL 规范应遵循 [hashicorp URL] 格式。该目录必须包含 `kustomization.yaml` 文件。

### secretGenerator

此列表中的每个条目都将创建一个 Secret 资源（它是n个 secrets 的生成器）。

```
secretGenerator:
- name: app-tls
  files:
  - secret/tls.cert
  - secret/tls.key
  type: "kubernetes.io/tls"
- name: app-tls-namespaced
  # you can define a namespace to generate secret in, defaults to: "default"
  namespace: apps
  files:
  - tls.crt=catsecret/tls.cert
  - tls.key=secret/tls.key
  type: "kubernetes.io/tls"
- name: env_file_secret
  envs:
  - env.txt
  type: Opaque
```

### vars

Vars 用于从一个 resource 字段中获取文本，并将该文本插入指定位置 - 反射功能。

例如，假设需要在容器的 command 中指定了 Service 对象的名称，并在容器的 env 中指定了 Secret 对象的名称来确保以下内容可以正常工作：

```
containers:
  - image: myimage
    command: ["start", "--host", "$(MY_SERVICE_NAME)"]
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
```

则可以在 `vars：` 中添加如下内容：

```
vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
- name: MY_SERVICE_NAME
  objref:
    kind: Service
    name: my-service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: ANOTHER_DEPLOYMENTS_POD_RESTART_POLICY
  objref:
    kind: Deployment
    name: my-deployment
    apiVersion: apps/v1
  fieldref:
    fieldpath: spec.template.spec.restartPolicy
```
var 是包含该对象的变量名、对象引用和字段引用的元组。

字段引用是可选的，默认为 `metadata.name`，这是正常的默认值，因为 kustomize 用于生成或修改 resources 的名称。

在撰写本文档时，仅支持字符串类型字段，不支持 ints，bools，arrays 等。例如，在某些pod模板的容器编号2中提取镜像的名称是不可能的。

变量引用，即字符串 '$(FOO)' ，只能放在 kustomize 配置指定的特定对象的特定字段中。

关于 vars 的默认配置数据可以查看：
https://github.com/kubernetes-sigs/kustomize/blob/master/pkg/transformers/config/defaultconfig/varreference.go

默认目标是所有容器 command args 和 env 字段。

Vars _不应该_ 被用于 kustomize 已经处理过的配置中插入 names 。
例如， Deployment 可以通过 name 引用 ConfigMap ，如果 kustomize 更改 ConfigMap 的名称，则知道更改 Deployment 中的引用的 name 。
