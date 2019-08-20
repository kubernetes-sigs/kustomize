[patch]: ../../docs/glossary.md#patch
[resource]: ../../docs/glossary.md#resource
[variant]: ../../docs/glossary.md#variant

## ConfigMap 的生成和滚动更新

kustomize 提供了两种添加 ConfigMap 的方法：
- 将 ConfigMap 声明为 [resource]
- 通过 ConfigMapGenerator 声明 ConfigMap

在 `kustomization.yaml` 中，这两种方法的格式分别如下：

> ```
> # 将 ConfigMap 声明为 resource
> resources:
> - configmap.yaml
> 
> # 在 ConfigMapGenerator 中声明 ConfigMap
> configMapGenerator:
> - name: a-configmap
>   files:
>     - configs/configfile
>     - configs/another_configfile
> ```

声明为 [resource] 的 ConfigMaps 的处理方式与其他 resource 相同，Kustomize 不会在为 ConfigMap 的名称添加哈希后缀。而在 ConfigMapGenerator 中声明 ConfigMap 的处理方式则与之前不同，默认将为名称添加哈希后缀，ConfigMap 中的任何更改都将触发滚动更新。

在 [hello_world](helloWorld.md) 示例中，使用 ConfigmapGenerator 来替换将 ConfigMap 声明为 [resource] 的方法。由此生成的 ConfigMap 中的更改将导致哈希值更改和滚动更新。

### 建立 base 和 staging

使用 configMapGenerator 建立 base
<!-- @establishBase @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

BASE=$DEMO_HOME/base
mkdir -p $BASE

curl -s -o "$BASE/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/helloWorld\
/{deployment,service}.yaml"

cat <<'EOF' >$BASE/kustomization.yaml
commonLabels:
  app: hello
resources:
- deployment.yaml
- service.yaml
configMapGenerator:	
- name: the-map	
  literals:	
    - altGreeting=Good Morning!	
    - enableRisky="false"
EOF
```

通过应用 ConfigMap patch 的方式建立 staging
<!-- @establishStaging @testAgainstLatestRelease -->
```
OVERLAYS=$DEMO_HOME/overlays
mkdir -p $OVERLAYS/staging

cat <<'EOF' >$OVERLAYS/staging/kustomization.yaml
namePrefix: staging-
nameSuffix: -v1
commonLabels:
  variant: staging
  org: acmeCorporation
commonAnnotations:
  note: Hello, I am staging!
resources:
- ../../base
patchesStrategicMerge:
- map.yaml
EOF

cat <<EOF >$OVERLAYS/staging/map.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Have a pineapple!"
  enableRisky: "true"
EOF
```

### Review

在集群中运行的 _hello-world_ 的 deployment 配置了来自 configMap 的数据。

deployment 按照名称引用此 ConfigMap ：

<!-- @showDeployment @testAgainstLatestRelease -->
```
grep -C 2 configMapKeyRef $BASE/deployment.yaml
```

当 ConfigMap 中的数据需要更新时，更改群集中的实时 ConfigMap 的数据并不是一个好的做法。 由于 Deployment 无法知道其引用的 ConfigMap 已更改，这类更新是无效。

更改 Deployment 配置的推荐方法是：

 1. 使用新名称创建一个新的 configMap
 2. 为_deployment_ 添加 patch，修改相应 `configMapKeyRef` 字段的名称值。

后一种更改会启动对 deployment 中的 pod 的滚动更新。旧的 configMap 在不再被任何其他资源引用时最终会被[垃圾回收](https://github.com/kubernetes-sigs/kustomize/issues/242)。

### 如何使用 kustomize 

_staging_ 的 [variant] 包含一个 configMap 的 [patch]：

<!-- @showMapPatch @testAgainstLatestRelease -->
```
cat $OVERLAYS/staging/map.yaml
```

根据定义，此 patch 是一个命名但不一定是完整的资源规范，旨在修改完整的资源规范。

在 ConfigMapGenerator 中声明 ConfigMap 的修改。

<!-- @showMapBase @testAgainstLatestRelease -->
```
grep -C 4 configMapGenerator $BASE/kustomization.yaml
```

要使这个 patch 正常工作，`metadata/name` 字段中的名称必须匹配。

但是，文件中指定的名称值不是群集中使用的名称值。根据设计，kustomize 修改从 ConfigMapGenerator 声明的 ConfigMaps 的名称。要查看最终在群集中使用的名称，只需运行 kustomize：

<!-- @grepStagingName @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging |\
    grep -B 8 -A 1 staging-the-map
```

根据 `$OVERLAYS/staging/kustomization.yaml` 中的 `namePrefix` 字段，configMap 名称以 _staging-_ 为前缀。

根据 `$OVERLAYS/staging/kustomization.yaml` 中的 `nameSuffix` 字段，configMap 名称以 _-v1_ 为后缀。

configMap 名称的后缀是由 map 内容的哈希生成的 - 在这种情况下，名称后缀是 _k25m8k5k5m_ ：

<!-- @grepStagingHash @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging | grep k25m8k5k5m
```

现在修改 map patch ，更改该服务将使用的问候消息：

<!-- @changeMap @testAgainstLatestRelease -->
```
sed -i.bak 's/pineapple/kiwi/' $OVERLAYS/staging/map.yaml
```

查看新的问候消息：

```
kustomize build $OVERLAYS/staging |\
  grep -B 2 -A 3 kiwi
```

再次运行 kustomize 查看新的 configMap 名称：

<!-- @grepStagingName @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging |\
    grep -B 8 -A 1 staging-the-map
```

确认 configMap 内容的更改将会生成以 _cd7kdh48fd_ 结尾的三个新名称 - 一个在 configMap 的名称中，另两个在使用 ConfigMap 的 deployment 中：

<!-- @countHashes @testAgainstLatestRelease -->
```
test 3 == \
  $(kustomize build $OVERLAYS/staging | grep cd7kdh48fd | wc -l); \
  echo $?
```

将这些资源应用于群集将导致 deployment pod 的滚动更新，将它们从 _k25m8k5k5m_ map 重新定位到 _cd7kdh48fd_ map 。系统稍后将垃圾收集未使用的 map。

## 回滚

回滚，可以撤消对源码配置所做的任何编辑，然后在还原的配置上重新运行 kustomize 并将其应用于群集。
