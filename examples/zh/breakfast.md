[kubernetes API 对象样式]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields
[variant]: ../../docs/glossary.md#variant

# 示例：早餐配置

定义一个工作空间：

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

创建目录用于存放早餐的 base 配置：

<!-- @baseDir @testAgainstLatestRelease -->
```
mkdir -p $DEMO_HOME/breakfast/base
```

创建一个 `kustomization` 来定义早餐所需的食物。包含咖啡和薄煎饼：

<!-- @baseKustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/breakfast/base/kustomization.yaml
resources:
- coffee.yaml
- pancakes.yaml
EOF
```

这里有一个 _coffee_ 类型。定义`kind`和 `metdata/name` 字段以符合 [kubernetes API 对象样式]，不需要其他文件或定义：

<!-- @coffee @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/breakfast/base/coffee.yaml
kind: Coffee
metadata:
  name: morningCup
temperature: lukewarm
data:
  greeting: "Good Morning!"
EOF
```

`name` 字段仅将这种咖啡实例与其他实例（如果有的话）区分开

同样，定义 _pancakes_：
<!-- @pancakes @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/breakfast/base/pancakes.yaml
kind: Pancakes
metadata:
  name: comfort
stacksize: 3
topping: none
EOF
```

为喜欢热咖啡的 Alice 定制她的早餐：

<!-- @aliceOverlay @testAgainstLatestRelease -->
```
mkdir -p $DEMO_HOME/breakfast/overlays/alice

cat <<EOF >$DEMO_HOME/breakfast/overlays/alice/kustomization.yaml
commonLabels:
  who: alice
resources:
- ../../base
patchesStrategicMerge:
- temperature.yaml
EOF

cat <<EOF >$DEMO_HOME/breakfast/overlays/alice/temperature.yaml
kind: Coffee
metadata:
  name: morningCup
temperature: hot!
EOF
```

同样的，Bob 想要 _5_ 块薄煎饼和草莓：

<!-- @bobOverlay @testAgainstLatestRelease -->
```
mkdir -p $DEMO_HOME/breakfast/overlays/bob

cat <<EOF >$DEMO_HOME/breakfast/overlays/bob/kustomization.yaml
commonLabels:
  who: bob
resources:
- ../../base
patchesStrategicMerge:
- topping.yaml
EOF

cat <<EOF >$DEMO_HOME/breakfast/overlays/bob/topping.yaml
kind: Pancakes
metadata:
  name: comfort
stacksize: 5
topping: strawberries
EOF
```

现在，可以为 Alice 的早餐生成配置了：

<!-- @generateAlice @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME/breakfast/overlays/alice
```

同样的，也为 Bob  的早餐生成配置：

<!-- @generateBob @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME/breakfast/overlays/bob
```
