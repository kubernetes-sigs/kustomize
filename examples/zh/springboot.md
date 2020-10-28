# 示例：SpringBoot

在本教程中，您将学会如何使用 `kustomize` 定制一个运行 Spring Boot 应用的 k8s 配置。

在生产环境中，我们需要定制如下内容：

- 为 Spring Boot 应用添加特定配置
- 配置数据库连接
- 以 'prod-' 前缀命名资源
- 资源具有 'env: prod' label
- 设置合适的 JVM 内存
- 健康检查和就绪检查

首先创建一个工作空间：
<!-- @makeDemoHome @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### 下载资源

为了保证文档的精简，基础资源都在补充目录中，如果需要请下载它们：
<!-- @downloadResources @testAgainstLatestRelease -->
```
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/springboot"

curl -s -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/base/{deployment,service}.yaml"
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

kustomize edit add resource service.yaml
kustomize edit add resource deployment.yaml

cat kustomization.yaml
```

执行上面的命令后，`kustomization.yaml` 的 resources 字段如下：

> ```
> resources:
> - service.yaml
> - deployment.yaml
> ```

### 添加 configMap 生成器

<!-- @addConfigMap @testAgainstLatestRelease -->
```
echo "app.name=Kustomize Demo" >$DEMO_HOME/application.properties

kustomize edit add configmap demo-configmap \
  --from-file application.properties

cat kustomization.yaml
```

执行上面的命令后，`kustomization.yaml` 的 configMapGenerator 字段如下：

> ```
> configMapGenerator:
> - files:
>   - application.properties
>   name: demo-configmap
> ```

### 定制 configMap

我们将为生产环境添加数据库连接凭证。通常这些凭据被存放在 `application.properties` 中，然而在有些时候，我们希望将这些凭证保存在其他文件中，而将应用的其他配置保存在 `application.properties` 中。通过这种清晰的分离，这些凭证和应用配置可由不同的团队管理和维护。例如，应用开发人员可以在 `application.properties` 中调整应用程序的配置，而数据库的连接凭证则由运维或 SRE 团队管理和维护。

对于 Spring Boot 应用，我们可以通过环境变量动态的设置 `spring.profiles.active`，然后应用将获取一个额外的 `application-<profile>.properties` 文件，我们可以分为两步定制这个 ConfigMap：

1. 通过 patch 添加一个环境变量
2. 将文件添加到 ConfigMap 中

<!-- @customizeConfigMap -->
```
cat <<EOF >$DEMO_HOME/patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sbdemo
spec:
  template:
    spec:
      containers:
        - name: sbdemo
          env:
          - name: spring.profiles.active
            value: prod
EOF

kustomize edit add patch --path patch.yaml --name sbdemo --kind Deployment --group apps --version v1

cat <<EOF >$DEMO_HOME/application-prod.properties
spring.jpa.hibernate.ddl-auto=update
spring.datasource.url=jdbc:mysql://<prod_database_host>:3306/db_example
spring.datasource.username=root
spring.datasource.password=admin
EOF

kustomize edit add configmap \
  demo-configmap --from-file application-prod.properties

cat kustomization.yaml
```

执行上面的命令后，`kustomization.yaml` 的 configMapGenerator 字段如下：

> ```
> configMapGenerator:
> - files:
>   - application.properties
>   - application-prod.properties
>   name: demo-configmap
> ```

### 定制名称

为资源添加 _prod-_ 前缀（这些资源将用于生产环境）：

<!-- @customizeLabel @testAgainstLatestRelease -->
```
cd $DEMO_HOME
kustomize edit set nameprefix 'prod-'
```

执行上面的命令后，`kustomization.yaml` 的 namePrefix 字段将会被更新：

> ```
> namePrefix: prod-
> ```

`namePrefix` 将在所有资源的名称前添加 _prod-_ 的前缀，可以通过如下命令查看：

<!-- @build1 @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME | grep prod-
```

### 定制 Label

我们希望生产环境的资源包含某些 Label，这样我们就可以通过 label selector 来查询到这些资源。

`kustomize` 没有 `edit set label` 命令来添加 label，但是可以通过编辑 `kustomization.yaml` 文件来实现：

<!-- @customizeLabels @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
commonLabels:
  env: prod
EOF
```

现在所有资源都包含 `prod-` 前缀和 `env:prod` label，可以通过下面的命令来查看：

<!-- @build2 @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME | grep -C 3 env
```

### 下载调整 JVM 内存的 Patch

当 Spring Boot 应用部署在 k8s 集群中时，JVM 会运行在容器中。我们要为容器设置内存限制，并确保 JVM 知道容器的内存限制。在 k8s 的 Deployment 中，我们可以设置资源容器的资源限制，并将限制注入到一些环境变量中，当容器启动时，其可以获取环境变量并设置相应的 JVM 选项。

下载 `memorylimit_patch.yaml` 其包含内存限制设置的 patch：

<!-- @downloadPatch @testAgainstLatestRelease -->
```
curl -s  -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/overlays/production/{memorylimit_patch}.yaml"

cat $DEMO_HOME/memorylimit_patch.yaml
```

输出内容

> ```
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   name: sbdemo
> spec:
>   template:
>     spec:
>       containers:
>         - name: sbdemo
>           resources:
>             limits:
>               memory: 1250Mi
>             requests:
>               memory: 1250Mi
>           env:
>           - name: MEM_TOTAL_MB
>             valueFrom:
>               resourceFieldRef:
>                 resource: limits.memory
> ```

### 下载健康检查的 Patch

我们还可以在生产环境中添加健康检查和就绪检查，Spring Boot 应用都具有类似 `/actuator/health` 的接口用于健康检查，我们可以定制 k8s 的 Deployment 资源来进行健康检查和就绪检查。

下载 `memorylimit_patch.yaml` 其包含存活和就绪探针的 patch：

<!-- @downloadPatch @testAgainstLatestRelease -->
```
curl -s  -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/overlays/production/{healthcheck_patch}.yaml"

cat $DEMO_HOME/healthcheck_patch.yaml
```

输出内容

> ```
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   name: sbdemo
> spec:
>   template:
>     spec:
>       containers:
>         - name: sbdemo
>           livenessProbe:
>             httpGet:
>               path: /actuator/health
>               port: 8080
>             initialDelaySeconds: 10
>             periodSeconds: 3
>           readinessProbe:
>             initialDelaySeconds: 20
>             periodSeconds: 10
>             httpGet:
>               path: /actuator/info
>               port: 8080
> ```

### 添加 patches

将这些 patch 添加到 `kustomization.yaml` 中：

<!-- @addPatch -->
```
cd $DEMO_HOME
kustomize edit add patch --path memorylimit_patch.yaml --name sbdemo --kind Deployment --group apps --version v1
kustomize edit add patch --path healthcheck_patch.yaml --name sbdemo --kind Deployment --group apps --version v1
```

执行上面的命令后，`kustomization.yaml` 的 patches 字段如下：

> ```
> patches:
> - path: patch.yaml
>   target:
>     group: apps
>     version: v1
>     kind: Deployment
>     name: sbdemo
> - path: memorylimit_patch.yaml
>   target:
>     group: apps
>     version: v1
>     kind: Deployment
>     name: sbdemo
> - path: healthcheck_patch.yaml
>   target:
>     group: apps
>     version: v1
>     kind: Deployment
>     name: sbdemo
> ```

现在就可以将完整的配置输出并在集群中部署（将结果通过管道输出给 `kubectl apply`），在生产环境创建 Spring Boot 应用。

<!-- @finalBuild @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME  # | kubectl apply -f -
```
