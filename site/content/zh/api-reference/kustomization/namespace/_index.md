---
title: "namespace"
linkTitle: "namespace"
type: docs
description: >
    为所有资源添加 namespace。
---

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: my-namespace
```

如果在资源上设置了现有 namespace，则将覆盖现有 namespace；如果在资源上未设置现有 namespace，则使用现有 namespace。
