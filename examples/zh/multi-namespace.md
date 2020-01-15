# 示例：使用通用的 base 应用多 namespace

`kustomize` 支持基于同一base具有不同 namespace 的多个 variants。

只需将 overlay 作为新的 kustomization 的 base，就可以创建一个额外的 overlay 将这些 variants 组合在一起。下面使用一个 pod 作为 base 来进行演示。

创建一个工作空间：

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

定义 namespace-a 的 variant：
<!-- @makeNamespaceA @testAgainstLatestRelease -->
```
NSA=$DEMO_HOME/namespace-a
mkdir $NSA

cat <<EOF >$NSA/kustomization.yaml
resources:
- namespace.yaml
- ../base
namespace: namespace-a
EOF

cat <<EOF >$NSA/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: namespace-a
EOF
```

定义 namespace-b 的 variant：
<!-- @makeNamespaceB @testAgainstLatestRelease -->
```
NSB=$DEMO_HOME/namespace-b
mkdir $NSB

cat <<EOF >$NSB/kustomization.yaml
resources:
- namespace.yaml
- ../base
namespace: namespace-b
EOF

cat <<EOF >$NSB/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: namespace-b
EOF
```

然后定义一个 _Kustomization_，将两个 variants 组合在一起：
<!-- @makeTopLayer @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- namespace-a
- namespace-b
EOF
```

现在工作空间有如下目录：
> ```
> .
> ├── base
> │   ├── kustomization.yaml
> │   └── pod.yaml
> ├── kustomization.yaml
> ├── namespace-a
> │   ├── kustomization.yaml
> │   └── namespace.yaml
> └── namespace-b
>     ├── kustomization.yaml
>     └── namespace.yaml
> ```

输出两个 namespace 的 pod 对象，分别在 namespace-a 和 namespace-b。

<!-- @confirmVariants @testAgainstLatestRelease -->
```
test 2 == \
  $(kustomize build $DEMO_HOME| grep -B 4 "namespace: namespace-[ab]" | grep "name: myapp-pod" | wc -l); \
  echo $?  
```
