[English](../README.md) | 简体中文

# 示例

这些示例默认 `kustomize` 在您的 `$PATH` 中。

这些示例通过了 [pre-commit](../../travis/pre-commit.sh) 测试，并且应该与 HEAD 一起使用。

```
go get sigs.k8s.io/kustomize/v3/cmd/kustomize
```

基本用法

  * [configGenerations](configGeneration.md) - 当 ConfigMapGenerator 修改时进行滚动更新。

  * [combineConfigs](../combineConfigs.md) - 融合来自不同用户的配置数据（例如来自 devops/SRE 和 developers）。

  * [generatorOptions](../generatorOptions.md) -修改所有 ConfigMapGenerator 和 SecretGenerator 的行为。

  * [vars](../wordpress/README.md) - 通过 vars 将一个资源的数据注入另一个资源的容器参数 （例如，为 wordpress 指定 SQL 服务）。

  * [image names and tags](../image.md) - 在不使用 patch 的情况下更新镜像名称和标签。

  * [remote target](../remoteBuild.md) - 通过 github URL 来构建 kustomization 。

  * [json patch](jsonpatch.md) -在 kustomization 中应用 json patch 。

  * [patch multiple objects](patchMultipleObjects.md) - 通过一个patch来修改多个资源。

高级用法

  - generator 插件:

    * [last mile helm](../chart.md) - 对 helm chart 进行 last mile 修改。

    * [secret generation](../secretGeneratorPlugin.md) - 生成 Secret。

  - transformer 插件:

   * [validation transformer](../validationTransformer/README.md) - 通过 transformer 验证资源。

  - 定制内建 transformer 配置

    * [transformer configs](../transformerconfigs/README.md) - 自定义 transformer 配置。

多 Variant 示例

 * [hello world](helloWorld.md) - 部署多个不同配置的 Hello World 服务。

 * [LDAP](../ldap/README.md) - 部署多个配置不同的 LDAP 服务。

 * [springboot](../springboot/README.md) - 从头开始创建一个 Spring Boot 项目的生产配置。

 * [mySql](../mySql/README.md) - 从头开始创建一个 MySQL 的生产配置。

 * [breakfast](../breakfast.md) - 给 Alice 和 Bob 定制一顿早餐 :)

 * [multibases](../multibases/README.md) - 使用相同的 base 生成三个 variants（dev，staging，production）。

>声明：部分文档可能稍微滞后于英文版本，同步工作持续进行中