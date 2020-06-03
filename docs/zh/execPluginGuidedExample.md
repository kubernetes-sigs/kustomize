# 60 秒构建一个 Exec 插件

本教程只是一个快速开始的示例，完整的插件文档请看：[kustomize 插件](plugins.md)

本示例将使用 bash 编写了一个傻瓜 _exec_ 插件，用来生成一个 `ConfigMap`。

在不破坏当前的设置的情况下，尝试本教程。

#### 环境要求

 * `linux`
 * `git`
 * `curl`
 * `Go 1.13`


## 定义一个工作空间

```
DEMO=$(mktemp -d)
```

## 编写 kustomization

新建一个目录来保存所有的配置：

```
MYAPP=$DEMO/myapp
mkdir -p $MYAPP
```

编写一个 Deployment 配置：

```
cat <<'EOF' >$MYAPP/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: the-container
        image: monopole/hello:1
        command: ["/hello",
                  "--port=8080",
                  "--date=$(THE_DATE)",
                  "--enableRiskyFeature=$(ENABLE_RISKY)"]
        ports:
        - containerPort: 8080
        env:
        - name: THE_DATE
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: today
        - name: ALT_GREETING
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: altGreeting
        - name: ENABLE_RISKY
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: enableRisky
EOF
```

编写一个 service 配置：

```
cat <<EOF >$MYAPP/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: the-service
spec:
  type: LoadBalancer
  ports:
  - protocol: TCP
    port: 8666
    targetPort: 8080
EOF
```

现在为您要编写的插件创建一个配置文件。

这个配置文件的内容也是 k8s 资源对象。其中 `apiVersion` 和 `kind` 字段的值用于在文件系统中查找插件代码（稍后会对此进行更多介绍）。

```
cat <<'EOF' >$MYAPP/cmGenerator.yaml
apiVersion: myDevOpsTeam
kind: SillyConfigMapGenerator
metadata:
  name: whatever
argsOneLiner: Bienvenue true
EOF
```

最后在 kustomization 文件中引用以上所有内容：

```
cat <<EOF >$MYAPP/kustomization.yaml
commonLabels:
  app: hello
resources:
- deployment.yaml
- service.yaml
generators:
- cmGenerator.yaml
EOF
```

检查这些文件

```
ls -C1 $MYAPP
```

## 为插件创建目录

插件必须位于特定的目录，以便 Kustomize 能够找到它们。

该示例将使用临时目录：

```
PLUGIN_ROOT=$DEMO/kustomize/plugin
```

在上面定义的插件配置 `$MYAPP/cmGenerator.yaml` 中指定：

> ```
> apiVersion: myDevOpsTeam
> kind: SillyConfigMapGenerator
> ```

这意味着该插件必须位于以下目录中：

```
MY_PLUGIN_DIR=$PLUGIN_ROOT/myDevOpsTeam/sillyconfigmapgenerator

mkdir -p $MY_PLUGIN_DIR
```

插件的目录结构为： `apiVersion 的 value/小写 kind 的 value`。

插件拥有自己的目录不但可以保存插件代码，还可以保存在测试以及任何可能需要的补充数据文件。

## 编写插件

插件有 _exec_ 和 _Go_ 两种.

编写一个 _exec_ 插件，将其安装到正确的目录，文件名必须与插件的类型匹配（在本例中为 `SillyConfigMapGenerator`）：

```
cat <<'EOF' >$MY_PLUGIN_DIR/SillyConfigMapGenerator
#!/bin/bash
# Skip the config file name argument.
shift
today=`date +%F`
echo "
kind: ConfigMap
apiVersion: v1
metadata:
  name: the-map
data:
  today: $today
  altGreeting: "$1"
  enableRisky: "$2"
"
EOF
```

根据定义，_exec_ 插件必须是可执行的：

```
chmod a+x $MY_PLUGIN_DIR/SillyConfigMapGenerator
```

## 安装 kustomize

根据[文档](INSTALL.md)安装 kustomize:

```
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
mkdir -p $DEMO/bin
mv kustomize $DEMO/bin
```

## 检查这个目录

```
tree $DEMO
```

## 使用插件构建 APP:

```
XDG_CONFIG_HOME=$DEMO $DEMO/bin/kustomize build --enable_alpha_plugins $MYAPP
```

之前如果您已经设置了 `PLUGIN_ROOT=$HOME/.config/kustomize/plugin`，则无需在 _kustomize_ 命令前使用 `XDG_CONFIG_HOME`。
