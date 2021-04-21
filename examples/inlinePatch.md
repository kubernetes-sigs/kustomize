[Strategic Merge Patch]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md
[JSON Patch]: https://tools.ietf.org/html/rfc6902

# Demo: Inline Patch

A kustomization file supports patching in three ways:
- patchesStrategicMerge: A list of patch files where each file is parsed as a [Strategic Merge Patch].
- patchesJSON6902: A list of patches and associated targets, where each file is parsed as a [JSON Patch] and can only be applied to one target resource.
- patches: A list of patches and their associated targets. The patch can be applied to multiple objects. It auto detects whether the patch is a [Strategic Merge Patch] or [JSON Patch].

Since 3.2.0, all three support inline patch, where the patch content is put inside the kustomization file as a single string. With this feature, no separate patch files need to be created.

Make a base kustomization containing a Deployment resource.
<!-- @createKustomization @test -->
```
DEMO_HOME=$(mktemp -d)

BASE=$DEMO_HOME/base
mkdir $BASE

cat <<EOF >$BASE/kustomization.yaml
resources:
- deployments.yaml
EOF

cat <<EOF >$BASE/deployments.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: nginx
          image: nginx
          args:
          - one
          - two
EOF
```


## Inline Patch for PatchesStrategicMerge

Create an overlay and add an inline patch in `patchesStrategicMerge` field to the kustomization file
to change the image from `nginx` to `nginx:latest`.

<!-- @addSMPatch @test -->
```
SMP_OVERLAY=$DEMO_HOME/smp
mkdir $SMP_OVERLAY
cat <<EOF >$SMP_OVERLAY/kustomization.yaml
resources:
- ../base

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deploy
  spec:
    template:
      spec:
        containers:
        - name: nginx
          image: nginx:latest
          
EOF
```

Running `kustomize build $SMP_OVERLAY`, in the output confirm that image is updated successfully.

<!-- @confirmSMPatch @test -->
```
test 1 == \
  $(kustomize build $SMP_OVERLAY | grep "image: nginx:latest" | wc -l); \
  echo $?
```

The output is
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: nginx
          image: nginx:latest
          args:
          - one
          - two
```

`$patch: delete` and `$patch: replace` also work in the inline patch. Change the inline patch to delete the container `nginx`.

<!-- @addDeleteSMPatch @test -->
```
cat <<EOF >$SMP_OVERLAY/kustomization.yaml
resources:
- ../base

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deploy
  spec:
    template:
      spec:
        containers:
        - name: nginx
          $patch: delete
          
EOF
```
Running `kustomize build $SMP_OVERLAY`, in the output confirm that the `nginx` container has been deleted.

<!-- @confirmDeleteSMPatch @test -->
```
test 0 == \
  $(kustomize build $SMP_OVERLAY | grep "image: nginx" | wc -l); \
  echo $?
```

The output is
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers: []
```

## Inline Patch for PatchesJson6902

Create an overlay and add an inline patch in `patchesJSON6902` field to the kustomization file
to change the image from `nginx` to `nginx:latest`.

<!-- @addJSONPatch @test -->
```
JSON_OVERLAY=$DEMO_HOME/json
mkdir $JSON_OVERLAY
cat <<EOF >$JSON_OVERLAY/kustomization.yaml
resources:
- ../base

patchesJSON6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: deploy
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/image
      value: nginx:latest
EOF
```

Running `kustomize build $JSON_OVERLAY`, in the output confirm that image is updated successfully.

<!-- @confirmJSONPatch @test -->
```
test 1 == \
  $(kustomize build $JSON_OVERLAY | grep "image: nginx:latest" | wc -l); \
  echo $?
```

The output is
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: nginx
          image: nginx:latest
          args:
          - one
          - two
```

## Inline Patch for Patches

Create an overlay and add an inline patch in `patches` field to the kustomization file
to change the image from `nginx` to `nginx:latest`.

<!-- @addPatch @test -->
```
PATCH_OVERLAY=$DEMO_HOME/patch
mkdir $PATCH_OVERLAY
cat <<EOF > $PATCH_OVERLAY/kustomization.yaml
resources:
- ../base

patches:
- target:
    kind: Deployment
    name: deploy
  patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: deploy
    spec:
      template:
        spec:
          containers:
          - name: nginx
            image: nginx:latest
EOF
```

Running `kustomize build $PATCH_OVERLAY`, in the output confirm that image is updated successfully.

<!-- @confirmPatch @test -->
```
test 1 == \
  $(kustomize build $PATCH_OVERLAY | grep "image: nginx:latest" | wc -l); \
  echo $?
```

The output is
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: nginx
          image: nginx:latest
          args:
          - one
          - two
```
