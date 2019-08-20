# remote targets

`kustomize build` can be run on a URL.

The effect is the same as cloning the repo, checking out a particular
_ref_ (commit hash, branch name, release tag, etc.),
then running `kustomize build` against the desired
directory in the local copy.

To try this immediately, run a build against the kustomization
in the [multibases](multibases/README.md) example.  There's
one pod in the output:

<!-- @remoteOverlayBuild @testAgainstLatestRelease -->

```
target="github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"
test 1 == \
  $(kustomize build $target | grep dev-myapp-pod | wc -l); \
  echo $?
```

Run against the overlay in that example to get three pods
(the overlay combines the dev, staging and prod bases for
someone who wants to send them all at the same time):

<!-- @remoteBuild @testAgainstLatestRelease -->
```
target="https://github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6"
test 3 == \
  $(kustomize build $target | grep cluster-a-.*-myapp-pod | wc -l); \
  echo $?
```

A base can be a URL:

<!-- @createOverlay @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
namePrefix: remote-
EOF
```

Build this to confirm that all three pods from the base
have the `remote-` prefix.

<!-- @remoteBases @testAgainstLatestRelease -->
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
