# Demo: Components

Suppose you've written a very simple Web application:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    spec:
      containers:
      - name: example
        image: example:1.0
```

You want to deploy a **community** edition of this application as SaaS, so you
add support for persistence (e.g. an external database), and bot detection
(e.g. Google reCAPTCHA).

You've now attracted **enterprise** customers who want to deploy it
on-premises, so you add LDAP support, and disable Google reCAPTCHA. At the same
time, the **devs** need to be able to test parts of the application, so they
want to deploy it with some features enabled and others not.

Here's a matrix with the deployments of this application and the features
enabled for each one:


|            | External DB        | LDAP               | reCAPTCHA          |
|------------|:------------------:|:------------------:|:------------------:|
| Community  | :heavy_check_mark: |                    | :heavy_check_mark: |
| Enterprise | :heavy_check_mark: | :heavy_check_mark: |                    |
| Dev        | :white_check_mark: | :white_check_mark: | :white_check_mark: |

So, you want to make it easy to deploy your application in any of the above
three environments. This seems like a work for [variants], so you try to create
three overlays; a `community/`, an `enterprise/` and a `dev/` overlay, that each
has the appropriate features. However, there are two issues:

1. The external DB feature is repeated in the `community/` and `enterprise/`
   overlays. The rest of the features are optionally repeated on the `dev/`
   overlay as well.
2. The `dev/` overlay is dynamic, and uncommenting many lines of YAML to enable
   a single feature is cumbersome.

Ideally, you want to move each feature under a separate overlay, and enable
them per deployment. Enter components.

## Components example

Here's a way to solve this issue, by using a Kustomize feature called
"components".

First, define a place to work:

```shell
DEMO_HOME=$(mktemp -d)
```

Define a common **base** that has a `Deployment` and a simple `ConfigMap`, that
is mounted on the application's container.

```shell
BASE=$DEMO_HOME/base
mkdir $BASE

cat <<EOF >$BASE/kustomization.yaml
resources:
- deployment.yaml

configMapGenerator:
- name: conf
  literals:
    - main.conf=|
        color=cornflower_blue
        log_level=info
EOF

cat <<EOF >$BASE/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    spec:
      containers:
      - name: example
        image: example:1.0
        volumeMounts:
        - name: conf
          mountPath: /etc/config
      volumes:
        - name: conf
          configMap:
            name: conf
EOF
```

Define an `external_db` component, using `kind: KustomizationPatch`, that
creates a `Secret` for the DB password and a new entry in the `ConfigMap`:

```shell
EXT_DB=$DEMO_HOME/components/external_db
mkdir -p $EXT_DB

cat <<EOF >$EXT_DB/kustomization.yaml
kind: KustomizationPatch  # <-- Component notation

secretGenerator:
- name: dbpass
  files:
    - dbpass.txt

patchesStrategicMerge:
  - configmap.yaml

patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: example
  path: deployment.yaml
EOF

cat <<EOF >$EXT_DB/deployment.yaml
- op: add
  path: /spec/template/spec/volumes/0
  value:
    name: dbpass
    secret:
      secretName: dbpass
- op: add
  path: /spec/template/spec/containers/0/volumeMounts/0
  value:
    mountPath: /var/run/secrets/db/
    name: dbpass
EOF

cat <<EOF >$EXT_DB/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: conf
data:
  db.conf: |
    endpoint=127.0.0.1:1234
    name=app
    user=admin
    pass=/var/run/secrets/db/dbpass.txt
EOF
```

Define an `ldap` component, that creates a `Secret` for the LDAP password
and a new entry in the `ConfigMap`:

```shell
LDAP=$DEMO_HOME/components/ldap
mkdir -p $LDAP

cat <<EOF >$LDAP/kustomization.yaml
kind: KustomizationPatch

secretGenerator:
- name: ldappass
  files:
    - ldappass.txt

patchesStrategicMerge:
  - configmap.yaml

patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: example
  path: deployment.yaml
EOF

cat <<EOF >$LDAP/deployment.yaml
- op: add
  path: /spec/template/spec/volumes/0
  value:
    name: ldappass
    secret:
      secretName: ldappass
- op: add
  path: /spec/template/spec/containers/0/volumeMounts/0
  value:
    mountPath: /var/run/secrets/ldap/
    name: ldappass
EOF

cat <<EOF >$LDAP/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: conf
data:
  ldap.conf: |
    endpoint=ldap://ldap.example.com
    bindDN=cn=admin,dc=example,dc=com
    pass=/var/run/secrets/ldap/ldappass.txt
EOF
```

Define a `recaptcha` component, that creates a `Secret` for the reCAPTCHA
site/secret keys and a new entry in the `ConfigMap`:

```shell
RECAPTCHA=$DEMO_HOME/components/recaptcha
mkdir -p $RECAPTCHA

cat <<EOF >$RECAPTCHA/kustomization.yaml
kind: KustomizationPatch

secretGenerator:
- name: recaptcha
  files:
    - site_key.txt
    - secret_key.txt

patchesStrategicMerge:
  - configmap.yaml

patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: example
  path: deployment.yaml
EOF

cat <<EOF >$RECAPTCHA/deployment.yaml
- op: add
  path: /spec/template/spec/volumes/0
  value:
    name: recaptcha
    secret:
      secretName: recaptcha
- op: add
  path: /spec/template/spec/containers/0/volumeMounts/0
  value:
    mountPath: /var/run/secrets/recaptcha/
    name: recaptcha
EOF

cat <<EOF >$RECAPTCHA/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: conf
data:
  recaptcha.conf: |
    enabled=true
    site_key=/var/run/secrets/recaptcha/site_key.txt
    secret_key=/var/run/secrets/recaptcha/secret_key.txt
EOF
```

Define a `community` overlay, that bundles the external DB and reCAPTCHA
components:

```shell
COMMUNITY=$DEMO_HOME/overlays/community
mkdir -p $COMMUNITY

cat <<EOF >$COMMUNITY/kustomization.yaml
kind: Kustomization
resources:
  - ../../components/external_db
  - ../../components/recaptcha
EOF
```

Define an `enterprise` overlay, that bundles the external DB and LDAP
components:

```shell
ENTERPRISE=$DEMO_HOME/overlays/enterprise
mkdir -p $ENTERPRISE

cat <<EOF >$ENTERPRISE/kustomization.yaml
kind: Kustomization
resources:
  - ../../components/external_db
  - ../../components/ldap
EOF
```

Define a `dev` overlay, that point's to all the components and has LDAP
disabled:

```shell
DEV=$DEMO_HOME/overlays/dev
mkdir -p $DEV

cat <<EOF >$DEV/kustomization.yaml
kind: Kustomization
resources:
  - ../../components/external_db
  #- ../../components/ldap
  - ../../components/recaptcha
EOF
```

Now the workspace has following directories:

```shell
├── base
│   ├── deployment.yaml
│   └── kustomization.yaml
├── components
│   ├── external_db
│   │   ├── configmap.yaml
│   │   ├── dbpass.txt
│   │   ├── deployment.yaml
│   │   └── kustomization.yaml
│   ├── ldap
│   │   ├── configmap.yaml
│   │   ├── deployment.yaml
│   │   ├── kustomization.yaml
│   │   └── ldappass.txt
│   └── recaptcha
│       ├── configmap.yaml
│       ├── deployment.yaml
│       ├── kustomization.yaml
│       ├── secret_key.txt
│       └── site_key.txt
└── overlays
    ├── community
    │   └── kustomization.yaml
    ├── dev
    │   └── kustomization.yaml
    └── enterprise
        └── kustomization.yaml
```

With this structure, you can create the YAML files for each deployment as
follows:

```shell
kustomize build overlays/community
kustomize build overlays/enterprise
kustomize build overlays/dev
```

[variants]: multibases/README.md
