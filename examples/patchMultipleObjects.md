[Strategic Merge Patch]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md
[JSON patches]: https://tools.ietf.org/html/rfc6902
[label selector]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors


# Demo: applying a patch to multiple resources

A kustomization file supports customizing resources via both
[Strategic Merge Patch] and [JSON patches]. Now one patch can be
applied to multiple resources.

This can be done by specifying a patch and a target selector as follows:
```
patches:
- path: <PatchFile>
  target:
    group: <Group>
    version: <Version>
    kind: <Kind>
    name: <Name>
    namespace: <Namespace>
    labelSelector: <LabelSelector>
    annotationSelector: <AnnotationSelector>
```
Both `labelSelector` and `annotationSelector` should follow the convention in [label selector].
Kustomize selects the targets which match all the fields in `target` to apply the patch.

The example below shows how to inject a sidecar container for all deployment resources.

Make a `kustomization` containing a Deployment resource.

<!-- @createDeployment @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- deployments.yaml
EOF

cat <<EOF >$DEMO_HOME/deployments.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
        - name: nginx
          image: nginx
          args:
          - one
          - two
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    metadata:
      labels:
        key: value
    spec:
      containers:
        - name: busybox
          image: busybox
EOF
```

Declare a Strategic Merge Patch file to inject a sidecar container:

<!-- @addPatch @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-important
spec:
  template:
    spec:
      containers:
        - name: istio-proxy
          image: docker.io/istio/proxyv2
          args:
          - proxy
          - sidecar
EOF
```

Apply the patch by adding _patches_ field in kustomization.yaml

<!-- @applyPatch @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
EOF
```

Running `kustomize build $DEMO_HOME`, in the output confirm that both Deployment resources are patched correctly.

<!-- @confirmPatch @testAgainstLatestRelease -->
```
test 2 == \
  $(kustomize build $DEMO_HOME | grep "image: docker.io/istio/proxyv2" | wc -l); \
  echo $?
```

The output is as follows:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
      - args:
        - one
        - two
        image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    metadata:
      labels:
        key: value
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
      - image: busybox
        name: busybox
```

## Target selector
- Select resources with name matching `name*`
  ```yaml
  target:
    name: name*
  ```
- Select all Deployment resources
  ```yaml
  target:
    kind: Deployment
  ```
- Select resources matching label `app=hello`
  ```yaml
  target:
    labelSelector: app=hello
  ```
- Select resources matching annotation `app=hello`
  ```yaml
  target:
    annotationSelector: app=hello
  ```
- Select all Deployment resources matching label `app=hello`
  ```yaml
  target:
    kind: Deployment
    labelSelector: app=hello
  ```