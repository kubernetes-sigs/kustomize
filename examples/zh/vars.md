# 示例: 将 k8s runtime 数据注入容器

本教程将会介绍如何声明变量以及如何在容器中的命令使用变量。要注意的是，变量的查找和替换并不适用于任意字段，默认仅适用于容器的env，args和command。

运行WordPress，以下是必须的：

- WordPress 连接 MySQL 数据库
- MySQL 服务可以被 WordPress 容器访问

首先构建一个工作空间：
<!-- @makeDemoHome @test -->
```bash
DEMO_HOME=$(mktemp -d)
MYSQL_HOME=$DEMO_HOME/mysql
mkdir -p $MYSQL_HOME
WORDPRESS_HOME=$DEMO_HOME/wordpress
mkdir -p $WORDPRESS_HOME
```

### 下载 resources

下载 WordPress 的 resources 和 `kustomization.yaml` 。

<!-- @downloadResources @test -->
```bash
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress/wordpress"

curl -s -o "$WORDPRESS_HOME/#1.yaml" \
  "$CONTENT/{deployment,service,kustomization}.yaml"
```

下载 MySQL 的 resources 和 `kustomization.yaml` 。

<!-- @downloadResources @test -->
```bash
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress/mysql"

curl -s -o "$MYSQL_HOME/#1.yaml" \
  "$CONTENT/{deployment,service,secret,kustomization}.yaml"
```

### 创建 kustomization.yaml

基于 `wordpress` 和 `mysql` 的两个 bases 创建一个新的 `kustomization.yaml` ：

<!-- @createKustomization @test -->
```bash
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- wordpress
- mysql
namePrefix: demo-
patchesStrategicMerge:
- patch.yaml
EOF
```

### 下载 WordPress 的 patchs

在新的 kustomization 中应用 WordPress Deployment 的 patch ，该 patch 包含：
- 添加初始容器来显示mysql的服务名称
- 添加允许 wordpress 查找到 mysql 数据库的环境变量

<!-- @downloadPatch @test -->
```bash
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress"

curl -s -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/{patch}.yaml"
```
该 patch 内容如下：
> ```yaml
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   name: wordpress
> spec:
>   template:
>     spec:
>       initContainers:
>       - name: init-command
>         image: debian
>         command: ["/bin/sh"]
>         args: ["-c", "echo $(WORDPRESS_SERVICE); echo $(MYSQL_SERVICE)"]
>       containers:
>       - name: wordpress
>         env:
>         - name: WORDPRESS_DB_HOST
>           value: $(MYSQL_SERVICE)
>         - name: WORDPRESS_DB_PASSWORD
>           valueFrom:
>             secretKeyRef:
>               name: mysql-pass
>               key: password
> ```
初始化容器的命令需要依赖于k8s资源对象字段的信息，由占位符变量 $(WORDPRESS_SERVICE) 和 $(MYSQL_SERVICE) 表示。

### 将变量绑定到k8s对象字段

<!-- @addVarRef @test -->
```bash
cat <<EOF >>$DEMO_HOME/kustomization.yaml
vars:
  - name: WORDPRESS_SERVICE
    objref:
      kind: Service
      name: wordpress
      apiVersion: v1
    fieldref:
      fieldpath: metadata.name
  - name: MYSQL_SERVICE
    objref:
      kind: Service
      name: mysql
      apiVersion: v1
EOF
```
`WORDPRESS_SERVICE` 来自 `wordpress` 服务的 `metadata.name` 字段。如果不指定 `fieldref` ，则使用默认的 `metadata.name` 。因此 `MYSQL_SERVICE` 来自 `mysql` 服务的 `metadata.name` 字段。

### 替换

运行命令查看替换结果：

<!-- @kustomizeBuild @test -->
```bash
kustomize build $DEMO_HOME
```

预期的输出为：

> ```yaml
> (truncated)
> ...
>     initContainers:
>     - args:
>       - -c
>       - echo demo-wordpress; echo demo-mysql
>       command:
>       - /bin/sh
>       image: debian
>       name: init-command
>
> ```
