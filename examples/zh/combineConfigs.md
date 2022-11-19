[overlay]: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#overlay
[target]: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#target

# 示例：devops和开发配合管理配置数据

场景：在生产环境中有一个基于 Java 由多个内部团队（注册、结账和搜索等）共同开发的商店服务。

这个服务在不同的环境中运行：_development_、 _testing_、 _staging_ 和 _production_，从 Java 的 properties 文件中读取配置。

为每个环境维护一个大的 properties 文件是很困难的。这个文件需要频繁的修改，并且这些修改都需要由 devops 工程师来进行，因为：

  1. 这个文件包含 devops 工程师需要知道，而开发人员不必知道的值
  2. 比如生产环境的 properties 包含敏感数据，比如生产数据库的登录凭据。

## Property sharding

通过一些研究，我们注意到属性可以分为不同的类别。

### Property sharding

例如：国际化数据、物理常量，外部服务位置等静态数据。

_这些无论哪个环境，都一样的配置。_

这些都只需要一组配置。将这组配置放在一个文件中：

 * `common.properties`

### Plumbing properties

例如：静态资源（HTML、CSS、JavaScript）的位置，产品和用户的数据表，负载均衡的端口，日志收集等。

_这些属性的不同，恰恰是环境的不同之处。_

DevOps 或 SRE 工程师需要完全控制生产环境中的这些配置；测试需要调整数据库来支持测试；而开发则希望尝试开发中遇到的各种不同的情景。

将这些值放入

 * `development/plumbing.properties`
 * `staging/plumbing.properties`
 * `production/plumbing.properties`


### Secret properties

例如：用户表的位置、数据库凭证、解密密钥等。

_这些需要 devops 工程师控制，其他人没有访问权限。_

将这些值放入

 * `development/secret.properties`
 * `staging/secret.properties`
 * `production/secret.properties`

[kubernetes secret]: https://kubernetes.io/docs/tasks/inject-data-application/distribute-credentials-secure/

例如使用 unix 文件权限和模式来限制访问控制，或者使用更好的方法-使用专门用于存储密码的服务，并且使用 kustomize 中的 `secretGenerator` 字段在 Kubernetes 中创建 secret 来存储密码。

<!--
secretGenerator:
- name: app-tls
  files:
    tls.crt=tls.cert
    tls.key=tls.key
  type: "kubernetes.io/tls"
EOF
-->

## 混合管理方法

基于相同的 base 创建 _n_ 个 overlays 来创建 _n_ 个集群环境的方法。

在本例的其余部分，我们将使用 _n==2_，这里只使用 _development_ 和 _production_ ，可以使用相同的方法来增加更多的环境。

运行 `kustomize build` 基于 [overlay] 的 [target] 来创建集群环境。

[helloworld]: helloWorld.md

以下示例将执行此操作，但将侧重于 configMap 构建，而不用担心如何将 configMaps 关联到 Deployment（[helloworld] 示例中介绍的）。

所有文件（包括共享 property 文件）都将在目录树中创建，目录中包含 base 和 overlay 文件的目录，这些都与 [helloworld] 中演示的一致。

它将全部存在于此工作目录中：

<!-- @makeWorkplace @test -->
```bash
DEMO_HOME=$(mktemp -d)
```

### 创建 base

<!-- kubectl create configmap BOB --dry-run -o yaml --from-file db. -->

创建放置 base 配置的路径：

<!-- @baseDir @test -->
```bash
mkdir -p $DEMO_HOME/base
```

向 base 中的插入数据，base 中应该包含所有环境共有的资源，这里我们只定义一个 java properties 文件，以及一个引用他们的 `kustomization` 文件。

<!-- @baseKustomization @test -->
```bash
cat <<EOF >$DEMO_HOME/base/common.properties
color=blue
height=10m
EOF

cat <<EOF >$DEMO_HOME/base/kustomization.yaml
configMapGenerator:
- name: my-configmap
  files:
  - common.properties
EOF
```

### 创建并使用 overlay 用于 _开发_

创建一个 overlays 目录：

<!-- @overlays @test -->
```bash
OVERLAYS=$DEMO_HOME/overlays
```

创建 _development_ overlay：

<!-- @developmentFiles @test -->
```bash
mkdir -p $OVERLAYS/development

cat <<EOF >$OVERLAYS/development/plumbing.properties
port=30000
EOF

cat <<EOF >$OVERLAYS/development/secret.properties
dbpassword=mothersMaidenName
EOF

cat <<EOF >$OVERLAYS/development/kustomization.yaml
resources:
- ../../base
namePrefix: dev-
nameSuffix: -v1
configMapGenerator:
- name: my-configmap
  behavior: merge
  files:
  - plumbing.properties
  - secret.properties
EOF
```

现在可以生成开发使用的 configMaps ：

<!-- @runDev @test -->
```bash
kustomize build $OVERLAYS/development
```

#### 检查 ConfigMap 名称

可以在输出中看到生成的 `ConfigMap` 名称。

名称应该是这样的：`dev-my-configmap-v1-2gccmccgd5`：

 * `"dev-"` 来自 `namePrefix` 字段
 * `"my-configmap"` 来自 `configMapGenerator/name` 字段
 * `"-v1"` 来自 `nameSuffix` 字段
 * `"-2gccmccgd5"` 为哈希值，是 `kustomize` 根据 configMap 的内容计算的

哈希后缀很关键，如果 configMap 内容发生变化， configMap 的名称也会发生变化，以及从 `kustomize` 出现在 YAML 输出中的对该名称的所有引用。

名称更改意味着如果使用类似命令将此 YAML 应用于群集，则 Deployment 将执行滚动更新重启以获取新数据。

> ```bash
> kustomize build $OVERLAYS/development | kubectl apply -f -
> ```

Deployment 无法自动检测 ConfigMap 是否发生改变。

如果更改 configMap 的数据, 而不更改其名称以及对该名称的所有引用, 则必须重新启动Deployment中的那些Pods以获取更改。

最佳的做法就是将 configMap 视为不变的。

不去编辑 configMap ，而是使用 _新_ 的名称的 _新_ configMap，并在 Deployment 中引用新的 configMap 。而 `kustomize` 使用 `configMapGenerator` 指令和相关的命名控件使这很容易。

### 创建并且使用 overlay 用于 _生产_

接下来创建 _production_ overlay 的文件：

<!-- @productionFiles @test -->
```bash
mkdir -p $OVERLAYS/production

cat <<EOF >$OVERLAYS/production/plumbing.properties
port=8080
EOF

cat <<EOF >$OVERLAYS/production/secret.properties
dbpassword=thisShouldProbablyBeInASecretInstead
EOF

cat <<EOF >$OVERLAYS/production/kustomization.yaml
resources:
- ../../base
namePrefix: prod-
configMapGenerator:
- name: my-configmap
  behavior: merge
  files:
  - plumbing.properties
  - secret.properties
EOF
```

现在可以生成用于生产的 configMap：

<!-- @runProd @test -->
```bash
kustomize build $OVERLAYS/production
```

可以直接在 CI/CD 流程中执行如下命令，将应用部署到集群：

> ```bash
> kustomize build $OVERLAYS/production | kubectl apply -f -
> ```
