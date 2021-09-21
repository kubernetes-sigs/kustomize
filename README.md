# kustomize

`kustomize` lets you customize raw, template-free YAML
files for multiple purposes, leaving the original YAML
untouched and usable as is.

`kustomize` targets kubernetes; it understands and can
patch [kubernetes style] API objects.  It's like
[`make`], in that what it does is declared in a file,
and it's like [`sed`], in that it emits edited text.

This tool is sponsored by [sig-cli] ([KEP]).

 - [Installation instructions](https://kubernetes-sigs.github.io/kustomize/installation)
 - [General documentation](https://kubernetes-sigs.github.io/kustomize)
 - [Examples](examples)

[![Build Status](https://prow.k8s.io/badge.svg?jobs=kustomize-presubmit-master)](https://prow.k8s.io/job-history/kubernetes-jenkins/pr-logs/directory/kustomize-presubmit-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-sigs/kustomize)](https://goreportcard.com/report/github.com/kubernetes-sigs/kustomize)

## Hashicorp Vault Integration with Kustomize

This fork of Kustomize allows for integration with [Hashicorp Vault](https://www.vaultproject.io/) by reading secrets from Vault and 
dropping the secrets into a `ConfigMap`. It does so by exposing a `vaultSecretGenerator` as an option in your `kustomization.yml`. Each
entry in the generator corresponds to a secret in an instance of Hashicorp Vault that you provision yourself, which will then be accessible
as a ConfigMap in your `base` or `overlays`.

For instance, let's say you have a base with an application that wants access to an instance of MongoDB, and want to mount the secret in your
kubernetes manifest:

`~/base/manifest.yml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: some-app
  name: some-app-deployment
spec:
  replicas: 2
  selector:
    matchLabels:
      app: some-app
  template:
    metadata:
      labels:
        app: some-app
    spec:
      containers:
        - command: [ "/bin/bash", "-c", "--" ]
          args: [ "while true; do sleep 30; done;" ]
          volumeMounts:
          - name: example-secret
            mountPath: /etc/config/secrets
            readOnly: true
          image: busybox
          imagePullPolicy: Always
          name: some-app
      volumes:
      - name: example-secret
        configMap:
          name: mongo
```

Then you would reference it in your overlay kustomization file as such:

`~/overlay/kustomization.yml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
  - ../base

vaultSecretGenerator:
- name: mongo
  path: secrets/environment/mongo
  secretKey: value
```

***NOTE***: this assumes that you have a secret mounted in your vault cluster at the `path` designated as `secrets/environment/mongo`, 
and with a secret keyed at the `secretKey` value of `value`. It also assumes that the path enabled is running the V1 Secret Engine.

Read more about Vault Secret engines [here](https://www.vaultproject.io/docs/secrets/kv/kv-v1).

### Environment Variables for Vault Access

This version of kustomize relies on a few environment variables to be set.

- `VAULT_ADDR`: Public/Private Address where your Vault Cluster is located
- `VAULT_USERNAME`: Username to access the cluster
- `VAULT_PASSWORD`: Password to access the cluster

***Without these variables set, this plugin will fail.***

### End Result - Vault Access

After applying this in my own kubernetes cluster, I was able to then exec into the kubernetes Pod and check the contents
that were mounted into the pod from Hashicorp Vault.

```bash
(⎈ |stage-eks:default)➜  staging git:(master) ✗ kubectl exec -it some-app-deployment-84bfc64b8d-vvvjq -- sh
$ cat /etc/config/secrets/mongo
this is a vault secret!

```

## kubectl integration

The kustomize build flow at [v2.0.3] was added
to [kubectl v1.14][kubectl announcement].  The kustomize
flow in kubectl remained frozen at v2.0.3 until kubectl v1.21,
which [updated it to v4.0.5][kust-in-kubectl update]. It will
be updated on a regular basis going forward, and such updates
will be reflected in the Kubernetes release notes.

| Kubectl version | Kustomize version |
| --- | --- |
| < v1.14 | n/a |
| v1.14-v1.20 | v2.0.3 |
| v1.21 | v4.0.5 |
| v1.22 | v4.2.0 |

[v2.0.3]: /../../tree/v2.0.3
[#2506]: https://github.com/kubernetes-sigs/kustomize/issues/2506
[#1500]: https://github.com/kubernetes-sigs/kustomize/issues/1500
[kust-in-kubectl update]: https://github.com/kubernetes/kubernetes/blob/4d75a6238a6e330337526e0513e67d02b1940b63/CHANGELOG/CHANGELOG-1.21.md#kustomize-updates-in-kubectl

For examples and guides for using the kubectl integration please
see the [kubectl book] or the [kubernetes documentation].

## Usage


### 1) Make a [kustomization] file

In some directory containing your YAML [resource]
files (deployments, services, configmaps, etc.), create a
[kustomization] file.

This file should declare those resources, and any
customization to apply to them, e.g. _add a common
label_.

![base image][imageBase]

File structure:

> ```
> ~/someApp
> ├── deployment.yaml
> ├── kustomization.yaml
> └── service.yaml
> ```

The resources in this directory could be a fork of
someone else's configuration.  If so, you can easily
rebase from the source material to capture
improvements, because you don't modify the resources
directly.

Generate customized YAML with:

```
kustomize build ~/someApp
```

The YAML can be directly [applied] to a cluster:

> ```
> kustomize build ~/someApp | kubectl apply -f -
> ```


### 2) Create [variants] using [overlays]

Manage traditional [variants] of a configuration - like
_development_, _staging_ and _production_ - using
[overlays] that modify a common [base].

![overlay image][imageOverlay]

File structure:
> ```
> ~/someApp
> ├── base
> │   ├── deployment.yaml
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── development
>     │   ├── cpu_count.yaml
>     │   ├── kustomization.yaml
>     │   └── replica_count.yaml
>     └── production
>         ├── cpu_count.yaml
>         ├── kustomization.yaml
>         └── replica_count.yaml
> ```

Take the work from step (1) above, move it into a
`someApp` subdirectory called `base`, then
place overlays in a sibling directory.

An overlay is just another kustomization, referring to
the base, and referring to patches to apply to that
base.

This arrangement makes it easy to manage your
configuration with `git`.  The base could have files
from an upstream repository managed by someone else.
The overlays could be in a repository you own.
Arranging the repo clones as siblings on disk avoids
the need for git submodules (though that works fine, if
you are a submodule fan).

Generate YAML with

```sh
kustomize build ~/someApp/overlays/production
```

The YAML can be directly [applied] to a cluster:

> ```sh
> kustomize build ~/someApp/overlays/production | kubectl apply -f -
> ```

## Community

- [file a bug](https://kubernetes-sigs.github.io/kustomize/contributing/bugs/) instructions
- [contribute a feature](https://kubernetes-sigs.github.io/kustomize/contributing/features/) instructions

### Code of conduct

Participation in the Kubernetes community
is governed by the [Kubernetes Code of Conduct].

[`make`]: https://www.gnu.org/software/make
[`sed`]: https://www.gnu.org/software/sed
[DAM]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#declarative-application-management
[KEP]: https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/2377-Kustomize/README.md
[Kubernetes Code of Conduct]: code-of-conduct.md
[applied]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#apply
[base]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#base
[declarative configuration]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#declarative-application-management
[imageBase]: docs/images/base.jpg
[imageOverlay]: docs/images/overlay.jpg
[kubectl announcement]: https://kubernetes.io/blog/2019/03/25/kubernetes-1-14-release-announcement
[kubectl book]: https://kubectl.docs.kubernetes.io/guides/introduction/kustomize/
[kubernetes documentation]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/
[kubernetes style]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#kubernetes-style-object
[kustomization]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#kustomization
[overlay]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#overlay
[overlays]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#overlay
[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[resource]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#resource
[resources]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#resource
[sig-cli]: https://github.com/kubernetes/community/blob/master/sig-cli/README.md
[variant]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#variant
[variants]: https://kubernetes-sigs.github.io/kustomize/api-reference/glossary#variant
[v2.0.3]: https://github.com/kubernetes-sigs/kustomize/releases/tag/v2.0.3
[v2.1.0]: https://github.com/kubernetes-sigs/kustomize/releases/tag/v2.1.0
[workflows]: https://kubernetes-sigs.github.io/kustomize/guides
