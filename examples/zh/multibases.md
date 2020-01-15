# 示例：多 Overlay 使用相同 base

`kustomize` 鼓励定义多个 variants：例如在通用的 base 上使用 dev、staging 和 prod overlay。

可以创建其他 overlay 来将这些 variants 组合在一起：只需将 overlay 声明为新 kustomization 的 base 即可。

如果 base 由于某种原因无法控制，将多个 variants 组合在一起也可以为他们添加通用的 label 或 annotation。

下面使用一个 pod 作为 base 来进行演示。

首先创建一个工作空间：

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

定义一个通用的 base：
<!-- @makeBase @testAgainstLatestRelease -->
```
BASE=$DEMO_HOME/base
mkdir $BASE

cat <<EOF >$BASE/kustomization.yaml
resources:
- pod.yaml
EOF

cat <<EOF >$BASE/pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
EOF
```

定义 dev variant：
<!-- @makeDev @testAgainstLatestRelease -->
```
DEV=$DEMO_HOME/dev
mkdir $DEV

cat <<EOF >$DEV/kustomization.yaml
resources:
- ./../base
namePrefix: dev-
EOF
```

定义 staging variant：
<!-- @makeStaging @testAgainstLatestRelease -->
```
STAG=$DEMO_HOME/staging
mkdir $STAG

cat <<EOF >$STAG/kustomization.yaml
resources:
- ./../base
namePrefix: stag-
EOF
```

定义 production variant：
<!-- @makeProd @testAgainstLatestRelease -->
```
PROD=$DEMO_HOME/production
mkdir $PROD

cat <<EOF >$PROD/kustomization.yaml
resources:
- ./../base
namePrefix: prod-
EOF
```

然后定义一个 _Kustomization_，将三个 variants 组合在一起：
<!-- @makeTopLayer @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- ./dev
- ./staging
- ./production

namePrefix: cluster-a-
EOF
```

现在工作空间有如下目录：
> ```
> .
> ├── base
> │   ├── kustomization.yaml
> │   └── pod.yaml
> ├── dev
> │   └── kustomization.yaml
> ├── kustomization.yaml
> ├── production
> │   └── kustomization.yaml
> └── staging
>     └── kustomization.yaml
> ```

输出包含三个 pod 对象，分别来自 dev、staging 和 production variants。

<!-- @confirmVariants @testAgainstLatestRelease -->
```
test 1 == \
  $(kustomize build $DEMO_HOME | grep cluster-a-dev-myapp-pod | wc -l); \
  echo $?
  
test 1 == \
  $(kustomize build $DEMO_HOME | grep cluster-a-stag-myapp-pod | wc -l); \
  echo $?
  
test 1 == \
  $(kustomize build $DEMO_HOME | grep cluster-a-prod-myapp-pod | wc -l); \
  echo $?    
```

与在不同的 variants 中添加不同的 `namePrefix` 类似，也可以添加不同的 `namespace` 并在一个  _kustomization_ 中组成这些 variants。更多的详细信息，请查看[multi-namespaces](multi-namespace.md)。
