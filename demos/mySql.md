# Demo: MySql

This example takes some off-the-shelf k8s resources
designed for MySQL, and customizes them to suit a
production scenario.

In the production environment we want:

- MySQL resource names to be prefixed by 'prod-'.
- MySQL resources to have 'env: prod' labels.
- MySQL to use persistent disk for storing data.

### Download resources

Download `deployment.yaml`, `service.yaml` and
`secret.yaml`.  These are plain k8s resources files one
could add to a k8s cluster to run MySql.

<!-- @makeMySQLDir @test -->
```
DEMO_HOME=$(mktemp -d)
cd $DEMO_HOME

CONTENT=https://raw.githubusercontent.com/kinflate

# Get MySQL configs
for f in service secret deployment ; do \
  wget -q $CONTENT/mysql/master/emptyDir/$f.yaml ; \
done
```


### Initialize kustomize.yaml

The `kustomize` program gets its instructions from
a file called `kustomize.yaml`.

Start this file:

<!-- @kustomizeYaml @test -->
```
touch $DEMO_HOME/kustomize.yaml
```

### Add the resources

<!-- @addResources @test -->
```
cd $DEMO_HOME

kustomize edit add resource secret.yaml
kustomize edit add resource service.yaml
kustomize edit add resource deployment.yaml

cat kustomize.yaml
```

`kustomize.yaml`'s resources section should contain:

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

<!-- @customizeLabel @test -->
```
cd $DEMO_HOME

kustomize edit set nameprefix 'prod-'

cat kustomize.yaml
```

`kustomize.yaml` should have updated value of namePrefix field:

> ```
> namePrefix: prod-
> ```

This `namePrefix` directive adds _prod-_ to all
resource names.

<!-- @genNamePrefixConfig @test -->
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

`kustomize` does not have `set label` command to add
label, but we can edit `kustomize.yaml` file under
`prod` directory and add the production labels under
`labelsToAdd` fields as highlighted below.

<!-- @customizeLabels @test -->
```
sed -i 's/app: helloworld/app: prod/' \
    $DEMO_HOME/kustomize.yaml
```

At this point, running `kustomize build` will
generate MySQL configs with name-prefix 'prod-' and
labels `env:prod`.

### Storage customization

Off the shelf MySQL uses `emptyDir` type volume, which
gets wiped away if the MySQL Pod is recreated, and that
is certainly not desirable for production
environment. So we want to use Persistent Disk in
production. kustomize lets you apply `patches` to the
resources.

<!-- @createPatchFile @test -->
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

Add the patch file to `kustomize.yaml`:

<!-- @specifyPatch @test -->
```
cat <<'EOF' >> $DEMO_HOME/kustomize.yaml
patches:
- persistent-disk.yaml
EOF
```

Lets break this down:

- In the first step, we created a YAML file named
  `persistent-disk.yaml` to patch the resource defined
  in deployment.yaml

- Then we added `persistent-disk.yaml` to list of
  `patches` in `kustomize.yaml`. `kustomize build`
  will apply this patch to the deployment resource with
  the name `mysql` as defined in the patch.


The output of the following command can now be applied
to the cluster (i.e. piped to `kubectl apply`) to
create the production environment.

<!-- @finalInflation @test -->
```
kustomize build $DEMO_HOME  # | kubectl apply -f -
```
