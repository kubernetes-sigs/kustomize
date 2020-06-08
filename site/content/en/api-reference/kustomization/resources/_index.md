---
title: "resources"
linkTitle: "resources"
type: docs
description: >
    Resources to include.
---

Each entry in this list must be a path to a _file_, or a path (or URL) referring to another
kustomization _directory_, e.g.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- myNamespace.yaml
- sub-dir/some-deployment.yaml
- ../../commonbase
- github.com/kubernetes-sigs/kustomize/examples/multibases?ref=v1.0.6
- deployment.yaml
- github.com/kubernets-sigs/kustomize/examples/helloWorld?ref=test-branch
```

Resources will be read and processed in depth-first order.

Files should contain k8s resources in YAML form. A file may contain multiple resources separated by
the document marker `---`.  File paths should be specified _relative_ to the directory holding the
kustomization file containing the `resources` field.

[hashicorp URL]: https://github.com/hashicorp/go-getter#url-format

Directory specification can be relative, absolute, or part of a URL.  URL specifications should
follow the [hashicorp URL] format.  The directory must contain a `kustomization.yaml` file.
