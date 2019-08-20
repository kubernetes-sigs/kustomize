[overlay]: ../docs/glossary.md#overlay
[target]: ../docs/glossary.md#target

# Demo: combining config data from devops and developers

Scenario: you have a Java-based server storefront in
production that various internal development teams
(signups, checkout, search, etc.) contribute to.

The server runs in different environments:
_development_, _testing_, _staging_ and _production_,
accepting configuration parameters from java property
files.

Using one big properties file for each environment is
difficult to manage.  The files change frequently, and
have to be changed by devops exclusively because

  1. the files must at least partially agree on certain
     values that devops cares about and that developers
     ignore and
  1. because the production
     properties contain sensitive data like production
     database credentials.

## Property sharding

With some study, we notice that the properties are
separable into categories.

### Common properties

E.g. internationalization data, static data like
physical constants, location of external services, etc.

_Things that are the same regardless of environment._

Only one set of values is needed.

Place them in a file called

 * `common.properties`

(relative location defined below).

### Plumbing properties

E.g. serving location of static content (HTML, CSS,
javascript), location of product and customer database
tables, ports expected by load balancers, log sinks,
etc.

_The different values for these properties are
precisely what sets the environments apart._

Devops or SRE will want full control over the values
used in production.  Testing will have fixed
databases supporting testing.  Developers will want
to do whatever they want to try scenarios under
development.

Places these values in

 * `development/plumbing.properties`
 * `staging/plumbing.properties`
 * `production/plumbing.properties`


### Secret properties

E.g. location of actual user tables, database
credentials, decryption keys, etc.

_Things that are a subset of devops controls, that
nobody else has (or should want) access to._

Places these values in

 * `development/secret.properties`
 * `staging/secret.properties`
 * `production/secret.properties`

[kubernetes secret]: https://kubernetes.io/docs/tasks/inject-data-application/distribute-credentials-secure/

and control access to them with (for example) unix file
owner and mode bits, or better yet, put them in
a server dedicated to storing password protected
secrets, and use a field called  `secretGenerator`
in your _kustomization_ to create a kubernetes
secret holding them (not covering that here).

<!--
secretGenerator:
- name: app-tls
  files:
    tls.crt=tls.cert
    tls.key=tls.key
  type: "kubernetes.io/tls"
EOF
-->

## A mixin approach to management

The way to create _n_ cluster environments that share
some common information is to create _n_ overlays of a
common base.

For the rest of this example, we'll do _n==2_, just
_development_ and _production_, since adding more
environments follows the same pattern.

A cluster environment is created by
running `kustomize build` on a [target] that happens to
be an [overlay].

[helloworld]: helloWorld/README.md

The following example will do that, but will focus on
configMap construction, and not worry about how to
connect the configMaps to deployments (that is covered
in the [helloworld] example).


All files - including the shared property files
discussed above - will be created in a directory tree
that is consistent with the base vs overlay file layout
defined in the [helloworld] demo.

It will all live in this work directory:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### Create the base

<!-- kubectl create configmap BOB --dry-run -o yaml --from-file db. -->

Make a place to put the base configuration:

<!-- @baseDir @testAgainstLatestRelease -->
```
mkdir -p $DEMO_HOME/base
```

Make the data for the base.  This direction by
definition should hold resources common to all
environments. Here we're only defining a java
properties file, and a `kustomization` file that
references it.

<!-- @baseKustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/base/common.properties
color=blue
height=10m
EOF

cat <<EOF >$DEMO_HOME/base/kustomization.yaml
configMapGenerator:
- name: my-configmap
  files:
  - common.properties
EOF
```


### Create and use the overlay for _development_

Make an abbreviation for the parent of the overlay
directories:

<!-- @overlays @testAgainstLatestRelease -->
```
OVERLAYS=$DEMO_HOME/overlays
```

Create the files that define the _development_ overlay:

<!-- @developmentFiles @testAgainstLatestRelease -->
```
mkdir -p $OVERLAYS/development

cat <<EOF >$OVERLAYS/development/plumbing.properties
port=30000
EOF

cat <<EOF >$OVERLAYS/development/secret.properties
dbpassword=mothersMaidenName
EOF

cat <<EOF >$OVERLAYS/development/kustomization.yaml
resources:
- ../../base
namePrefix: dev-
nameSuffix: -v1
configMapGenerator:
- name: my-configmap
  behavior: merge
  files:
  - plumbing.properties
  - secret.properties
EOF
```

One can now generate the configMaps for development:

<!-- @runDev @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/development
```

#### Check the ConfigMap name

The name of the generated `ConfigMap` is visible in this
output.

The name should be something like `dev-my-configmap-v1-2gccmccgd5`:

 * `"dev-"` comes from the `namePrefix` field,
 * `"my-configmap"` comes from the `configMapGenerator/name` field,
 * `"-v1"` comes from the `nameSuffix` field,
 * `"-2gccmccgd5"` comes from a deterministic hash that `kustomize`
    computes from the contents of the configMap.

The hash suffix is critical.  If the configMap content
changes, so does the configMap name, along with all
references to that name that appear in the YAML output
from `kustomize`.

The name change means deployments will do a rolling
restart to get new data if this YAML is applied to the
cluster using a command like

> ```
> kustomize build $OVERLAYS/development | kubectl apply -f -
> ```

A deployment has no means to automatically know when or
if a configMap in use by the deployment changes.

If one changes a configMap without changing its name
and all references to that name, one must imperatively
restart the cluster to pick up the change.

The best practice is to treat configMaps as immutable.

Instead of editing configMaps, modify your declarative
specification of the cluster's desired state to
point deployments to _new_ configMaps with _new_ names.
`kustomize` makes this easy with its
`configMapGenerator` directive and associated naming
controls.  A GC process in the k8s master eventually
deletes unused configMaps.


### Create and use the overlay for _production_

Next, create the files for the _production_ overlay:


<!-- @productionFiles @testAgainstLatestRelease -->
```
mkdir -p $OVERLAYS/production

cat <<EOF >$OVERLAYS/production/plumbing.properties
port=8080
EOF

cat <<EOF >$OVERLAYS/production/secret.properties
dbpassword=thisShouldProbablyBeInASecretInstead
EOF

cat <<EOF >$OVERLAYS/production/kustomization.yaml
resources:
- ../../base
namePrefix: prod-
configMapGenerator:
- name: my-configmap
  behavior: merge
  files:
  - plumbing.properties
  - secret.properties
EOF
```

One can now generate the configMaps for production:

<!-- @runProd @testAgainstLatestRelease -->
```
kustomize build $OVERLAYS/production
```

A CICD process could apply this directly to
the cluster using:

> ```
> kustomize build $OVERLAYS/production | kubectl apply -f -
> ```
