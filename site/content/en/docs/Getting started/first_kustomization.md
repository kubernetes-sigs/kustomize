---
title: "Creating Your First Kustomization"
linkTitle: "Creating Your First Kustomization"
date: 2022-02-27
weight: 20
description: >
  A simple project example to get you familiar with the concepts
---

We're going to use kustomize to deploy an nginx instance into our Kubernetes cluster.

## Creating the directory structure

Let's firt create a directory to store our kustomize project.
```bash
mkdir kustomize-nginx && cd kustomize-nginx
```
Create a `base` folder:
```bash
mkdir base
```
Inside this folder we will create two files:
* `kustomization.yaml` - the configuration file for kustomize
* `deployment.yaml` - the definition for our nginx deployment

`kustomization.yaml`
```bash
cat <<'EOF' >base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: kustomize-nginx

resources:
- deployment.yaml
EOF
```
The file defines the `apiVersion`, the `kind` and the `resources` it manages.

`deployment.yaml`
```bash
cat <<'EOF' >base/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
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
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
EOF
```
TBC...