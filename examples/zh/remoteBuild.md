# remote targets

`kustomize build` 可以将 URL 作为参数传入并运行.

运行效果与如下操作相同:

如果想要要立即尝试此操作，可以按照 [multibases](../multibases/README.md) 示例运行 kustomization 运行构建。然后查看输出中的pod：

<!-- @remoteOverlayBuild @test -->

```bash
target="https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"
test 1 == \
  $(kustomize build $target | grep dev-myapp-pod | wc -l); \
  echo $?
```

在该示例中运行 overlay 将获得三个 pod（在此 overlay 结合了dev、staging 和 prod 的 bases，以便同时将它们全部发送给所有人）：

<!-- @remoteBuild @test -->
```bash
target="https://github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6"
test 3 == \
  $(kustomize build $target | grep cluster-a-.*-myapp-pod | wc -l); \
  echo $?
```

将 URL 作为 base ：

<!-- @createOverlay @test -->
```bash
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
namePrefix: remote-
EOF
```

构建该 base 以确定所有的三个 pod 都有 `remote-` 前缀。

<!-- @remoteBases @testAgainstLatestRelease -->
```bash
test 3 == \
  $(kustomize build $DEMO_HOME | grep remote-.*-myapp-pod | wc -l); \
  echo $?
```

## Legacy URL format

URL 需要遵循 [hashicorp/go-getter URL 格式](https://github.com/hashicorp/go-getter#url-format) 。下面是一些遵循此约定的 Github repos 示例url。

- kustomization.yaml 在根目录

  `github.com/Liujingfang1/mysql`
- kustomization.yaml 在 test 分支的根目录

  `github.com/Liujingfang1/mysql?ref=test`
- kustomization.yaml 在 v1.0.6 版本的子目录

  `github.com/kubernetes-sigs/kustomize/examples/multibases?ref=v1.0.6`
- kustomization.yaml repoUrl2 分支的子目录

  `github.com/Liujingfang1/kustomize/examples/helloWorld?ref=repoUrl2`
- kustomization.yaml commit `7050a45134e9848fca214ad7e7007e96e5042c03` 的子目录

  `github.com/Liujingfang1/kustomize/examples/helloWorld?ref=7050a45134e9848fca214ad7e7007e96e5042c03`
