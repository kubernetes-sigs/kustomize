[English](../README.md) | 简体中文

# 文档

 * [安装说明](INSTALL.md)

 * [示例](../../examples) - 各种使用流程和概念的详细演示。

 * [术语表](glossary.md) - 用于消除术语歧义。

 * [Kustomize 字段](fields.md) - 介绍 [kustomization](../glossary.md#kustomization) 文件中各字段的含义。

 * [插件](../plugins) - 使用自定义的资源生成器和资源转换器来拓展 kustomize 功能。

 * [工作流](workflows.md) - 使用定制及使用现成配置使用的一些步骤。

 * [FAQ](../FAQ.md)


## 发行说明

 * [3.1](../v3.1.0.md) - 2019年7月下旬，扩展 patches 和改进的资源匹配。

 * [3.0](../v3.0.0.md) - 2019年6月下旬，插件开发者发布。

 * [2.1](../v2.1.0.md) - 2019年6月18日
 插件、有序资源等。

 * [2.0](../v2.0.0.md) - 2019年3月
   可以在 [kubectl v1.14][kubectl] 中使用 kustomize [v2.0.3] 。

 * [1.0](../v1.0.1.md) - 2018年5月
   于 [kubectl repository] 开发后的首发版本。


## 行为守则

 * [版本控制](../versioningPolicy.md) - kustomize 代码及 kustomization 文件的版本控制策略。

 * [规避功能](../eschewedFeatures.md) - 目前 Kustomize 不支持某些功能的原因。

 * [贡献指南](../../CONTRIBUTING.md) - 请在提交 PR 之前阅读。

 * [行为准则](../../code-of-conduct.md)

>声明：部分文档可能稍微滞后于英文版本，同步工作持续进行中

[v2.0.3]: https://github.com/kubernetes-sigs/kustomize/releases/tag/v2.0.3
[kubectl]: https://kubernetes.io/blog/2019/03/25/kubernetes-1-14-release-announcement
[kubectl repository]: https://github.com/kubernetes/kubectl
