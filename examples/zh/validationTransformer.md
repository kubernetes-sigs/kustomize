# 通过 transformer 验证资源

[kubeval]: https://github.com/instrumenta/kubeval
[插件]: ../../docs/plugins

kustomize 不会验证其输入或输出是否符合资源要求。

而另一个工具 [kubeval] 提供了验证 k8s 资源的功能，例如：

```shell
$ kubeval my-invalid-rc.yaml
The document my-invalid-rc.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
```

可以创建一个 Kustomize transformer [插件] 通过运行 [kubeval] 来进行验证资源。

创建一个工作空间：

<!-- @makeWorkplace @test -->
```bash
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/valid
mkdir -p $DEMO_HOME/invalid
PLUGINDIR=$DEMO_HOME/kustomize/plugin/someteam.example.com/v1/validator
mkdir -p $PLUGINDIR
```

## 创建 transformer 插件

根据操作系统下载 [kubeval] 的二进制文件并将其添加到 $PATH。

<!-- @downloadKubeval @test -->
```bash
OS=`uname | sed -e 's/Linux/linux/' -e 's/Darwin/darwin/'`
wget https://github.com/instrumenta/kubeval/releases/download/0.9.2/kubeval-${OS}-amd64.tar.gz
tar xf kubeval-${OS}-amd64.tar.gz
export PATH=$PATH:`pwd`
```

transformer 插件将执行逻辑如下：

- 从 stdin 中读取资源并传递到 transformer 插件。
- transformer 插件的配置文件作为第一个参数传入。
- transformer 插件的工作目录是 kustomization 所在目录。
- 转换后的资源由插件写入 stdout 。
- transformer 返回值为0，则转化成功；如果 transformer 插件的返回值不为0，则 kustomize 认为转化期间存在错误。

我们可以写一个 bash 脚本作为用于验证资源的 transformer 插件，该脚本执行 [kubeval] 二进制文件并返回正确的输出和退出码。
<!-- @writePlugin @test -->
```bash
cat <<'EOF' > $PLUGINDIR/Validator
#!/bin/bash

if ! [ -x "$(command -v kubeval)" ]; then
  echo "Error: kubeval is not installed."
  exit 1
fi

temp_file=$(mktemp)
output_file=$(mktemp)
cat - > $temp_file

kubeval $temp_file > $output_file

if [ $? -eq 0 ]; then
    cat $temp_file
    rm $temp_file $output_file
    exit 0
fi

cat $output_file
rm $temp_file $output_file
exit 1

EOF
chmod +x $PLUGINDIR/Validator
```

## 使用 transformer 插件

创建一个包含有效 ConfigMap 和 transformer 插件的 Kustomization。

<!-- @writeKustomization @test -->
```bash
cat <<'EOF' >$DEMO_HOME/valid/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  foo: bar
EOF

cat <<'EOF' >$DEMO_HOME/valid/validation.yaml
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
EOF

cat <<'EOF' >$DEMO_HOME/valid/kustomization.yaml
resources:
- configmap.yaml

transformers:
- validation.yaml
EOF
```

创建一个包含无效 ConfigMap 和 transformer 插件的 Kustomization。

<!-- @writeKustomization @test -->
```bash
cat <<'EOF' >$DEMO_HOME/invalid/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
- foo: bar
EOF
# ConfigMap 的 data 字段需要传入的数据类型为 object，这里传入一个 array

cat <<'EOF' >$DEMO_HOME/invalid/validation.yaml
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
EOF

cat <<'EOF' >$DEMO_HOME/invalid/kustomization.yaml
resources:
- configmap.yaml

transformers:
- validation.yaml
EOF
```

目录结构如下：

```bash
/tmp/tmp.fAYMfLZJs4
├── invalid
│   ├── configmap.yaml
│   ├── kustomization.yaml
│   └── validation.yaml
├── kustomize
│   └── plugin
│       └── someteam.example.com
│           └── v1
│               ├── kubeval
│               └── Validator
└── valid
    ├── configmap.yaml
    ├── kustomization.yaml
    └── validation.yaml
```

定义一个 helper 函数在正确的的环境和插件标记运行 kustomize 。

<!-- @defineKustomizeBd @test -->
```bash
function kustomizeBd {
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build \
    --enable_alpha_plugins \
    $DEMO_HOME/$1
}
```

构建有效的 variant

<!-- @buildValid @test -->
```bash
kustomizeBd valid
```
输出的 ConfigMap 内容为：

```yaml
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: cm
```

构建无效的 variant

```bash
kustomizeBd invalid
```

可以查看到输出错误日志为：

```shell
data: Invalid type. Expected: object, given: array
```

## 清理

<!-- @cleanup @test -->
```shell
rm -rf $DEMO_HOME
```
