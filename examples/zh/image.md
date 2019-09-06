# 示例: 改变镜像名称和标签

首先构建一个工作空间：

<!-- @makeWorkplace @testAgainstLatestRelease -->
```bash
DEMO_HOME=$(mktemp -d)
```

创建包含pod资源的 `kustomization`

<!-- @testAgainstLatestRelease to @test -->
```bash
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- pod.yaml
EOF
```

创建 pod 资源pod.yaml

<!-- @createDeployment @test -->
```bash
cat <<EOF >$DEMO_HOME/pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox:1.29.0
    command: ['sh', '-c', 'echo The app is running! && sleep 3600']
  initContainers:
  - name: init-mydb
    image: busybox:1.29.0
    command: ['sh', '-c', 'until nslookup mydb; do echo waiting for mydb; sleep 2; done;']
EOF
```

`myapp-pod` 包含一个init容器和一个普通容器，两者都使用 `busybox：1.29.0` 镜像。

在 `kustomization.yaml` 中添加 `images` 字段来更改镜像 `busybox` 和标签 `1.29.0` 。

- 通过 `kustomize` 添加 `images`：
    <!-- @addImages @test -->
    ```bash
    cd $DEMO_HOME
    kustomize edit set image busybox=alpine:3.6
    ```

- 将`images`字段将被添加到`kustomization.yaml`：
    > ```yaml
    > images:
    > - name: busybox
    >   newName: alpine
    >   newTag: 3.6
    > ```

构建 `kustomization`
<!-- @kustomizeBuild @testAgainstLatestRelease -->
```bash
kustomize build $DEMO_HOME
```

确认`busybox`镜像和标签是否被替换为`alpine：3.6`：
<!-- @confirmImages @testAgainstLatestRelease -->
```
test 2 = \
  $(kustomize build $DEMO_HOME | grep alpine:3.6 | wc -l); \
  echo $?
```
