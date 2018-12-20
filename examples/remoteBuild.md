# remote targets

`kustomize build` can be run against a url. The effect is the same as cloing the repo, checking out the specified ref,
then running `kustomize build` against the desired directory in the local copy.

Take `github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6` as an example.
According to [multibases](multibases/README.md) demo, this kustomization contains three Pod objects with names as
`cluster-a-dev-myapp-pod`, `cluster-a-stag-myapp-pod`, `cluster-a-prod-myapp-pod`.
Running `kustomize build` against the url gives the same output.

<!-- @remoteBuild @test -->
```
target=github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
test 3 == \
  $(kustomize build $target | grep cluster-a-.*-myapp-pod | wc -l); \
  echo $?
```

Overlays can be remote as well:

<!-- @remoteOverlayBuild @test -->

```
target=github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6
test 1 == \
  $(kustomize build $target | grep cluster-a-dev-myapp-pod | wc -l); \
  echo $?
```

A base can also be specified as a URL:

<!-- @createOverlay @test -->
```
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
bases:
- github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
namePrefix: remote-
EOF
```
Running `kustomize build $DEMO_HOME` and confirm the output contains three Pods and all have `remote-` prefix.
<!-- @remoteBases @test -->
```
test 3 == \
  $(kustomize build $DEMO_HOME | grep remote-.*-myapp-pod | wc -l); \
  echo $?
```

## URL format
The url should follow
[hashicorp/go-getter URL format](https://github.com/hashicorp/go-getter#url-format).
Here are some example urls pointing to Github repos following this convention.

- a repo with a root level kustomization.yaml

  `github.com/Liujingfang1/mysql`
- a repo with a root level kustomization.yaml on branch test

  `github.com/Liujingfang1/mysql?ref=test`
- a subdirectory in a repo on version v1.0.6

  `github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6`
- a subdirectory in a repo on branch repoUrl2

  `github.com/Liujingfang1/kustomize//examples/helloWorld?ref=repoUrl2`
- a subdirectory in a repo on commit `7050a45134e9848fca214ad7e7007e96e5042c03`

  `github.com/Liujingfang1/kustomize//examples/helloWorld?ref=7050a45134e9848fca214ad7e7007e96e5042c03`
