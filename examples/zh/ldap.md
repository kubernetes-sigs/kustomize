[base]: ../../docs/glossary.md#base
[gitops]: ../../docs/glossary.md#gitops
[kustomization]: ../../docs/glossary.md#kustomization
[overlay]: ../../docs/glossary.md#overlay
[overlays]: ../../docs/glossary.md#overlay
[variant]: ../../docs/glossary.md#variant
[variants]: ../../docs/glossary.md#variant

# 示例：LDAP 服务

步骤：

 1. 拉取已经存在的 [base] 配置
 2. 进行配置
 3. 基于 [base] 创建2个不同的 [overlays] (_staging_ 和 _production_)
 4. 运行 kustomize 或 kubectl 部署 staging 和 production

首先创建一个工作空间：

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

或者

> ```
> DEMO_HOME=~/ldap
> ```

## 创建 base

要使用 [overlays] 创建 [variant]，首先需要创建一个 [base]。

为了保证文档的精简，基础资源都在补充目录中，如果需要请下载它们：

<!-- @downloadBase @testAgainstLatestRelease -->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE

CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/ldap"

curl -s -o "$BASE/#1" "$CONTENT/base\
/{deployment.yaml,kustomization.yaml,service.yaml,env.startup.txt}"
```

检查这个目录：

<!-- @runTree -->
```
tree $DEMO_HOME
```

将会看到如下文件：

> ```
> /tmp/tmp.IyYQQlHaJP
> └── base
>     ├── deployment.yaml
>     ├── env.startup.txt
>     ├── kustomization.yaml
>     └── service.yaml
> ```

这些资源可以由 kubectl 立刻部署到集群上来实例化 _ldap_ 服务：

> ```
> kubectl apply -f $DEMO_HOME/base
> ```

注意 `kubectl -f` 只能识别 k8s 资源文件。

### The Base Kustomization

`base` 目录包含一个 [kustomization] 文件：

<!-- @showKustomization @testAgainstLatestRelease -->
```
more $BASE/kustomization.yaml
```

（可选）在 base 上运行 `kustomize`，并将结果打印到标准输出：

<!-- @buildBase @testAgainstLatestRelease -->
```
kustomize build $BASE
```

### Customize the base

为所有资源设置名称前缀：

<!-- @namePrefix @testAgainstLatestRelease -->
```
cd $BASE
kustomize edit set nameprefix "my-"
```

查看变化：
<!-- @checkNameprefix @testAgainstLatestRelease -->
```
kustomize build $BASE | grep -C 3 "my-"
```

## 创建 Overlays

创建 _staging_ 和 _production_ 的 [overlay]:

 * 为 _Staging_ 新增一个 ConfigMap
 * 为 _Production_ 添加持久化存储盘和更多的副本数
 * 现实两个 [variants] 的不同之处

<!-- @overlayDirectories @testAgainstLatestRelease -->
```
OVERLAYS=$DEMO_HOME/overlays
mkdir -p $OVERLAYS/staging
mkdir -p $OVERLAYS/production
```

#### Staging Kustomization

下载 staging 配置

<!-- @downloadStagingKustomization @testAgainstLatestRelease -->
```
curl -s -o "$OVERLAYS/staging/#1" "$CONTENT/overlays/staging\
/{config.env,deployment.yaml,kustomization.yaml}"
```

在 staging 配置中增加一个 ConfigMap
> ```cat $OVERLAYS/staging/kustomization.yaml
> (...truncated)
> configMapGenerator:
>   - name: env-config
>     files:
>       - config.env
> ```
和2个副本
> ```cat $OVERLAYS/staging/deployment.yaml
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   name: ldap
> spec:
>   replicas: 2
> ```

#### Production Kustomization

下载 production 配置
<!-- @downloadProductionKustomization @testAgainstLatestRelease -->
```
curl -s -o "$OVERLAYS/production/#1" "$CONTENT/overlays/production\
/{deployment.yaml,kustomization.yaml}"
```

在 production 的配置中增加为6副本和存储盘
> ```cat $OVERLAYS/production/deployment.yaml
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   name: ldap
> spec:
>   replicas: 6
>   template:
>     spec:
>       volumes:
>         - name: ldap-data
>           emptyDir: null
>           gcePersistentDisk:
>             pdName: ldap-persistent-storage
> ```

## 比较 overlays


`DEMO_HOME` 现在包括：

 * 一个 _base_ 目录：对拉取原始配置进行少量的定制

 * 一个 _overlays_ 目录：其中包含在集群中创建不同的 _staging_ 和 _production_ [variants] 所需的 kustomizations 文件和 patche 文件

查看目录结构和差异：

<!-- @listFiles -->
```
tree $DEMO_HOME
```

将会得到类似的内容：

> ```
> /tmp/tmp.IyYQQlHaJP1
> ├── base
> │   ├── deployment.yaml
> │   ├── env.startup.txt
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── production
>     │   ├── deployment.yaml
>     │   └── kustomization.yaml
>     └── staging
>         ├── config.env
>         ├── deployment.yaml
>         └── kustomization.yaml
> ```

直接对输出内容进行比较，以查看 _staging_ 和 _production_ 的不同之处：

<!-- @compareOutput -->
```
diff \
  <(kustomize build $OVERLAYS/staging) \
  <(kustomize build $OVERLAYS/production) |\
  more
```

输出的差异内容

> ```diff
> (...truncated)
> <   name: staging-my-ldap-configmap-kftftt474h
> ---
> >   name: production-my-ldap-configmap-k27f7hkg4f
> 85c75
> <   name: staging-my-ldap-service
> ---
> >   name: production-my-ldap-service
> 97c87
> <   name: staging-my-ldap
> ---
> >   name: production-my-ldap
> 99c89
> <   replicas: 2
> ---
> >   replicas: 6
> (...truncated)
> ```


## 部署

查看各个资源集：

<!-- @buildStaging @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging
```

<!-- @buildProduction @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/production
```

将上述命令通过管道传递给 kubectl 以进行部署：

> ```
> kustomize build $OVERLAYS/staging |\
>     kubectl apply -f -
> ```

> ```
> kustomize build $OVERLAYS/production |\
>    kubectl apply -f -
> ```
