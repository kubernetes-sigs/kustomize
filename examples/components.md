# Demo: Components

For more details regarding this feature you can read the
[Kustomize Components KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/1802-kustomize-components.md).

_This example requires Kustomize ``v3.7.0`` or newer_

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
provides the appropriate features for their audience, i.e., public, customers and
developers, respectfully.

## Variants example

Here's the common and most simplistic approach to solve this problem. As we will
soon see, this approach does not scale well in more complex scenarios. However,
it will help you get a better grasp of the problem we are about to tackle and
demonstrate where there is room for improvement.

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
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

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

Define a **community** overlay that:

- generates `Secrets` for external DB's password and reCAPTCHA's keys
- patches the `ConfigMap` of the common base with configurations for external DB
  and reCAPTCHA
- patches the `Deployment` of the common base to mount the generated `Secrets`
  for external DB and reCAPTCHA

```shell
COMMUNITY=$DEMO_HOME/overlays/community
mkdir -p $COMMUNITY

cat <<EOF >$COMMUNITY/kustomization.yaml
kind: Kustomization

resources:
  - ../../base

secretGenerator:
  - name: dbpass
    files:
      - dbpass.txt
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

cat <<EOF >$COMMUNITY/deployment.yaml
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

cat <<EOF >$COMMUNITY/configmap.yaml
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
  recaptcha.conf: |
    enabled=true
    site_key=/var/run/secrets/recaptcha/site_key.txt
    secret_key=/var/run/secrets/recaptcha/secret_key.txt
EOF
```

Create local input files for external DB's password and reCAPTCHA's keys:

```shell

cat <<EOF >$COMMUNITY/dbpass.txt
dbpass
EOF

cat <<EOF >$COMMUNITY/site_key.txt
sitekey
EOF

cat <<EOF >$COMMUNITY/secret_key.txt
secretkey
EOF
```

Define a **enterprise** overlay that:

- generates `Secrets` for LDAP's password and external DB's password
- patches the `ConfigMap` of the common base with configurations for LDAP and
  external DB
- patches the `Deployment` of the common base to mount the generated `Secrets`
  for LDAP and external DB

```shell
ENTERPRISE=$DEMO_HOME/overlays/enterprise
mkdir -p $ENTERPRISE

cat <<EOF >$ENTERPRISE/kustomization.yaml
kind: Kustomization

resources:
  - ../../base

secretGenerator:
  - name: ldappass
    files:
      - ldappass.txt
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

cat <<EOF >$ENTERPRISE/deployment.yaml
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

cat <<EOF >$ENTERPRISE/configmap.yaml
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
  ldap.conf: |
    endpoint=ldap://ldap.example.com
    bindDN=cn=admin,dc=example,dc=com
    pass=/var/run/secrets/ldap/ldappass.txt
EOF
```

Create local input files for LDAP's password and external DB's password:

```shell

cat <<EOF >$ENTERPRISE/ldappass.txt
ldappass
EOF

cat <<EOF >$ENTERPRISE/dbpass.txt
dbpass
EOF
```

Define a **dev** overlay that supports all three features(ExternalDB, LDAP,
reCAPTCHA) and conditionally enables some or all of them. In this example, we
define a dev overlay that supports all the features, but has disabled the LDAP
support, by doing the following::

- generates `Secrets` for external DB's password and reCAPTCHA's keys
- patches the `ConfigMap` of the common base with configurations for external DB
  and reCAPTCHA
- patches the `Deployment` of the common base to mount the generated `Secrets`
  for external DB and reCAPTCHA

```shell
DEV=$DEMO_HOME/overlays/dev
mkdir -p $DEV

cat <<EOF >$DEV/kustomization.yaml
kind: Kustomization

resources:
  - ../../base

secretGenerator:
  # - name: ldappass              <-- Commenting to disable LDAP support
  #   files:
  #     - ldappass.txt
  - name: dbpass
    files:
      - dbpass.txt
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

cat <<EOF >$DEV/deployment.yaml
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
# - op: add                             <-- Commenting to disable LDAP support
#   path: /spec/template/spec/volumes/0
#   value:
#     name: ldappass
#     secret:
#       secretName: ldappass
# - op: add
#   path: /spec/template/spec/containers/0/volumeMounts/0
#   value:
#     mountPath: /var/run/secrets/ldap/
#     name: ldappass
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

cat <<EOF >$DEV/configmap.yaml
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
  # ldap.conf: |                          <-- Commenting to disable LDAP support
  #   endpoint=ldap://ldap.example.com
  #   bindDN=cn=admin,dc=example,dc=com
  #   pass=/var/run/secrets/ldap/ldappass.txt
  recaptcha.conf: |
    enabled=true
    site_key=/var/run/secrets/recaptcha/site_key.txt
    secret_key=/var/run/secrets/recaptcha/secret_key.txt
EOF
```

Create local input files for external DB's password and reCAPTCHA's keys:

```shell

cat <<EOF >$DEV/dbpass.txt
dbpass
EOF

cat <<EOF >$DEV/site_key.txt
sitekey
EOF

cat <<EOF >$DEV/secret_key.txt
secretkey
EOF
```

The above commands result in the following structure:

```shell
├── base
│   ├── deployment.yaml
│   └── kustomization.yaml
└── overlays
    ├── community
    │   ├── configmap.yaml
    │   ├── dbpass.txt
    │   ├── deployment.yaml
    │   ├── kustomization.yaml
    │   ├── secret_key.txt
    │   └── site_key.txt
    ├── dev
    │   ├── configmap.yaml      <-- Refers to multiple features and might contain comments
    │   ├── dbpass.txt
    │   ├── deployment.yaml     <-- Refers to multiple features and might contain comments
    │   ├── kustomization.yaml  <-- Refers to multiple features and might contain comments
    │   ├── secret_key.txt
    │   └── site_key.txt
    └── enterprise
        ├── configmap.yaml
        ├── dbpass.txt
        ├── deployment.yaml
        ├── kustomization.yaml
        └── ldappass.txt
```

The main issues observed with this solution are:

1. Since some features are repeated in the `community/`, `enterprise/` and
   `dev/` overlays, one needs to manually define patches with content that is
   partially identical to patches of different overlays, that also enable this
   feature.
2. The `dev/` overlay is dynamic, i.e., supports multiple optional features. To
   enable/disable any single feature one needs to uncomment/comment many lines
   of YAML which is cumbersome and hard to maintain. Alternatively, one needs
   to maintain a multitude of overlays and track all possible combinations of
   features.
3. Overlays that combine more than one features define patches for resources
   whose content is not dedicated to a single feature. That is, there is no
   semantic isolation per feature, everything gets mixed into a single,
   multi-feature, resource-specific patch.

The variants approach may solve this simple example but it won't scale in the
long run, as the number of features and deployments grow. What if you have `N`
opt-in features and `M` real-world deployment scenarios that ship with `0-N` of
these features?

Ideally, you want to move each feature under a separate, reusable overlay and
enable them on-demand per deployment, i.e., in kustomization files of top-level
overlays. Enter components.

## Components example

Here's an alternative and more [DRY] approach that solves this issue by using a
Kustomize feature called "components". Each opt-in feature gets packaged as a
component, so that it can be referred to from higher-level overlays.

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

## Takeaway

At the end of the day, Kustomize components provide a more flexible way to
enable/disable features and configurations for applications directly from the
kustomization file. This results in more readable, concise and intuitive
overlays.

[variants]: multibases/README.md
[DRY principle]: https://en.wikipedia.org/wiki/Don%27t_repeat_yourself
