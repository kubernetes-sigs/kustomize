---
title: "configMapGenerator"
linkTitle: "configMapGenerator"
type: docs
description: >
    生成 ConfigMap 资源.
---

列表中的每个条目都将生成一个 ConfigMap （合计可以生成 n 个 ConfigMap）。

以下示例创建四个 ConfigMap：

- 第一个使用给定文件的名称和内容创建数据
- 第二个使用文件中的键/值对将数据创建为键/值
- 第三个使用 `literals` 中的键/值对创建数据作为键/值
- 第四个通过 `options` 设置单个 ConfigMap 的注释和标签

每个 configMapGenerator 项均接受的参数 `behavior: [create|replace|merge]`，这个参数允许修改或替换父级现有的 configMap。

此外，每个条目都有一个 `options` 字段，该字段具有与 kustomization 文件的 `generatorOptions` 字段相同的子字段。

`options` 字段允许用户为生成的实例添加标签和（或）注释，或者分别禁用该实例名称的哈希后缀。此处添加的标签和注释不会被 kustomization 文件 `generatorOptions` 字段关联的全局选项覆盖。但是如果全局 `generatorOptions` 字段指定 `disableNameSuffixHash: true`，其他 `options` 的设置将无法将其覆盖。

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

# 这个标签将添加到所有的 ConfigMap 和 Secret 中。
generatorOptions:
  labels:
    fruit: apple

configMapGenerator:
- name: my-java-server-props
  behavior: merge
  files:
  - application.properties
  - more.properties
- name: my-java-server-env-file-vars
  envs:
  - my-server-env.properties
  - more-server-props.env
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
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

configMapGenerator:
- name: app-whatever
  files:
  - myFileName.ini=whatever.ini
```
