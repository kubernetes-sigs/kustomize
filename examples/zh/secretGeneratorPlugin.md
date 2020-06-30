[ConfigMaps]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#configmap-v1-core
[ELF]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format
[Go plugin]: https://golang.org/pkg/plugin
[Secrets]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#secret-v1-core
[base64]: https://tools.ietf.org/html/rfc4648#section-4
[configuration directory]: https://wiki.archlinux.org/index.php/XDG_Base_Directory#Specification
[grpc]: https://grpc.io
[tag]: /../../releases
[v2.0.3]: /../../releases/tag/v2.0.3
[`exec.Command`]: https://golang.org/pkg/os/exec/#Command

# 生成 Secrets

## Secret 是什么？

Kubernetes 的 [ConfigMaps] 和 [Secrets] 都是key:value map，但 [Secrets] 的内容更为敏感，比如：密码或者 ssh 秘钥。

Kubernetes 开发者以各种方式工作，Secrets 保存的信息相比 ConfigMaps,Deployments 等的配置信息需要更谨慎的隐藏。

## 创建一个工作空间

<!-- @establishBase @test -->
```bash
DEMO_HOME=$(mktemp -d)
```

## 来自本地文件的 Secret

kustomize 可以通过三种不同的方式生成来自本地文件的 Secret 。

 * 从 _env_ 文件中获取（`NAME = VALUE`，每行一个）
 * 使用文件内容来生成一个 secret
 * 从 kustomization.yaml 文件获取 secret

这里有一个示例结合所有的三种方式：

创建一个包含一些短密码的 env 文件：

<!-- @makeEnvFile @test -->
```bash
cat <<'EOF' >$DEMO_HOME/foo.env
ROUTER_PASSWORD=admin
DB_PASSWORD=iloveyou
EOF
```

创建一个长密码的文本文件：

<!-- @makeLongSecretFile @test -->
```bash
cat <<'EOF' >$DEMO_HOME/longsecret.txt
Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
sed do eiusmod tempor incididunt
ut labore et dolore magna aliqua.
EOF
```

创建一个kustomization.yaml 文件, 其中包含引用上面文件的 secretGenerator, 并且另外定义一些文字 KV 对：

<!-- @makeKustomization1 @test -->
```bash
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
secretGenerator:
- name: mysecrets
  envs:
  - foo.env
  files:
  - longsecret.txt
  literals:
  - FRUIT=apple
  - VEGETABLE=carrot
EOF
```

生成 Secret ：

<!-- @build1 @test -->
```bash
result=$(kustomize build $DEMO_HOME)
echo "$result"
# Spot check the result:
test 1 == $(echo "$result" | grep -c "FRUIT: YXBwbGU=")
```

将会得到类似的内容：

> ```yaml
> apiVersion: v1
> kind: Secret
> metadata:
>   name: mysecrets-hfb5df789h
> type: Opaque
> data:
>   FRUIT: YXBwbGU=
>   VEGETABLE: Y2Fycm90
>   ROUTER_PASSWORD: YWRtaW4=
>   DB_PASSWORD: aWxvdmV5b3U=
>   longsecret.txt: TG9yZW0gaXBzdW0gZG9sb3Igc2l0I... (elided)
> ```

资源名称的前缀为 `mysecrets`（在 kustomization.yaml 中指定），后跟其内容的哈希值。

使用 base64 解码器确认这些值的原始版本。

这三种方法共同的问题是创建 Secret 所使用的敏感数据必须保存磁盘上。

这会增加额外的安全问题：对本地存储的敏感文件的查看、安装和删除权限的控制等。

## 来自任何地方的 Secret

一般的替代方案是在[generator](../../docs/plugins)中生成 secrets 。

然后，这些值可以通过经过身份验证和授权的 RPC 进入密码保险库服务。

[sgp]: ../../plugin/someteam.example.com/v1/secretsfromdatabase

这里有一个[secret 生成器][sgp]，它模拟从数据库中拉取 map 中的值。

下载

<!-- @copyPlugin @test -->
```bash
repo=https://raw.githubusercontent.com/kubernetes-sigs/kustomize
pPath=plugin/someteam.example.com/v1/secretsfromdatabase
dir=$DEMO_HOME/kustomize/$pPath

mkdir -p $dir

curl -s -o $dir/SecretsFromDatabase.go \
  ${repo}/master/$pPath/SecretsFromDatabase.go
```

运行 kustomize build 生成结果

<!-- @compilePlugin @xtest -->
```bash
go build -buildmode plugin \
  -o $dir/SecretsFromDatabase.so \
  $dir/SecretsFromDatabase.go
```

创建一个配置文件：

<!-- @makeConfiguration @test -->
```bash
cat <<'EOF' >$DEMO_HOME/secretFromDb.yaml
apiVersion: someteam.example.com/v1
kind: SecretsFromDatabase
metadata:
  name: mySecretGenerator
name: forbiddenValues
namespace: production
keys:
- ROCKET
- VEGETABLE
EOF
```

创建一个引用此生成器的新 kustomization.yaml 文件：

<!-- @makeKustomization2 @test -->
```bash
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
generators:
- secretFromDb.yaml
EOF
```

最终生成 secret ，设置 `XDG_CONFIG_HOME` 以便可以在 `$DEMO_HOME` 中找到该生成器：

<!-- @build2 @xtest -->
```bash
result=$( \
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build --enable_alpha_plugins $DEMO_HOME )
echo "$result"
# Spot check the result:
test 1 == $(echo "$result" | grep -c "FRUIT: YXBwbGU=")
```

将会得到类似的内容：

> ```yaml
> apiVersion: v1
> kind: Secret
> metadata:
>   name: mysecrets-bdt27dbkd6
> type: Opaque
> data:
>  FRUIT: YXBwbGU=
>  VEGETABLE: Y2Fycm90
> ```
