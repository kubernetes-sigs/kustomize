---
title: "images"
linkTitle: "images"
type: docs
weight: 9
description: >
    Modify the name, tags and/or digest for images.
---

Images modify the name, tags and/or digest for images without creating patches. 

One can change the `image` in the following ways (Refer the following example to know exactly how this is done):

- `postgres:8` to `my-registry/my-postgres:v1`,
- nginx tag `1.7.9` to `1.8.0`,
- image name `my-demo-app` to `my-app`,
- alpine's tag `3.7` to a digest value

It is possible to set image tags for container images through
the `kustomization.yaml` using the `images` field.  When `images` are
specified, Apply will override the images whose image name matches `name` with a new
tag.


| Field     | Description                                                              | Example Field | Example Result |
|-----------|--------------------------------------------------------------------------|----------| --- |
| `name`    | Match images with this image name| `name: nginx`| |
| `newTag`  | Override the image **tag** or **digest** for images whose image name matches `name`    | `newTag: new` | `nginx:old` -> `nginx:new` |
| `newName` | Override the image **name** for images whose image name matches `name`   | `newName: nginx-special` | `nginx:old` -> `nginx-special:old` |


## Example

### File Input

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  template:
    spec:
      containers:
      - name: mypostgresdb
        image: postgres:8
      - name: nginxapp
        image: nginx:1.7.9
      - name: myapp
        image: my-demo-app:latest
      - name: alpine-app
        image: alpine:3.7

```

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

images:
- name: postgres
  newName: my-registry/my-postgres
  newTag: v1
- name: nginx
  newTag: 1.8.0
- name: my-demo-app
  newName: my-app
- name: alpine
  digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3

resources:
- deployment.yaml

```

### Build Output

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  template:
    spec:
      containers:
      - image: my-registry/my-postgres:v1
        name: mypostgresdb
      - image: nginx:1.8.0
        name: nginxapp
      - image: my-app:latest
        name: myapp
      - image: alpine@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
        name: alpine-app
```

## Setting a Name

The name for an image may be set by specifying `newName` and the name of the old container image.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
images:
  - name: mycontainerregistry/myimage
    newName: differentregistry/myimage
```

## Setting a Tag

The tag for an image may be set by specifying `newTag` and the name of the container image.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
images:
  - name: mycontainerregistry/myimage
    newTag: v1
```

## Setting a Digest

The digest for an image may be set by specifying `digest` and the name of the container image.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
images:
  - name: alpine
    digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
```


## Setting a Tag from the latest commit SHA

A common CI/CD pattern is to tag container images with the git commit SHA of source code.  e.g. if
the image name is `foo` and an image was built for the source code at commit `1bb359ccce344ca5d263cd257958ea035c978fd3`
then the container image would be `foo:1bb359ccce344ca5d263cd257958ea035c978fd3`.

A simple way to push an image that was just built without manually updating the image tags is to
download the [kustomize standalone](https://github.com/kubernetes-sigs/kustomize/) tool and run
`kustomize edit set image` command to update the tags for you.

**Example:** Set the latest git commit SHA as the image tag for `foo` images.

```bash
kustomize edit set image foo:$(git log -n 1 --pretty=format:"%H")
kubectl apply -f .
```

## Setting a Tag from an Environment Variable

It is also possible to set a Tag from an environment variable using the same technique for setting from a commit SHA.

**Example:** Set the tag for the `foo` image to the value in the environment variable `FOO_IMAGE_TAG`.

```bash
kustomize edit set image foo:$FOO_IMAGE_TAG
kubectl apply -f .
```

{{< alert color="success" title="Committing Image Tag Updates" >}}
The `kustomization.yaml` changes *may* be committed back to git so that they
can be audited.  When committing the image tag updates that have already
been pushed by a CI/CD system, be careful not to trigger new builds +
deployments for these changes.
{{< /alert >}}