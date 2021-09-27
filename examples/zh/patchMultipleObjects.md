[Strategic Merge Patch]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md
[JSON patches]: https://tools.ietf.org/html/rfc6902
[label selector]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors


# 示例：通过一个 patch 来修改多个资源

kustomization.yaml 支持通过 [Strategic Merge Patch] 和 [JSON patch] 来自定义资源。自 3.1.0 起，一个 patch 可以修改多个资源。

这可以通过指定 patch 和它所修改的 target 来完成，如下所示：
```yaml
patches:
- path: <PatchFile>
  target:
    group: <Group>
    version: <Version>
    kind: <Kind>
    name: <Name>
    namespace: <Namespace>
    labelSelector: <LabelSelector>
    annotationSelector: <AnnotationSelector>
```
`labelSelector` 和 `annotationSelector` 都应遵循 [label selector] 中的约定。Kustomize 选择匹配`target`中所有字段的目标来应用 patch 。

下面的示例展示了如何为所有部署资源注入 sidecar 容器。

创建一个包含 Deployment 资源的 `kustomization` 。

<!-- @createDeployment @test -->
```bash
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- deployments.yaml
EOF

cat <<EOF >$DEMO_HOME/deployments.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
        - name: nginx
          image: nginx
          args:
          - one
          - two
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    metadata:
      labels:
        key: value
    spec:
      containers:
        - name: busybox
          image: busybox
EOF
```

声明 [Strategic Merge Patch] 文件以注入 sidecar 容器：

<!-- @addPatch @test -->
```bash
cat <<EOF >$DEMO_HOME/patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-important
spec:
  template:
    spec:
      containers:
        - name: istio-proxy
          image: docker.io/istio/proxyv2
          args:
          - proxy
          - sidecar
EOF
```

在 kustomization.yaml 中添加 _patches_ 字段

<!-- @applyPatch @testAgainstLatestRelease -->
```bash
cat <<EOF >>$DEMO_HOME/kustomization.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
EOF
```

运行 `kustomize build $DEMO_HOME`，可以在输出中确认两个 Deployment 资源都已正确应用。

<!-- @confirmPatch @test -->
```bash
test 2 == \
  $(kustomize build $DEMO_HOME | grep "image: docker.io/istio/proxyv2" | wc -l); \
  echo $?
```

输出如下：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
      - args:
        - one
        - two
        image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    metadata:
      labels:
        key: value
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
      - image: busybox
        name: busybox
```

## Target 选择

- 选择名称与 `name*` 匹配的资源
  ```yaml
  target:
    name: name*
  ```
- 选择所有 Deployment 资源
  ```yaml
  target:
    kind: Deployment
  ```
- 选择 label 与 `app=hello` 匹配的资源
  ```yaml
  target:
    labelSelector: app=hello
  ```
- 选择 annotation 与 `app=hello` 匹配的资源
  ```yaml
  target:
    annotationSelector: app=hello
  ```
- 选择所有 label 与 `app=hello` 匹配的 Deployment 资源
  ```yaml
  target:
    kind: Deployment
    labelSelector: app=hello
  ```