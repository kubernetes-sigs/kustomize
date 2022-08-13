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

The build output for our `Production` overlay would now be:

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

### Further customisations

Now that you have seen how kustomize works, let's add a few more requirements:

|Requirement|             Production             |           Staging           |
|-----------|------------------------------------|-----------------------------|
|Image      |nginx:1.20.2                        |nginx:latest                 |
|Label      |variant=var1                        |variant=var2                 |
|Env Var    |ENVIRONMENT=prod                    |ENVIRONMENT=stg              |

To keep the example brief, we will just be showing the changes for the `Production` overlay and then present the updated overlay files and builds for both overlays at the end.  The specific image tag can be set by making use of the `images` field. Add the following to the kustomization files in your overlays:

```yaml
images:
- name: nginx
  newTag: 1.20.2  ## For the Staging overlay set this to 'latest'
```

For setting the label, we can use the `labels` field. Add the following to the kustomization files in your overlays:

```yaml
labels:
- pairs:
    variant: var1  ## For the Staging overlay set this to 'var2'
  includeSelectors: false # Setting this to false so that the label is not added to the selectors
  includeTemplates: true  # Setting this to true will make the label available also on the pod and not just the deployment
```

At this point, your kustomization files for your `Production` overlay should be as follows:

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

images:
- name: nginx
  newTag: 1.20.2

labels:
- pairs:
    variant: var1
  includeSelectors: false
  includeTemplates: true

resources:
- ../../base
```

Rebuilding the `Production` overlay gives the following:

```yaml
$ kustomize build overlays/production/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
    variant: var1                ### label has been set here
  name: env1-example-nginx-production
  namespace: production
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
    variant: var1                ### label has been set here
  name: env1-example-nginx-production
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
        variant: var1            ### label has been set here
    spec:
      containers:
      - image: nginx:1.20.2      ### image tag has been set to 1.20.2
        name: nginx
        ports:
        - containerPort: 80
```

Our last requirement to meet is to set the environment variable, and to do that we will create a patch.  To do this, create the following file for the `Production` overlay:

```bash
cat <<'EOF' >overlays/production/patch-env-vars.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ENVIRONMENT
          value: prod
EOF
```

Next step, add a reference to that patch file in `kustomization.yaml`:

```yaml
patches:
- patch-env-vars.yaml
```

One important thing to note here is that the name of the deployment used is the name that we are getting from our base and not the deployment name that has the prefix and suffix added.

Rebuilding the overlay shows that the environment variable has been added to our container:

```yaml
$ kustomize build overlays/production/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
    variant: var1
  name: env1-example-nginx-production
  namespace: production
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
    variant: var1
  name: env1-example-nginx-production
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
        variant: var1
    spec:
      containers:
      - env:
        - name: ENVIRONMENT               ### Environment variable has been added here
          value: prod
        image: nginx:1.20.2
        name: nginx
        ports:
        - containerPort: 80
```

Looking at the output of `kustomize build` we can see that the additional requirements we set out to meet have been met.  Below are the files as they should be at this point in your overlays and the `kustomize build` output:

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

images:
- name: nginx
  newTag: 1.20.2

labels:
- pairs:
    variant: var1
  includeSelectors: false
  includeTemplates: true

resources:
- ../../base

patches:
- patch-env-vars.yaml
```

_kustomize-example/overlays/production/patch-env-vars.yaml_:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ENVIRONMENT
          value: prod
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

images:
- name: nginx
  newTag: latest

labels:
- pairs:
    variant: var2
  includeSelectors: false
  includeTemplates: true

resources:
- ../../base

patches:
- patch-env-vars.yaml
```

_kustomize-example/overlays/staging/patch-env-vars.yaml_:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ENVIRONMENT
          value: stg
```

_Production overlay build_:

```yaml
$ kustomize build overlays/production/

apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
    variant: var1
  name: env1-example-nginx-production
  namespace: production
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
    variant: var1
  name: env1-example-nginx-production
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
        variant: var1
    spec:
      containers:
      - env:
        - name: ENVIRONMENT
          value: prod
        image: nginx:1.20.2
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
    variant: var2
  name: env2-example-nginx-staging
  namespace: staging
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
    variant: var2
  name: env2-example-nginx-staging
  namespace: staging
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
        variant: var2
    spec:
      containers:
      - env:
        - name: ENVIRONMENT
          value: stg
        image: nginx:latest
        name: nginx
        ports:
        - containerPort: 80
```

### Next steps

Congratulations on making it to the end of this tutorial.  As a summary for you, these are the customisations that we did in this tutorial:

- Add a name prefix and a name suffix
- Set the namespace for our resources
- Set the number of replicas for our deployment
- Set the image to use
- Add a label to our resources
- Add an environment variable to a container by using a patch

These are just a few of the things kustomize can do, and if you are interested to learn more the kustomization reference (to add link here) is your next step. You will see how you can use components to define base resources and add them to specific overlays where needed, use generators to create configMaps from files and much more!
