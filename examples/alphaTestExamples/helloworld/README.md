[base]: ../../docs/glossary.md#base
[config]: https://github.com/kinflate/example-hello
[gitops]: ../../docs/glossary.md#gitops
[hello]: https://github.com/monopole/hello
[kustomization]: ../../docs/glossary.md#kustomization
[original]: https://github.com/kinflate/example-hello
[overlay]: ../../docs/glossary.md#overlay
[overlays]: ../../docs/glossary.md#overlay
[patch]: ../../docs/glossary.md#patch
[variant]: ../../docs/glossary.md#variant
[variants]: ../../docs/glossary.md#variant

# Demo: hello world

Steps:

 1. Clone an existing configuration as a [base].
 1. Customize it.

First define a place to work:

<!-- @makeWorkplace @testE2EAgainstLatestRelease-->
```
DEMO_HOME=$(mktemp -d)
```

Alternatively, use

> ```
> DEMO_HOME=~/hello
> ```

## Establish the base

Let's run the [hello] service.

To keep this document shorter, the base resources are
off in a supplemental data directory rather than
declared here as HERE documents.  Download them:

<!-- @downloadBase @testE2EAgainstLatestRelease-->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE

curl -s -o "$BASE/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/alphaTestExamples/helloWorld\
/{configMap,deployment,grouping,kustomization,service}.yaml"
```

### The Base Kustomization

The `base` directory has a [kustomization] file:

<!-- @showKustomization @testE2EAgainstLatestRelease -->
```
more $BASE/kustomization.yaml
```

### Customize the base

A first customization step could be to change the _app
label_ applied to all resources:

<!-- @addLabel @testE2EAgainstLatestRelease -->
```
sed -i.bak 's/app: hello/app: my-hello/' \
    $BASE/kustomization.yaml
```

To do end to end tests using kustomize, go through the following section. You should have GOPATH set up and "kind" installed(https://github.com/kubernetes-sigs/kind).

<!-- @setGoBin @testE2EAgainstLatestRelease -->
```
MYGOBIN=$GOPATH/bin
```

Delete any existing kind cluster and create a new one. By default the name of the cluster is "kind"
<!-- @deleteAndCreateKindCluster @testE2EAgainstLatestRelease -->
```
kind delete cluster;
kind create cluster;
```

Use the kustomize binary in MYGOBIN to apply a deployment, fetch the status and verify the status.
<!-- @e2eTestUsingKustomize @testE2EAgainstLatestRelease -->
```
export KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true

$MYGOBIN/kustomize resources apply $BASE --status;

status=$(mktemp);
$MYGOBIN/resource status fetch $BASE > $status

test 1 == \
  $(grep "the-deployment" $status | grep "Deployment is available. Replicas: 3" | wc -l); \
  echo $?

test 1 == \
  $(grep "the-map" $status | grep "Resource is always ready" | wc -l); \
  echo $?

test 1 == \
  $(grep "the-service" $status | grep "Service is ready" | wc -l); \
  echo $?
```

Clean-up the cluster 
<!-- @createKindCluster @testE2EAgainstLatestRelease -->
```
kind delete cluster;
```