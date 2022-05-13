# remote targets

## remote directories

`kustomize build` can be run on a URL. Resources can also reference other
kustomization directories via URLs too.

The URL format is an HTTPS or SSH `git clone` URL with an optional directory and
some query string parameters. Kustomize does not currently support ports in the
URL. The directory is specified by appending a `//` after the repo URL. The
following query string parameters can also be specified:

 * `ref` - a `git fetch`-able ref, typically a branch, tag, or full commit hash
   (short hashes are not supported)
 * `timeout` (default `27s`) - a number in seconds, or a go duration. specifies
   the timeout for fetching the resource
 * `submodules` (default `true`) - a boolean specifying whether to clone
   submodules or not

For example,
`https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`
will essentially clone the git repo via HTTPS, checkout `v1.0.6` and run
`kustomize build` inside the `examples/multibases/dev` directory.

SSH clones are also supported either with `git@github.com:owner/repo` or
`ssh://git@github.com/owner/repo` URLs.

`file:///` clones are not supported.

## remote files
Resources can reference remote files via their raw GitHub urls, such
as `https://raw.githubusercontent.com/kubernetes-sigs/kustomize/8ea501347443c7760217f2c1817c5c60934cf6a5/examples/helloWorld/deployment.yaml`
.

# Examples

To try this immediately, run a build against the kustomization
in the [multibases](multibases/README.md) example.  There's
one pod in the output:

<!-- @remoteOverlayBuild @testAgainstLatestRelease -->
```
target="https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"
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

A remote kustomization directory resource can also be a URL:

<!-- @createOverlay @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
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

## Legacy URL format

Historically, kustomize has supported a modified [hashicorp/go-getter URL format](https://github.com/hashicorp/go-getter#url-format).

This is still supported for backwards compatibility but is no longer recommended
to be used as kustomize supports different query parameters and the semantics of
what gets fetched in `go-getter` itself are different (particularly with
subdirectories).

Here are some examples of legacy urls

<!-- @createOverlay @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

cat <<'EOF' >$DEMO_HOME/kustomization.yaml
resources:
# a repo with a root level kustomization.yaml
- github.com/Liujingfang1/mysql
# a repo with a root level kustomization.yaml on branch test
- github.com/Liujingfang1/mysql?ref=test
# a subdirectory in a repo on branch repoUrl2
- github.com/Liujingfang1/kustomize/examples/helloWorld?ref=repoUrl2
# a subdirectory in a repo on commit `7050a45134e9848fca214ad7e7007e96e5042c03`
- github.com/Liujingfang1/kustomize/examples/helloWorld?ref=7050a45134e9848fca214ad7e7007e96e5042c03
EOF
```
