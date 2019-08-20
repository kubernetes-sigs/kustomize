[patch]: ../docs/glossary.md#patch
[resource]: ../docs/glossary.md#resource
[variant]: ../docs/glossary.md#variant

## ConfigMap generation and rolling updates

Kustomize provides two ways of adding ConfigMap in one `kustomization`, either by declaring ConfigMap as a [resource] or declaring ConfigMap from a ConfigMapGenerator. The formats inside `kustomization.yaml` are 

> ```
> # declare ConfigMap as a resource
> resources:
> - configmap.yaml
> 
> # declare ConfigMap from a ConfigMapGenerator
> configMapGenerator:
> - name: a-configmap
>   files:
>     - configs/configfile
>     - configs/another_configfile
> ```

The ConfigMaps declared as [resource] are treated the same way as other resources. Kustomize doesn't append any hash to the ConfigMap name. The ConfigMap declared from a ConfigMapGenerator is treated differently. A hash is appended to the name and any change in the ConfigMap will trigger a rolling update.

In this demo, the same [hello_world](helloWorld/README.md) is used while the ConfigMap declared as [resources] is replaced by a ConfigMap declared from a ConfigmapGenerator. The change in this ConfigMap will result in a hash change and a rolling update.

### Establish base and staging

Establish the base with a configMapGenerator
<!-- @establishBase @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

BASE=$DEMO_HOME/base
mkdir -p $BASE

curl -s -o "$BASE/#1.yaml" "https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/helloWorld\
/{deployment,service}.yaml"

cat <<'EOF' >$BASE/kustomization.yaml
commonLabels:
  app: hello
resources:
- deployment.yaml
- service.yaml
configMapGenerator:	
- name: the-map	
  literals:	
    - altGreeting=Good Morning!	
    - enableRisky="false"
EOF
```

Establish the staging with a patch applied to the ConfigMap
<!-- @establishStaging @testAgainstLatestRelease -->
```
OVERLAYS=$DEMO_HOME/overlays
mkdir -p $OVERLAYS/staging

cat <<'EOF' >$OVERLAYS/staging/kustomization.yaml
namePrefix: staging-
nameSuffix: -v1
commonLabels:
  variant: staging
  org: acmeCorporation
commonAnnotations:
  note: Hello, I am staging!
resources:
- ../../base
patchesStrategicMerge:
- map.yaml
EOF

cat <<EOF >$OVERLAYS/staging/map.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Have a pineapple!"
  enableRisky: "true"
EOF
```

### Review

The _hello-world_ deployment running in this cluster is
configured with data from a configMap.

The deployment refers to this map by name:


<!-- @showDeployment @testAgainstLatestRelease -->
```
grep -C 2 configMapKeyRef $BASE/deployment.yaml
```

Changing the data held by a live configMap in a cluster
is considered bad practice. Deployments have no means
to know that the configMaps they refer to have
changed, so such updates have no effect.

The recommended way to change a deployment's
configuration is to

 1. create a new configMap with a new name,
 1. patch the _deployment_, modifying the name value of
    the appropriate `configMapKeyRef` field.

This latter change initiates rolling update to the pods
in the deployment.  The older configMap, when no longer
referenced by any other resource, is eventually [garbage
collected](https://github.com/kubernetes-sigs/kustomize/issues/242).

### How this works with kustomize

The _staging_ [variant] here has a configMap [patch]:

<!-- @showMapPatch @testAgainstLatestRelease -->
```
cat $OVERLAYS/staging/map.yaml
```

This patch is by definition a named but not necessarily
complete resource spec intended to modify a complete
resource spec.

The ConfigMap it modifies is declared from a configMapGenerator.

<!-- @showMapBase @testAgainstLatestRelease -->
```
grep -C 4 configMapGenerator $BASE/kustomization.yaml
```

For a patch to work, the names in the `metadata/name`
fields must match.

However, the name values specified in the file are
_not_ what gets used in the cluster.  By design,
kustomize modifies names of ConfigMaps declared from ConfigMapGenerator.  To see the names
ultimately used in the cluster, just run kustomize:

<!-- @grepStagingName @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging |\
    grep -B 8 -A 1 staging-the-map
```

The configMap name is prefixed by _staging-_, per the
`namePrefix` field in
`$OVERLAYS/staging/kustomization.yaml`.

The configMap name is suffixed by _-v1_, per the
`nameSuffix` field in
`$OVERLAYS/staging/kustomization.yaml`.

The suffix to the configMap name is generated from a
hash of the maps content - in this case the name suffix
is _k25m8k5k5m_:

<!-- @grepStagingHash @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging | grep k25m8k5k5m
```

Now modify the map patch, to change the greeting
the server will use:

<!-- @changeMap @testAgainstLatestRelease -->
```
sed -i.bak 's/pineapple/kiwi/' $OVERLAYS/staging/map.yaml
```

See the new greeting:

```
kustomize build $OVERLAYS/staging |\
  grep -B 2 -A 3 kiwi
```

Run kustomize again to see the new configMap names:

<!-- @grepStagingName @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging |\
    grep -B 8 -A 1 staging-the-map
```

Confirm that the change in configMap content resulted
in three new names ending in _cd7kdh48fd_ - one in the
configMap name itself, and two in the deployment that
uses the map:

<!-- @countHashes @testAgainstLatestRelease -->
```
test 3 == \
  $(kustomize build $OVERLAYS/staging | grep cd7kdh48fd | wc -l); \
  echo $?
```

Applying these resources to the cluster will result in
a rolling update of the deployments pods, retargetting
them from the _k25m8k5k5m_ maps to the _cd7kdh48fd_
maps.  The system will later garbage collect the
unused maps.

## Rollback

To rollback, one would undo whatever edits were made to
the configuation in source control, then rerun kustomize
on the reverted configuration and apply it to the
cluster.
