---
title: "Binaries"
linkTitle: "Binaries"
weight: 3
type: docs
description: >
    Install Kustomize by downloading precompiled binaries.
---

<meta http-equiv="refresh" content="0; url=https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/" />

Binaries at various versions for linux, MacOs and Windows are published on the [releases page].

The following [script] detects your OS and downloads the appropriate kustomize binary to your
current working directory.  

```
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
```

[releases page]: https://github.com/kubernetes-sigs/kustomize/releases
[script]: https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh
