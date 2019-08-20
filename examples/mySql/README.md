# Demo: MySql

This example takes some off-the-shelf k8s resources
designed for MySQL, and customizes them to suit a
production scenario.

In the production environment we want:

- MySQL resource names to be prefixed by 'prod-'.
- MySQL resources to have 'env: prod' labels.
- MySQL to use persistent disk for storing data.

First make a place to work:
<!-- @makeDemoHome @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### Download resources

To keep this document shorter, the base resources
needed to run MySql on a k8s cluster are off in a
supplemental data directory rather than declared here
as HERE documents.

Download them:

<!-- @downloadResources @testAgainstLatestRelease -->
```
curl -s  -o "$DEMO_HOME/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/mySql\
/{deployment,secret,service}.yaml"
```

### Initialize kustomization.yaml

The `kustomize` program gets its instructions from
a file called `kustomization.yaml`.

Start this file:

<!-- @kustomizeYaml @testAgainstLatestRelease -->
```
touch $DEMO_HOME/kustomization.yaml
```

### Add the resources

<!-- @addResources @testAgainstLatestRelease -->
```
cd $DEMO_HOME

kustomize edit add resource secret.yaml
kustomize edit add resource service.yaml
kustomize edit add resource deployment.yaml

cat kustomization.yaml
```

`kustomization.yaml`'s resources section should contain:

> ```
> resources:
> - secret.yaml
> - service.yaml
> - deployment.yaml
> ```

### Name Customization

Arrange for the MySQL resources to begin with prefix
_prod-_ (since they are meant for the _production_
environment):

<!-- @customizeLabel @testAgainstLatestRelease -->
```
cd $DEMO_HOME

kustomize edit set nameprefix 'prod-'

cat kustomization.yaml
```

`kustomization.yaml` should have updated value of namePrefix field:

> ```
> namePrefix: prod-
> ```

This `namePrefix` directive adds _prod-_ to all
resource names.

<!-- @genNamePrefixConfig @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME
```

The output should contain:

> ```
> apiVersion: v1
> data:
>   password: YWRtaW4=
> kind: Secret
> metadata:
>   ....
>   name: prod-mysql-pass-d2gtcm2t2k
> ---
> apiVersion: v1
> kind: Service
> metadata:
>   ....
>   name: prod-mysql
> spec:
>   ....
> ---
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   ....
>   name: prod-mysql
> spec:
>   selector:
>     ....
> ```

### Label Customization

We want resources in production environment to have
certain labels so that we can query them by label
selector.

`kustomize` does not have `edit set label` command to add
a label, but one can always edit `kustomization.yaml` directly:

<!-- @customizeLabels @testAgainstLatestRelease -->
```
sed -i.bak 's/app: helloworld/app: prod/' \
    $DEMO_HOME/kustomization.yaml
```

At this point, running `kustomize build` will
generate MySQL configs with name-prefix 'prod-' and
labels `env:prod`.

### Storage customization

Off the shelf MySQL uses `emptyDir` type volume, which
gets wiped away if the MySQL Pod is recreated, and that
is certainly not desirable for production
environment. So we want to use Persistent Disk in
production. kustomize lets you apply `patchesStrategicMerge` to the
resources.

<!-- @createPatchFile @testAgainstLatestRelease -->
```
cat <<'EOF' > $DEMO_HOME/persistent-disk.yaml
apiVersion: apps/v1beta2 # for versions before 1.9.0 use apps/v1beta2
kind: Deployment
metadata:
  name: mysql
spec:
  template:
    spec:
      volumes:
      - name: mysql-persistent-storage
        emptyDir: null
        gcePersistentDisk:
          pdName: mysql-persistent-storage
EOF
```

Add the patch file to `kustomization.yaml`:

<!-- @specifyPatch @testAgainstLatestRelease -->
```
cat <<'EOF' >> $DEMO_HOME/kustomization.yaml
patchesStrategicMerge:
- persistent-disk.yaml
EOF
```

A `mysql-persistent-storage` persistent disk needs to exist for it to run successfully.

Lets break this down:

- In the first step, we created a YAML file named
  `persistent-disk.yaml` to patch the resource defined
  in deployment.yaml

- Then we added `persistent-disk.yaml` to list of
  `patchesStrategicMerge` in `kustomization.yaml`. `kustomize build`
  will apply this patch to the deployment resource with
  the name `mysql` as defined in the patch.


The output of the following command can now be applied
to the cluster (i.e. piped to `kubectl apply`) to
create the production environment.

<!-- @finalInflation @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME  # | kubectl apply -f -
```
