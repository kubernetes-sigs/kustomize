[kind]: https://github.com/kubernetes-sigs/kind

# Demo: Multiple Deployments

This demo helps you to multiple services on same kubenetes cluster using kustomize.

Steps:
1. Create the resources files for wordpress service.
2. Create the resources files for mysql service.
3. Spin-up kubernetes cluster on local using [kind].
4. Deploy the wordpress app using kustomize and verify the status.
5. Deploy the mysql app using kustomize on same "kind" cluster and verify the status.
6. Add and remove a resource to mysql service and verify prune.

First define a place to work:

<!-- @makeWorkplace @testE2EAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Alternatively, use

> ```
> DEMO_HOME=~/hello
> ```

## Establish the base

<!-- @createBase @testE2EAgainstLatestRelease -->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE

mkdir $BASE/wordpress
mkdir $BASE/mysql

curl -s -o "$BASE/wordpress/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress/wordpress\
/{deployment,kustomization,service}.yaml"

curl -s -o "$BASE/mysql/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress/mysql\
/{secret,deployment,kustomization,service}.yaml"
```

Create a `grouping.yaml` resource. By this, you are defining the grouping of the current directory, `mysql`. Kustomize uses the unique label in this file to track any future state changes made to this directory. Make sure the label key is `kustomize.config.k8s.io/inventory-id` and give any unique label value and DO NOT change it in future.
<!-- @createGroupingYaml @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/mysql/grouping.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: inventory-map
  labels:
    kustomize.config.k8s.io/inventory-id: mysql-app
EOF
```

<!-- @setGoBin @testE2EAgainstLatestRelease -->
```
MYGOBIN=$GOPATH/bin
```

Delete any existing kind cluster and create a new one. By default the name of the cluster is "kind"
<!-- @deleteAndCreateKindCluster @testE2EAgainstLatestRelease -->
```
$MYGOBIN/kind delete cluster;
$MYGOBIN/kind create cluster;
```

Let's run the wordpress and mysql services.
<!-- @RunWordpressAndMysql @testE2EAgainstLatestRelease -->
```
export KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true

$MYGOBIN/kustomize resources apply $BASE/mysql --status;

status=$(mktemp);
$MYGOBIN/resource status fetch $BASE/mysql > $status

test 1 == \
  $(grep "mysql" $status | grep "Deployment is available. Replicas: 1" | wc -l); \
  echo $?

test 1 == \
  $(grep "mysql-pass" $status | grep "Resource is always ready" | wc -l); \
  echo $?

test 1 == \
  $(grep "mysql" $status | grep "Service is ready" | wc -l); \
  echo $?

$MYGOBIN/kustomize resources apply $BASE/wordpress --status;

status=$(mktemp);
$MYGOBIN/resource status fetch $BASE/wordpress > $status

test 1 == \
  $(grep "wordpress" $status | grep "Deployment is available. Replicas: 1" | wc -l); \
  echo $?

test 1 == \
  $(grep "wordpress" $status | grep "Service is ready" | wc -l); \
  echo $?
```

Let's replace the secret resource from mysql service and verify prune and addition of resource.
<!-- @ReplaceResourceInMysql @testE2EAgainstLatestRelease -->

```
cat <<EOF >$BASE/mysql/secret2.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysql-pass2
type: Opaque
data:
  # Default password is "admin".
  password: YWRtaW5=
EOF

rm $BASE/mysql/secret.yaml

sed -i.bak 's/secret/secret2/' \
    $BASE/mysql/kustomization.yaml

sed -i.bak 's/mysql-pass/mysql-pass2/' \
    $BASE/mysql/deployment.yaml

$MYGOBIN/kustomize resources apply $BASE/mysql --status;

status=$(mktemp);
$MYGOBIN/resource status fetch $BASE/mysql > $status

test 1 == \
  $(grep "mysql" $status | grep "Deployment is available. Replicas: 1" | wc -l); \
  echo $?

test 1 == \
  $(grep "mysql-pass2" $status | grep "Resource is always ready" | wc -l); \
  echo $?

test 1 == \
  $(grep "mysql" $status | grep "Service is ready" | wc -l); \
  echo $?
```

Clean-up the cluster 
<!-- @deleteKindCluster @testE2EAgainstLatestRelease -->
```
$MYGOBIN/kind delete cluster;
```