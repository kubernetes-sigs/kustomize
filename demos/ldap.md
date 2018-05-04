[base]: ../docs/glossary.md#base
[gitops]: ../docs/glossary.md#gitops
[instance]: ../docs/glossary.md#instance
[instances]: ../docs/glossary.md#instance
[kustomization]: ../docs/glossary.md#kustomization
[overlay]: ../docs/glossary.md#overlay
[overlays]: ../docs/glossary.md#overlay

# Demo: LDAP with instances

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
> DEMO_HOME=~/ldap
> ```

## Establish the base

To use [overlays] to create [instances], we must
first establish a common [base].

To keep this document shorter, the base resources are
off in a supplemental data directory rather than
declared here as HERE documents.  Download them:

<!-- @downloadBase @test -->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE

resources="https://raw.githubusercontent.com/kubernetes/kubectl\
/master/cmd/kustomize/demos/data/ldap/base\
/{deployment.yaml,kustomization.yaml,service.yaml,env.startup.txt}"

curl -s $resources -o "$BASE/#1"
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
>     ├── deployment.yaml
>     ├── env.startup.txt
>     ├── kustomization.yaml
>     └── service.yaml
> ```


One could immediately apply these resources to a
cluster:

> ```
> kubectl apply -f $DEMO_HOME/base
> ```

to instantiate the _ldap_ service.  `kubectl`
would only recognize the resource files.

### The Base Kustomization

The `base` directory has a [kustomization] file:

<!-- @showKustomization @test -->
```
more $BASE/kustomization.yaml
```

Optionally, run `kustomize` on the base to emit
customized resources to `stdout`:

<!-- @buildBase @test -->
```
kustomize build $BASE
```

### Customize the base

A first customization step could be to set the name prefix to all resources:

<!-- @namePrefix @test -->
```
cd $BASE
kustomize edit set nameprefix "my-"
```

See the effect:
<!-- @checkNameprefix @test -->
```
kustomize build $BASE | grep -C 3 "my-"
```

## Create Overlays

Create a _staging_ and _production_ [overlay]:

 * _Staging_ adds a configMap.
 * _Production_ has a higher replica count and a persistent disk.
 * [instances] will differ from each other.

<!-- @overlayDirectories @test -->
```
OVERLAYS=$DEMO_HOME/overlays
mkdir -p $OVERLAYS/staging
mkdir -p $OVERLAYS/production
```

#### Staging Kustomization

Download the staging customization and patch.
<!-- @downloadStagingKustomization @test -->
```
resources="https://raw.githubusercontent.com/kubernetes/kubectl\
/master/cmd/kustomize/demos/data/ldap/overlays/staging\
/{config.env,deployment.yaml,kustomization.yaml}"

curl -s $resources -o "$OVERLAYS/staging/#1"
```
The staging customization adds a configMap.
> ```cat $OVERLAYS/staging/kustomization.yaml
> (...truncated)
> configMapGenerator:
>   - name: env-config
>     files:
>       - config.env
> ```
as well as 2 replica
> ```cat $OVERLAYS/staging/deployment.yaml
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   name: ldap
> spec:
>   replicas: 2
> ```

#### Production Kustomization

Download the production customization and patch.
<!-- @downloadProductionKustomization @test -->
```
resources="https://raw.githubusercontent.com/kubernetes/kubectl\
/master/cmd/kustomize/demos/data/ldap/overlays/production\
/{deployment.yaml,kustomization.yaml}"

curl -s $resources -o "$OVERLAYS/production/#1"
```

The production customization adds 6 replica as well as a consistent disk.
> ```cat $OVERLAYS/production/deployment.yaml
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   name: ldap
> spec:
>   replicas: 6
>   template:
>     spec:
>       volumes:
>         - name: ldap-data
>           emptyDir: null
>           gcePersistentDisk:
>             pdName: ldap-persistent-storage
> ```

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
> │   ├── deployment.yaml
> │   ├── env.startup.txt
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── production
>     │   ├── deployment.yaml
>     │   └── kustomization.yaml
>     └── staging
>         ├── config.env
>         ├── deployment.yaml
>         └── kustomization.yaml
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

The difference output should look something like

> ```diff
> (...truncated)
> <   name: staging-my-ldap-configmap-kftftt474h
> ---
> >   name: production-my-ldap-configmap-k27f7hkg4f
> 85c75
> <   name: staging-my-ldap-service
> ---
> >   name: production-my-ldap-service
> 97c87
> <   name: staging-my-ldap
> ---
> >   name: production-my-ldap
> 99c89
> <   replicas: 2
> ---
> >   replicas: 6
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

