[English](../README.md) | 简体中文

# 示例

要运行以下实例，请确保 `kustomize` 在环境变量 `$PATH` 中。

这些示例通过了 [tested](../../hack/testExamplesAgainstKustomize.sh) 测试，可以与最新的**发布版本**一起使用。

基本用法

  * [value add](valueAdd.md) -
    轻松的将一个字符串添加到变量字段中，包括包含路径的字段。
  
  * [config generation](configGeneration.md) -
    ConfigMapGenerator 发生变更时，滚动更新配置。
  
  * [combine configs](combineConfigs.md) -
    合并不通来源的配置数据。
    （例如：运维人员/SRE 和开发人员）
  
  * [generator options](generatorOptions.md) -
    修改所有 ConfigMap 和 Secret 的生成器的行为。
  
  * [vars](wordpress/README.md) - 
    通过变量将 k8s 运行时数据注入到容器参数。
    （例如：将WordPress指向SQL服务）
  
  * [image name and tags](image.md) - 直接更新镜像名称和标签（不使用 patch）。 
  
  * [remote target](remoteBuild.md) - 从 github URL 创建 kustomization 。 
  
  * [JSON patch](jsonpatch.md) - 在 kustomization 中应用 JSON 补丁。
  
  * [patch multiple objects](patchMultipleObjects.md) - 将一个补丁应用到多个对象。

高级用法

  - generator 插件:

    * [last mile helm](chart.md) - 对 helm chart 进行 last mile 修改。

    * [secret generation](secretGeneratorPlugin.md) - 生成 Secret。

  - transformer 插件:

    * [validation transformer](validationTransformer.md) - 通过 transformer 验证资源。

  - 定制内建 transformer 配置

    * [transformer configs](transformerconfigs.md) - 自定义 transformer 配置。

多功能整合示例

 * [hello world](helloWorld.md) - 部署多个不同配置的 Hello World 服务。

 * [LDAP](ldap.md) - 部署多个配置不同的 LDAP 服务。

 * [springboot](springboot.md) - 从头开始创建一个 Spring Boot 项目的生产配置。

 * [mySql](mysql.md) - 从头开始创建一个 MySQL 的生产配置。

 * [breakfast](breakfast.md) - 给 Alice 和 Bob 定制一顿早餐 :)

 * [multibases](multibases.md) - 使用相同的 base 生成三个 variants（dev，staging，production）。

 * [components](../components.md) - 通过重用配置，组合三个具有共同数据的变体（community, enterprise, dev）。

>声明：部分文档可能稍微滞后于英文版本，同步工作持续进行中
