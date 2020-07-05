# Patching multiple resources at once.

kustomize supports patching via either a
[strategic merge patch] (wherein you
partially re-specify the thing you want to
modify, with in-place changes) or a
[JSON patch] (wherein you specify specific
operation/target/value tuples in a particular
syntax).

A kustomize file lets one specify many
patches.  Each patch must be associated with
a _target selector_:

[strategic merge patch]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md
[JSON patch]: jsonpatch.md

> ```yaml
> patches:
> - path: <relative path to file containing patch>
>   target:
>     group: <optional group>
>     version: <optional version>
>     kind: <optional kind>
>     name: <optional name>
>     namespace: <optional namespace>
>     labelSelector: <optional label selector>
>     annotationSelector: <optional annotation selector>
> ```

E.g. select resources with _name_ matching `foo*`:

> ```yaml
> target:
>   name: foo*
> ```

Select all resources of _kind_ `Deployment`:

> ```yaml
> target:
>   kind: Deployment
> ```

[label/annotation selector rules]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors

Using multiple fields just makes the target
more specific. The following selects only
Deployments that also have the _label_ `app=hello`
(full [label/annotation selector rules]):

> ```yaml
> target:
>   kind: Deployment
>   labelSelector: app=hello
> ```

### Demo

The example below shows how to inject a
sidecar container for multiple Deployment
resources.


Make a place to work:

<!-- @demoHome @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Make a file describing two Deployments:

<!-- @createDeployments @testAgainstLatestRelease -->
```
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

Declare a [strategic merge patch] file
to inject a sidecar container:

<!-- @definePatch @testAgainstLatestRelease -->
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

Finally, define a kustomization file
that specifies both a `patches` and `resources`
entry:

<!-- @createKustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- deployments.yaml

patches:
- path: patch.yaml
  target:
    kind: Deployment
EOF
```

The expected result is:

<!-- @definedExpectedOutput @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/out_expected.yaml
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
        - one
        - two
        image: nginx
        name: nginx
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
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
      - image: busybox
        name: busybox
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
EOF
```

Run the build:
<!-- @runIt @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME >$DEMO_HOME/out_actual.yaml
```

Confirm expectations:

<!-- @diffShouldExitZero @testAgainstLatestRelease -->
```
diff $DEMO_HOME/out_actual.yaml $DEMO_HOME/out_expected.yaml
```

To see how to do this with JSON patches,
try the [JSON patch] demo.
