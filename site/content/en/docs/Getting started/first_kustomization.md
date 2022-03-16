---
title: "Creating Your First Kustomization"
linkTitle: "Creating Your First Kustomization"
date: 2022-02-27
weight: 20
description: >
  A step by step tutorial for absolute kustomize beginners
---

This page will help you get started with this amazing tool called kustomize! We will start off with a simple nginx deployment manifest and then use it to explore kustomize basics.

### Create initial manifests and directory structure

Let's start off by creating our nginx deployment and service manifests in a dedicated folder structure.

```bash
mkdir -p kustomize-example/base kustomize-example/overlays/staging kustomize-example/overlays/production
cd kustomize-example

cat <<'EOF' >base/deployment.yaml
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

cat <<'EOF' >base/service.yaml
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

Now that we have our `deployment.yaml` and `service.yaml` files created, let's create our Kustomization. [Insert brief explanation of what a Kustomization is], and is defined in a file named `kustomization.yaml`:

```bash
cat <<'EOF' >base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- deployment.yaml
- service.yaml
EOF
```

In this `kustomization.yaml` file, all that we are doing is telling kustomize to include the `deployment.yaml` and `service.yaml` as resources for it to use. So if we now run `kustomize build base/` from our current working directory, all kustomize will do at this point is to generate a manifest which contains the contents of our `deployment.yaml` and `service.yaml` files with no additional changes.

```yaml
$ kustomize build base/

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

### What are Bases and Overlays?

Let's think of a scenario where we need to deploy the nginx manifests from the previous section to two environments called `Staging` and `Production`. The manifests for these two environments will be mostly identical, with only a few minor changes between them. Traditionally to achieve these changes, we could duplicate the manifests and apply the changes manually or rely on some templating engine. With kustomize we can avoid templating and duplication of our manifests and apply the different changes we need using overlays. With this approach, the `base` would contain the common part of the our manifests and the `overlays` contain our environment specific changes.

To continue with this tutorial, create the `kustomization.yaml` files for our two overlays:

```bash
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

The `kustomization.yaml` files in the overlays are just including the `base` folder, so if you were to run `kustomize build` on the overlay folders at this point you would get the same output we got when we built `base`. The directory structure you created so far should look like this:

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

|Requirement|         Production         |         Staging         |
|-----------|----------------------------|-------------------------|
|Name       |overlay1-nginx-production   |overlay2-nginx-staging   |
|Namespace  |production                  |staging                  |
|Replicas   |3                           |2                        |


We can achieve the names required by making use of `namePrefix` and `nameSuffix` as follows:


_kustomize-example/overlays/production/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: overlay1-
nameSuffix: -production

resources:
- ../../base
```

_kustomize-example/overlays/staging/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: overlay2-
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
  name: overlay1-nginx-production    ### service name changed here
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
  name: overlay1-nginx-production    ### deployment name changed here
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

namePrefix: overlay1-
nameSuffix: -production

namespace: production

replicas:
- name: nginx
  count: 3

resources:
- ../../base
```

_kustomize-example/overlays/staging/kustomization.yaml_:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: overlay2-
nameSuffix: -staging

namespace: staging

replicas:
- name: nginx
  count: 2

resources:
- ../../base
```

Looking at the output of `kustomize build` we can see that we have met all the requirements we set out to meet:

_Production overlay build_:

```yaml
$ kustomize build overlays/production/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: overlay1-nginx-production
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
  name: overlay1-nginx-production
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
  name: overlay2-nginx-staging
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
  name: overlay2-nginx-staging
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
