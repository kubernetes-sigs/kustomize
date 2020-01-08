# 示例：MySql

本示例采用现成的专为 MySql 设计的 k8s 资源，并对其进行定制使其适合生产环境。

在生产环境中，我们希望：

- 以 'prod-' 为前缀的 MySQL 资源
- MySQL 资源具有 'env: prod' label
- 使用持久化磁盘来存储 MySQL 数据

首先创建一个工作空间：
<!-- @makeDemoHome @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### 下载资源

为了保证文档的精简，基础资源都在补充目录中，如果需要请下载它们：

<!-- @downloadResources @testAgainstLatestRelease -->
```
curl -s  -o "$DEMO_HOME/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/mySql\
/{deployment,secret,service}.yaml"
```

### 初始化 kustomization.yaml

`kustomize` 会从 `kustomization.yaml` 文件中获取指令，创建这个文件：

<!-- @kustomizeYaml @testAgainstLatestRelease -->
```
touch $DEMO_HOME/kustomization.yaml
```

### 添加资源

<!-- @addResources @testAgainstLatestRelease -->
```
cd $DEMO_HOME

kustomize edit add resource secret.yaml
kustomize edit add resource service.yaml
kustomize edit add resource deployment.yaml

cat kustomization.yaml
```

执行上面的命令后，`kustomization.yaml` 的 resources 字段如下：

> ```
> resources:
> - secret.yaml
> - service.yaml
> - deployment.yaml
> ```

### 定制名称

为 MySQL 资源添加 _prod-_ 前缀（这些资源将用于生产环境）：

<!-- @customizeLabel @testAgainstLatestRelease -->
```
cd $DEMO_HOME

kustomize edit set nameprefix 'prod-'

cat kustomization.yaml
```

执行上面的命令后，`kustomization.yaml` 的 namePrefix 字段将会被更新：

> ```
> namePrefix: prod-
> ```

`namePrefix` 将在所有资源的名称前添加 _prod-_ 的前缀，可以通过如下命令查看：

<!-- @genNamePrefixConfig @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME
```

输出内容：

> ```
> apiVersion: v1
> data:
>   password: YWRtaW4=
> kind: Secret
> metadata:
>   ....
>   name: prod-mysql-pass-d2gtcm2t2k
> ---
> apiVersion: v1
> kind: Service
> metadata:
>   ....
>   name: prod-mysql
> spec:
>   ....
> ---
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   ....
>   name: prod-mysql
> spec:
>   selector:
>     ....
> ```

### 定制 Label

我们希望生产环境的资源包含某些 Label，这样我们就可以通过 label selector 来查询到这些资源。

`kustomize` 没有 `edit set label` 命令来添加 label，但是可以通过编辑 `kustomization.yaml` 文件来实现：

<!-- @customizeLabels @testAgainstLatestRelease -->
```
sed -i.bak 's/app: helloworld/app: prod/' \
    $DEMO_HOME/kustomization.yaml
```

这时，执行 `kustomize build` 命令将会生成包含 `prod-` 前缀和 `env:prod` label 的 MySQL 配置。

### 存储定制

现成的 MySQL 使用 `emptyDir` 类型的 volume，如果 MySQL Pod 被重新部署，则该类型的 volume 将会消失，这是不能应用于生产环境的，因此在生产环境中我们需要使用持久化磁盘。在 kustomize 中可以使用`patchesStrategicMerge` 来应用资源。

<!-- @createPatchFile @testAgainstLatestRelease -->
```
cat <<'EOF' > $DEMO_HOME/persistent-disk.yaml
apiVersion: apps/v1 # for versions before 1.9.0 use apps/v1beta2
kind: Deployment
metadata:
  name: mysql
spec:
  template:
    spec:
      volumes:
      - name: mysql-persistent-storage
        emptyDir: null
        gcePersistentDisk:
          pdName: mysql-persistent-storage
EOF
```

将 patch 文件添加到 `kustomization.yaml` 中：

<!-- @specifyPatch @testAgainstLatestRelease -->
```
cat <<'EOF' >> $DEMO_HOME/kustomization.yaml
patchesStrategicMerge:
- persistent-disk.yaml
EOF
```

`mysql-persistent-storage` 必须存在一个持久化磁盘才能使其成功运行，分为两步：

1. 创建一个名为 `persistent-disk.yaml` 的 YAML 文件，用于修改 deployment.yaml 的定义。
2. 在 `kustomization.yaml` 中添加 `persistent-disk.yaml` 到 `patchesStrategicMerge` 列表中。运行 `kustomize build` 将 patch 应用于 Deployment 资源。

现在就可以将完整的配置输出并在集群中部署（将结果通过管道输出给 `kubectl apply`），在生产环境创建MySQL 应用。

<!-- @finalInflation @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME  # | kubectl apply -f -
```
