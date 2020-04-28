# Kustomization 文件字段

[field-name-namespace]: ../plugins/builtins.md#field-name-namespace
[field-name-images]: ../plugins/builtins.md#field-name-images
[field-names-namePrefix-nameSuffix]: ../plugins/builtins.md#field-names-namePrefix-nameSuffix
[field-name-patches]: ../plugins/builtins.md#field-name-patches
[field-name-patchesStrategicMerge]: ../plugins/builtins.md#field-name-patchesStrategicMerge
[field-name-patchesJson6902]: ../plugins/builtins.md#field-name-patchesJson6902
[field-name-replicas]: ../plugins/builtins.md#field-name-replicas
[field-name-secretGenerator]: ../plugins/builtins.md#field-name-secretGenerator
[field-name-commonLabels]: ../plugins/builtins.md#field-name-commonLabels
[field-name-commonAnnotations]: ../plugins/builtins.md#field-name-commonAnnotations
[field-name-configMapGenerator]: ../plugins/builtins.md#field-name-configMapGenerator

介绍 [kustomization.yaml](../glossary.md#kustomization) 配置文件中各字段的含义。

## Resources

现有可定制对象。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
|[resources](#resources) | list | 包含 k8s API 对象的文件，或其他包含 `kustomization.yaml` 文件的目录。 |
|[CRDs](#crds)| list | CDR 文件，允许在资源列表中指定自定义资源。 |

## Generators

资源生成器。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
|[configMapGenerator](#configmapgenerator)| list  | 列表中的每个条目都将生成一个 ConfigMap （合计可以生成 n 个 ConfigMap）。 |
|[secretGenerator](#secretgenerator)| list  | 列表中的每个条目都将生成一个 Secret（合计可以生成 n 个 Secrets）。 |
|[generatorOptions](#generatoroptions)| string | generatorOptions 可以修改所有 ConfigMapGenerator 和 SecretGenerator 的行为。 |
|[generators](#generators)| list | [插件](../plugins)配置文件。 |

## Transformers

资源字段转换项。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| [commonLabels](#commonlabels) | string | 为所有资源和 selectors 增加 Labels 。 |
| [commonAnnotations](#commonannotations) | string | 为所有资源增加 Annotations 。 |
| [images](#images) | list | 无需使用 patches，即可修改镜像的名称、tag 或 image digest。 |
| [inventory](#inventory) | struct | 用于生成一个包含清单信息的对象。 |
| [namespace](#namespace)   | string | 为所有 resources 添加 namespace 。 |
| [namePrefix](#nameprefix) | string | 为所有资源的名称添加前缀。 |
| [nameSuffix](#namesuffix) | string | 为所有资源的名称添加后缀。 |
| [replicas](#replicas) | list | 修改资源的副本数。 |
| [patchesStrategicMerge](#patchesstrategicmerge) | list | 此列表中的每个条目都应可以解析为部分或完整的资源定义文件。 |
| [patchesJson6902](#patchesjson6902) | list | 列表中的每个条目都应可以解析为 Kubernetes 对象和将应用于该对象的 JSON patch 。 |
| [transformers](#transformers) | list | [插件](../plugins)配置文件。 |

## Meta

[k8s metadata]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| [vars](#vars) | string | 将一个对象中的字段值插入另一个对象中。 |
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

详见 [field-name-commonLabels]。

### commonAnnotations

详见 [field-name-commonAnnotations].

### configMapGenerator

详见 [field-name-configMapGenerator].

### crds

此列表中的每个条目都应该是自定义资源定义（CRD）文件的相对路径。

该字段的存在是为了让 kustomize 识别用户自定义的 CRD ，并对这些类型中的对象应用适当的转换。

典型用例：CRD 引用 ConfigMap 对象

在 kustomization 中，ConfigMap 对象名称可能会通过 `namePrefix` 、`nameSuffix` 或 `hashing` 来更改 CRD 对象中该 ConfigMap 对象的名称，
引用时需要以相同的方式使用 `namePrefix` 、 `nameSuffix` 或 `hashing` 来进行更新。

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

generatorOptions 可以修改所有 [ConfigMapGenerator](#configmapgenerator) 和 [SecretGenerator](#secretgenerator) 的行为。

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

详见 [field-name-images]。

### inventory

详见 [inventory object](inventory_object.md)。

### kind

该字段默认值为：

```
kind: Kustomization
```


### namespace

详见 [field-name-namespace]。

### namePrefix

详见 [field-names-namePrefix-nameSuffix]。

### nameSuffix

详见 [field-names-namePrefix-nameSuffix]。

### patches

详见 [field-name-patches]。

### patchesStrategicMerge

详见 [field-name-patchesStrategicMerge]。

### patchesJson6902

详见 [field-name-patchesJson6902]。

### replicas

详见 [field-name-replicas]。

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

文件应包含 YAML 格式的 k8s 资源。一个资源描述文件可以含有多个由（`---`）分隔的资源。
应该包含 `resources` 字段的 kustomization 文件的指定文件目录的相对路径。

[hashicorp URL]: https://github.com/hashicorp/go-getter#url-format

目录规范可以是相对、绝对或部分的 URL。URL 规范应遵循 [hashicorp URL] 格式。该目录必须包含 `kustomization.yaml` 文件。

### secretGenerator

详见 [field-name-secretGenerator]。

### vars

Vars 用于从一个 resource 字段中获取值，并将该值插入指定位置 - 反射功能。

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
[/api/konfig/builtinpluginconsts/varreference.go](/api/konfig/builtinpluginconsts/varreference.go)

默认目标是所有容器 command args 和 env 字段。

Vars _不应该_ 被用于 kustomize 已经处理过的配置中插入 names 。
例如， Deployment 可以通过 name 引用 ConfigMap ，如果 kustomize 更改 ConfigMap 的名称，则知道更改 Deployment 中的引用的 name 。
