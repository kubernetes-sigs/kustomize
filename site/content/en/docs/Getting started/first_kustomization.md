---
title: "Creating Your First Kustomization"
linkTitle: "Creating Your First Kustomization"
date: 2022-02-27
weight: 20
description: >
  A step by step tutorial for absolute kustomize beginners
---

This page will help you get started with this amazing tool called kustomize! We will start off with a simple nginx deployment manifest and then use it to explore kustomize basics.

### Create your resource manifests and Kustomization

Let's start off by creating our nginx deployment and service manifests in a dedicated folder:

```bash
mkdir kustomize-example
cd kustomize-example

cat <<'EOF' >deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
EOF

cat <<'EOF' >service.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
EOF
```

Now that we have our `deployment.yaml` and `service.yaml` files created, let's create our Kustomization. We can think of Kustomization as the set of instructions that tell kustomize what it needs to do, and it is defined in a file named `kustomization.yaml`:

```bash
cat <<'EOF' >kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- deployment.yaml
- service.yaml
EOF
```

In this kustomization file, we are telling kustomize to include the `deployment.yaml` and `service.yaml` as its resources. If we now run `kustomize build .` from our current working directory, kustomize will generate a manifest containing the contents of our `deployment.yaml` and `service.yaml` files with no additional changes.

```yaml
$ kustomize build .

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
```

### Customise your resources

So far we have not used kustomize to do any modifications, so let's see how we can do that. Kustomize comes with a considerable number of transformers that apply changes to our manifests, and in this section we will have a look at the `namePrefix` transformer.  This transformer will add a prefix to the deployment and service names.  Modify the `kustomization.yaml` file as follows:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: example-  ### add this line

resources:
- deployment.yaml
- service.yaml
```

After re-building we see can see our modified manifest which now has the prefixed deployment and service names:

```yaml
$ kustomize build .

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: example-nginx    ### service name changed here
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: example-nginx    ### deployment name changed here
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
```

### Create variants using overlays

Now let's assume we need to deploy the nginx manifests from the previous section to two environments called `Staging` and `Production`. The manifests for these two environments will be mostly identical, with only a few minor changes between them. We call these two mostly identical manifests "variants". Traditionally to create variants, we could duplicate the manifests and apply the changes manually or rely on some templating engine. With kustomize, we can avoid templating and duplication of our manifests and apply the different changes we need using overlays. With this approach, the `base` would contain the common part of the our variants and the `overlays` contain our environment specific changes.

Create the `kustomization.yaml` files for our two overlays and move the files we have so far into `base`:

```bash
mkdir -p base overlays/staging overlays/production
mv deployment.yaml kustomization.yaml service.yaml base


cat <<'EOF' >overlays/staging/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../../base
EOF

cat <<'EOF' >overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../../base
EOF
```

The kustomization files for the overlays include just the `base` folder, so if you were to run `kustomize build` on the overlay folders at this point you would get the same output you would get if you built `base`.  It is important to note that bases can be included in the `resources` field in the same way that we included our other deployment and service resource files.

The directory structure you created so far should look like this:

```
kustomize-example
├── base
│   ├── deployment.yaml
│   ├── kustomization.yaml
│   └── service.yaml
└── overlays
    ├── production
    │   └── kustomization.yaml
    └── staging
        └── kustomization.yaml
```

### Customising our overlays

For the purposes our example, let's define some requirements of how our deployment should look like in the two environments:

|Requirement|             Production             |           Staging           |
|-----------|------------------------------------|-----------------------------|
|Name       |env1-example-nginx-production       |env2-example-nginx-staging   |
|Namespace  |production                          |staging                      |
|Replicas   |3                                   |2                            |


We can achieve the names required by making use of `namePrefix` and `nameSuffix` as follows:


_kustomize-example/overlays/production/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: env1-
nameSuffix: -production

resources:
- ../../base
```

_kustomize-example/overlays/staging/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: env2-
nameSuffix: -staging

resources:
- ../../base
```

The build output for our Production overlay would now be:

```yaml
$ kustomize build overlays/production/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: env1-example-nginx-production    ### service name changed here
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: env1-example-nginx-production    ### deployment name changed here
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
```

It is important to note here that the name for _both_ the `deployment` and the `service` were updated with the `namePrefix` and `nameSuffix` defined. If we had additional kubernetes objects (like an `ingress`) their name would be updated as well.


Moving on to our next requirements, we can set the namespace and the number of replicas we want by using `namespace` and `replicas` respectively:


_kustomize-example/overlays/production/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: env1-
nameSuffix: -production

namespace: production

replicas:
- name: example-nginx
  count: 3

resources:
- ../../base
```

_kustomize-example/overlays/staging/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: env2-
nameSuffix: -staging

namespace: staging

replicas:
- name: example-nginx
  count: 2

resources:
- ../../base
```

Note that the deployment name being referenced in `replicas` is the modified name that was output by `base`. Looking at the output of `kustomize build` we can see that we have met all the requirements we set out to meet:

_Production overlay build_:

```yaml
$ kustomize build overlays/production/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: env1-example-nginx-production
  namespace: production              ### namespace has been set to production
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: env1-example-nginx-production
  namespace: production              ### namespace has been set to production
spec:
  replicas: 3                        ### replicas have been updated from 1 to 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
```

_Staging overlay build_:

```yaml
$ kustomize build overlays/staging/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: env2-example-nginx-staging
  namespace: staging                 ### namespace has been set to staging
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: env2-example-nginx-staging
  namespace: staging                 ### namespace has been set to staging
spec:
  replicas: 2                        ### replicas have been from 1 to 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
```
