---
title: "Go 插件示例"
linkTitle: "Go 插件示例"
type: docs
weight: 4
description: >
    Go 插件示例
---

[SopsEncodedSecrets repository]: https://github.com/monopole/sopsencodedsecrets
[Go plugin]: https://golang.org/pkg/plugin
[Go plugin caveats]: ../goplugincaveats

本教程只是一个快速开始的示例，完整的插件文档请看：[kustomize 插件](..)

请务必阅读 [Go 插件注意事项](../goplugincaveats)。

该示例使用 Go 插件 `SopsEncodedSecrets`，该插件位于 [sopsencodedsecrets repository]中。这是一个进程内的 Go 插件，而不是恰巧用 Go 编写的 exec 插件（这是 Go 作者的另一种选择）。

尝试本教程不会破坏你的当前设置。

#### 环境要求

* `linux`
* `git`
* `curl`
* `Go 1.13`

用于加密

* gpg

或

* Google cloud (gcloud) 安装
* 具有 KMS 权限的 Google帐户

## 创建一个工作空间/目录

```shell
# 将这些目录分开，以免造成 DEMO 目录的混乱。
DEMO=$(mktemp -d)
tmpGoPath=$(mktemp -d)
```

## 安装 kustomize

需要安装 kustomize v3.0.0，并且必须对其进行 _编译_（而不是从 release 页面下载二进制文件）：

```shell
GOPATH=$tmpGoPath go install sigs.k8s.io/kustomize/kustomize
```

## 为插件创建目录

kustomize 插件完全由其配置文件和源代码确定。

Kustomize 插件的配置文件的格式与 kubernetes 资源对象相同，这就意味着在配置文件中 `apiVersion`，`kind` 和 `metadata` 都是[必须的字段](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields)。

因为配置文件名出现在 kustomization 文件的 `generatorsor` 或 `transformers` 字段中，kustomize 会读取配置文件，然后在以下位置找到 Go 插件的目标代码：

> ```shell
> $XDG_CONFIG_HOME/kustomize/plugin/$apiVersion/$lKind/$kind.so
> ```

`lKind` 必须是小写字母的，然后将插件加载并提供其配置，插件的输出将成为整个 `kustomize build` 程序的一部分 。

同一插件在一个 kustomize 构建中可能会多次使用不同的配置文件。此外，kustomize 可能会先自定义 config 数据，然后再发送给插件。由于这些原因，插件不能自己去读取配置文件，而需要通过 kustomize 来读取配置。

该示例将在如下临时目录中存放其使用的插件：

```shell
PLUGIN_ROOT=$DEMO/kustomize/plugin
```

并在下面的命令行中临时设置 `XDG_CONFIG_HOME`。

### 使用什么 apiVersion 和 kind

在 kustomize 插件的开发时，插件代码不关心也不知道配置文件中的 `apiVersion` 或 `kind`。

插件会检查这些字段，但是剩下的字段提供了实际的配置数据，在这一点上，成功解析其他字段对于插件很重要。

本示例使用一个名为 _SopsEncodedSecrets_ 的插件，其位于 [SopsEncodedSecrets repository] 中。

我们选择安装插件到

```shell
apiVersion=mygenerators
kind=SopsEncodedSecrets
```

### 定义插件的主目录

按照惯例，存放插件代码和补充数据，测试，文档等的目录名称必须是 kind 的小写形式。

```shell
lKind=$(echo $kind | awk '{print tolower($0)}')
```

### 下载 SopsEncodedSecrets 插件

在这种情况下，存储库名称已经与小写字母的 kind 匹配，因此我们只需克隆存储库并自动获取正确的目录名称即可：

```shell
mkdir -p $PLUGIN_ROOT/${apiVersion}
cd $PLUGIN_ROOT/${apiVersion}
git clone git@github.com:monopole/sopsencodedsecrets.git
```

记住这个目录：

```shell
MY_PLUGIN_DIR=$PLUGIN_ROOT/${apiVersion}/${lKind}
```

### 尝试测试插件

插件可能会自己带有测试文件。因此可以通过如下方式：

```shell
cd $MY_PLUGIN_DIR
go test SopsEncodedSecrets_test.go
```

构建对象代码以供 kustomize 使用：

```shell
cd $MY_PLUGIN_DIR
GOPATH=$tmpGoPath go build -buildmode plugin -o ${kind}.so ${kind}.go
```

此步骤可能会成功，但是由于依赖关系 [skew]，kustomize 最终可能无法加载该插件。

[skew]: /docs/plugins/README.md#caveats

在加载失败时

* 确保使用相同版本的Go (_go1.13_)，在相同的 `$GOOS`(_linux_)和 `$GOARCH`(_amd64_) 上构建插件，用于构建本演示中使用的 [kustomize](#安装-kustomize)。

* 修改插件中的依赖文件 `go.mod` 以匹配 kustomize 使用的版本。

缺乏工具和元数据来实现自动化，就不会有一个完整的 Go 插件生态。

Kustomize 采用了 Go 插件架构，可以轻松的接受新的生成器和转换器（只需编写一个插件），并确保本机操作（也已作为插件构建和测试）是分段的、可排序的和可重用的，而不是奇怪的插入在整体代码中。

## 编写 kustomization

新建一个 kustomization 目录存放你的配置：

```shell
MYAPP=$DEMO/myapp
mkdir -p $MYAPP
```

为 SopsEncodedSecrets 插件编写一个配置文件。

插件可以通过 `apiVersion` 和 `kind` 找到：

```shell
cat <<EOF >$MYAPP/secGenerator.yaml
apiVersion: ${apiVersion}
kind: ${kind}
metadata:
  name: mySecretGenerator
name: forbiddenValues
namespace: production
file: myEncryptedData.yaml
keys:
- ROCKET
- CAR
EOF
```

插件可以在 `myEncryptedData.yaml` 中找到更多的数据。

编写一个引用插件配置的 kustomization 文件：

```shell
cat <<EOF >$MYAPP/kustomization.yaml
commonLabels:
  app: hello
generators:
- secGenerator.yaml
EOF
```

接下来生成真实的加密数据。

### 确保您已安装加密工具

我们将使用 [sops](https://github.com/mozilla/sops) 对文件进行编码。选择 GPG 或 Google Cloud KMS 作为加密提供者以继续。

#### GPG

尝试这个命令：

```shell
gpg --list-keys
```

如果返回 list，则您已经成功创建了密钥。如果不是，请尝试从 sops 导入测试密钥。

```shell
curl https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc | gpg --import
SOPS_PGP_FP="1022470DE3F0BC54BC6AB62DE05550BC07FB1A0A"
```

#### Google Cloude KMS

尝试这个命令：

```shell
gcloud kms keys list --location global --keyring sops
```

如果成功了，想必你已经创建了密钥，并将其放置在一个名为 sops 的钥匙圈中。如果没有，那就这样做：

```shell
gcloud kms keyrings create sops --location global
gcloud kms keys create sops-key --location global \
    --keyring sops --purpose encryption
```

通过如下方法，获取你的 keyLocation：

```shell
keyLocation=$(\
    gcloud kms keys list --location global --keyring sops |\
    grep GOOGLE | cut -d " " -f1)
echo $keyLocation
```

### 安装 `sops`

```shell
GOPATH=$tmpGoPath go install go.mozilla.org/sops/cmd/sops
```

### 用你的私钥创建加密数据

创建需要加密的原始数据：

```shell
cat <<EOF >$MYAPP/myClearData.yaml
VEGETABLE: carrot
ROCKET: saturn-v
FRUIT: apple
CAR: dymaxion
EOF
```

将数据加密插入到插件要读取的文件中：

使用 PGP

```shell
$tmpGoPath/bin/sops --encrypt \
  --pgp $SOPS_PGP_FP \
  $MYAPP/myClearData.yaml >$MYAPP/myEncryptedData.yaml
```

或者使用 GCP KMS

```shell
$tmpGoPath/bin/sops --encrypt \
  --gcp-kms $keyLocation \
  $MYAPP/myClearData.yaml >$MYAPP/myEncryptedData.yaml
```

查看文件

```shell
tree $DEMO
```

结果如下：

> ```shell
> /tmp/tmp.0kIE9VclPt
> ├── kustomize
> │   └── plugin
> │       └── mygenerators
> │           └── sopsencodedsecrets
> │               ├── go.mod
> │               ├── go.sum
> │               ├── LICENSE
> │               ├── README.md
> │               ├── SopsEncodedSecrets.go
> │               ├── SopsEncodedSecrets.so
> │               └── SopsEncodedSecrets_test.go
> └── myapp
>     ├── kustomization.yaml
>     ├── myClearData.yaml
>     ├── myEncryptedData.yaml
>     └── secGenerator.yaml
> ```

## 使用插件构建您的应用

```shell
XDG_CONFIG_HOME=$DEMO $tmpGoPath/bin/kustomize build --enable_alpha_plugins $MYAPP
```

这将生成一个 kubernetes secret，并对名称 `ROCKET` 和 `CAR` 的数据进行加密。

之前如果您已经设置了 `PLUGIN_ROOT=$HOME/.config/kustomize/plugin`，则无需在 _kustomize_ 命令前使用 `XDG_CONFIG_HOME`。
