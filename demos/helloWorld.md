[base]: ../docs/glossary.md#base
[config]: https://github.com/kinflate/example-hello
[gitops]: ../docs/glossary.md#gitops
[hello]: https://github.com/monopole/hello
[instance]: ../docs/glossary.md#instance
[instances]: ../docs/glossary.md#instance
[kustomization]: ../docs/glossary.md#kustomization
[original]: https://github.com/kinflate/example-hello
[overlay]: ../docs/glossary.md#overlay
[overlays]: ../docs/glossary.md#overlay

# Demo: hello world with instances

Steps:

 1. Clone an existing configuration as a [base].
 1. Customize it.
 1. Create two different [overlays] (_staging_ and _production_)
    from the customized base.
 1. Run kustomize and kubectl to deploy staging and production.

First define a place to work:

<!-- @makeWorkplace @test -->
```
DEMO_HOME=$(mktemp -d)
```

Alternatively, use

> ```
> DEMO_HOME=~/hello
> ```

## Clone an example

Let's run the [hello] service.

We'll first need a [base] configuration for it -
the resource files we'll build on with overlays.

To keep this document shorter, we'll copy them in
(rather than declare them as HERE documents):

<!-- @downloadBase @test -->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE

exRepo=https://raw.githubusercontent.com/kubernetes/kubectl
exDir=master/cmd/kustomize/demos/data/helloWorld

curl -s "$exRepo/$exDir/{configMap,deployment,kustomization,service}.yaml" \
    -o "$BASE/#1.yaml"
```

Look at the directory:

<!-- @runTree @test -->
```
tree $DEMO_HOME
```

Expect something like:

> ```
> /tmp/tmp.IyYQQlHaJP
> └── base
>     ├── configMap.yaml
>     ├── deployment.yaml
>     ├── kustomization.yaml
>     ├── LICENSE
>     ├── README.md
>     └── service.yaml
> ```


One could immediately apply these resources to a
cluster:

> ```
> kubectl apply -f $DEMO_HOME/base
> ```

to instantiate the _hello_ service.  `kubectl`
would only recognize the resource files.

## The Base Kustomization

The `base` directory has a [kustomization] file:

<!-- @showKustomization @test -->
```
more $BASE/kustomization.yaml
```

Run `kustomize` on the base to emit customized resources
to `stdout`:

<!-- @buildBase @test -->
```
kustomize build $BASE
```

## Customize the base

A first customization step could be to change the _app
label_ applied to all resources:

<!-- @addLabel @test -->
```
sed -i 's/app: hello/app: my-hello/' \
    $BASE/kustomization.yaml
```

See the effect:
<!-- @checkLabel @test -->
```
kustomize build $BASE | grep -C 3 app:
```

## Create Overlays

Create a _staging_ and _production_ [overlay]:

 * _Staging_ enables a risky feature not enabled in production.
 * _Production_ has a higher replica count.
 * Web server greetings from these cluster
   [instances] will differ from each other.

<!-- @overlayDirectories @test -->
```
OVERLAYS=$DEMO_HOME/overlays
mkdir -p $OVERLAYS/staging
mkdir -p $OVERLAYS/production
```

#### Staging Kustomization

In the `staging` directory, make a kustomization
defining a new name prefix, and some different labels.

<!-- @makeStagingKustomization @test -->
```
cat <<'EOF' >$OVERLAYS/staging/kustomization.yaml
namePrefix: staging-
commonLabels:
  instance: staging
  org: acmeCorporation
commonAnnotations:
  note: Hello, I am staging!
bases:
- ../../base
patches:
- map.yaml
EOF
```

#### Staging Patch

Add a configMap customization to change the server
greeting from _Good Morning!_ to _Have a pineapple!_

Also, enable the _risky_ flag.

<!-- @stagingMap @test -->
```
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

#### Production Kustomization

In the production directory, make a kustomization
with a different name prefix and labels.

<!-- @makeProductionKustomization @test -->
```
cat <<EOF >$OVERLAYS/production/kustomization.yaml
namePrefix: production-
commonLabels:
  instance: production
  org: acmeCorporation
commonAnnotations:
  note: Hello, I am production!
bases:
- ../../base
patches:
- deployment.yaml
EOF
```


#### Production Patch

Make a production patch that increases the replica
count (because production takes more traffic).

<!-- @productionDeployment @test -->
```
cat <<EOF >$OVERLAYS/production/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 10
EOF
```

## Compare overlays


`DEMO_HOME` now contains:

 - a _base_ directory - a slightly customized clone
   of the original configuration, and

 - an _overlays_ directory, containing the kustomizations
   and patches required to create distinct _staging_
   and _production_ instances in a cluster.

Review the directory structure and differences:

<!-- @listFiles @test -->
```
tree $DEMO_HOME
```

Expecting something like:

> ```
> /tmp/tmp.IyYQQlHaJP1
> ├── base
> │   ├── configMap.yaml
> │   ├── deployment.yaml
> │   ├── kustomization.yaml
> │   ├── LICENSE
> │   ├── README.md
> │   └── service.yaml
> └── overlays
>     ├── production
>     │   ├── deployment.yaml
>     │   └── kustomization.yaml
>     └── staging
>         ├── kustomization.yaml
>         └── map.yaml
> ```

Compare the output directly
to see how _staging_ and _production_ differ:

<!-- @compareOutput -->
```
diff \
  <(kustomize build $OVERLAYS/staging) \
  <(kustomize build $OVERLAYS/production) |\
  more
```

The first part of the difference output should look
something like

> ```diff
> <   altGreeting: Have a pineapple!
> <   enableRisky: "true"
> ---
> >   altGreeting: Good Morning!
> >   enableRisky: "false"
> 8c8
> <     note: Hello, I am staging!
> ---
> >     note: Hello, I am production!
> 11c11
> <     instance: staging
> ---
> >     instance: production
> 13c13
> (...truncated)
> ```


## Deploy

The individual resource sets are:

<!-- @buildStaging @test -->
```
kustomize build $OVERLAYS/staging
```

<!-- @buildProduction @test -->
```
kustomize build $OVERLAYS/production
```

To deploy, pipe the above commands to kubectl apply:

> ```
> kustomize build $OVERLAYS/staging |\
>     kubectl apply -f -
> ```

> ```
> kustomize build $OVERLAYS/production |\
>    kubectl apply -f -
> ```

## Rolling updates

### Review

The _hello-world_ deployment running in this cluster is
configured with data from a configMap.

The deployment refers to this map by name:


<!-- @showDeployment @test -->
```
grep -C 2 configMapKeyRef $DEMO_HOME/base/deployment.yaml
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
referenced by any other resource, is eventually garbage
collected.

### How this works with kustomize

[patch]: ../docs/glossary.md#patch

The _staging_ instance here has a configMap [patch]:

<!-- @showMapPatch @test -->
```
cat $OVERLAYS/staging/map.yaml
```

This patch is by definition a named but not necessarily
complete resource spec intended to modify a complete
resource spec.

The resource it modifies is here:

<!-- @showMapBase @test -->
```
cat $DEMO_HOME/base/configMap.yaml
```

For a patch to work, the names in the `metadata/name`
fields must match.

However, the name values specified in the file are
_not_ what gets used in the cluster.  By design,
kustomize modifies these names.  To see the names
ultimately used in the cluster, just run kustomize:

<!-- @grepStagingName @test -->
```
kustomize build $OVERLAYS/staging |\
    grep -B 8 -A 1 staging-the-map
```

The configMap name is prefixed by _staging-_, per the
`namePrefix` field in
`$OVERLAYS/staging/kustomization.yaml`.

The suffix to the configMap name is generated from a
hash of the maps content - in this case the name suffix
is _hhhhkfmgmk_:

<!-- @grepStagingHash @test -->
```
kustomize build $OVERLAYS/staging | grep hhhhkfmgmk
```

Now modify the map patch, to change the greeting
the server will use:

<!-- @changeMap @test -->
```
sed -i 's/pineapple/kiwi/' $OVERLAYS/staging/map.yaml
```

Run kustomize again to see the new names:

<!-- @grepStagingName @test -->
```
kustomize build $OVERLAYS/staging |\
    grep -B 8 -A 1 staging-the-map
```

Confirm that the change in configMap content resulted
in three new names ending in _khk45ktkd9_ - one in the
configMap name itself, and two in the deployment that
uses the map:

<!-- @countHashes @test -->
```
test 3 == $(kustomize build $OVERLAYS/staging | grep khk45ktkd9 | wc -l)
```

Applying these resources to the cluster will result in
a rolling update of the deployments pods, retargetting
them from the _hhhhkfmgmk_ maps to the _khk45ktkd9_
maps.  The system will later garbage collect the
unused maps.

## Rollback

To rollback, one would undo whatever edits were made to
the configuation in source control, then rerun kustomize
on the reverted configuration and apply it to the
cluster.
