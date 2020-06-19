---
title: "vars"
linkTitle: "vars"
type: docs
description: >
    Substitute name references.
---

Vars 用于从一个 resource 字段中获取值，并将该值插入指定位置 - 反射功能。

例如，假设需要在容器的 command 中指定了 Service 对象的名称，并在容器的 env 中指定了 Secret 对象的名称来确保以下内容可以正常工作：

```yaml

containers:
  - image: myimage
    command: ["start", "--host", "$(MY_SERVICE_NAME)"]
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
```

则可以在 `vars：` 中添加如下内容：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

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
