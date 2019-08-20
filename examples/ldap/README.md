[base]: ../../docs/glossary.md#base
[gitops]: ../../docs/glossary.md#gitops
[kustomization]: ../../docs/glossary.md#kustomization
[overlay]: ../../docs/glossary.md#overlay
[overlays]: ../../docs/glossary.md#overlay
[variant]: ../../docs/glossary.md#variant
[variants]: ../../docs/glossary.md#variant

# Demo: LDAP with variants

Steps:

 1. Clone an existing configuration as a [base].
 1. Customize it.
 1. Create two different [overlays] (_staging_ and _production_)
    from the customized base.
 1. Run kustomize and kubectl to deploy staging and production.

First define a place to work:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Alternatively, use

> ```
> DEMO_HOME=~/ldap
> ```

## Establish the base

To use [overlays] to create [variants], we must
first establish a common [base].

To keep this document shorter, the base resources are
off in a supplemental data directory rather than
declared here as HERE documents.  Download them:

<!-- @downloadBase @testAgainstLatestRelease -->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE

CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/ldap"

curl -s -o "$BASE/#1" "$CONTENT/base\
/{deployment.yaml,kustomization.yaml,service.yaml,env.startup.txt}"
```

Look at the directory:

<!-- @runTree @testAgainstLatestRelease -->
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

<!-- @showKustomization @testAgainstLatestRelease -->
```
more $BASE/kustomization.yaml
```

Optionally, run `kustomize` on the base to emit
customized resources to `stdout`:

<!-- @buildBase @testAgainstLatestRelease -->
```
kustomize build $BASE
```

### Customize the base

A first customization step could be to set the name prefix to all resources:

<!-- @namePrefix @testAgainstLatestRelease -->
```
cd $BASE
kustomize edit set nameprefix "my-"
```

See the effect:
<!-- @checkNameprefix @testAgainstLatestRelease -->
```
kustomize build $BASE | grep -C 3 "my-"
```

## Create Overlays

Create a _staging_ and _production_ [overlay]:

 * _Staging_ adds a configMap.
 * _Production_ has a higher replica count and a persistent disk.
 * [variants] will differ from each other.

<!-- @overlayDirectories @testAgainstLatestRelease -->
```
OVERLAYS=$DEMO_HOME/overlays
mkdir -p $OVERLAYS/staging
mkdir -p $OVERLAYS/production
```

#### Staging Kustomization

Download the staging customization and patch.

<!-- @downloadStagingKustomization @testAgainstLatestRelease -->
```
curl -s -o "$OVERLAYS/staging/#1" "$CONTENT/overlays/staging\
/{config.env,deployment.yaml,kustomization.yaml}"
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
<!-- @downloadProductionKustomization @testAgainstLatestRelease -->
```
curl -s -o "$OVERLAYS/production/#1" "$CONTENT/overlays/production\
/{deployment.yaml,kustomization.yaml}"
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
   and _production_ [variants] in a cluster.

Review the directory structure and differences:

<!-- @listFiles @testAgainstLatestRelease -->
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

<!-- @buildStaging @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/staging
```

<!-- @buildProduction @testAgainstLatestRelease -->
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
