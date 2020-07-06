---
title: "Kustomize Components"
linkTitle: "Kustomize Components"
type: docs
description: >
    Kustomize components guide
---

As of ``v3.7.0`` Kustomize supports a special type of kustomization that allows
one to define reusable pieces of configuration logic that can be included from
multiple overlays.

Components come in handy when dealing with applications that support multiple
optional features and you wish to enable only a subset of them in different
overlays, i.e., different features for different environments or audiences.

For more details regarding this feature you can read the
[Kustomize Components KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/1802-kustomize-components.md).

## Use case

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
| Community  | ✔️                  |                    | ✔️                  |
| Enterprise | ✔️                  | ✔️                  |                    |
| Dev        | ✅                 | ✅                 | ✅                 |

(✔️ enabled, ✅: optional)

So, you want to make it easy to deploy your application in any of the above
three environments. Here's how you can do this with Kustomize components: each
opt-in feature gets packaged as a component, so that it can be referred to from
multiple higher-level overlays.

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

Define an `external_db` component, using `kind: Component`, that creates a
`Secret` for the DB password and a new entry in the `ConfigMap`:

```shell
EXT_DB=$DEMO_HOME/components/external_db
mkdir -p $EXT_DB

cat <<EOF >$EXT_DB/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1alpha1  # <-- Component notation
kind: Component

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
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component

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
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component

secretGenerator:
- name: recaptcha
  files:
    - site_key.txt
    - secret_key.txt

# Updating the ConfigMap works with generators as well.
configMapGenerator:
- name: conf
  behavior: merge
  literals:
    - recaptcha.conf=|
        enabled=true
        site_key=/var/run/secrets/recaptcha/site_key.txt
        secret_key=/var/run/secrets/recaptcha/secret_key.txt

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
```

Define a `community` variant, that bundles the external DB and reCAPTCHA
components:

```shell
COMMUNITY=$DEMO_HOME/overlays/community
mkdir -p $COMMUNITY

cat <<EOF >$COMMUNITY/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

components:
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
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

components:
  - ../../components/external_db
  - ../../components/ldap
EOF
```

Define a `dev` overlay, that points to all the components and has LDAP
disabled:

```shell
DEV=$DEMO_HOME/overlays/dev
mkdir -p $DEV

cat <<EOF >$DEV/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

components:
  - ../../components/external_db
  #- ../../components/ldap
  - ../../components/recaptcha
EOF
```

Now, the workspace has the following directories:

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

With this structure, you can generate the YAML manifests for each deployment
using `kustomize build`:

```shell
kustomize build overlays/community
kustomize build overlays/enterprise
kustomize build overlays/dev
```
