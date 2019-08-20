# Demo: change image names and tags


Define a place to work:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Make a `kustomization` containing a pod resource

<!-- @createKustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- pod.yaml
EOF
```

Declare the pod resource

<!-- @createDeployment @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox:1.29.0
    command: ['sh', '-c', 'echo The app is running! && sleep 3600']
  initContainers:
  - name: init-mydb
    image: busybox:1.29.0
    command: ['sh', '-c', 'until nslookup mydb; do echo waiting for mydb; sleep 2; done;']
EOF
```

The `myapp-pod` resource declares an initContainer and a container, both use the image `busybox:1.29.0`.
The image `busybox` and tag `1.29.0` can be changed by adding `images` in `kustomization.yaml`.


Add `images`:
<!-- @addImages @testAgainstLatestRelease -->
```
cd $DEMO_HOME
kustomize edit set image busybox=alpine:3.6
```

The following `images` will be added to `kustomization.yaml`:
> ```
> images:
> - name: busybox
>   newName: alpine
>   newTag: 3.6
> ```

Now build this `kustomization`
<!-- @kustomizeBuild @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME
```

Confirm that this replaces _both_ busybox images and tags for `alpine:3.6`:

<!-- @confirmImages @testAgainstLatestRelease -->
```
test 2 = \
  $(kustomize build $DEMO_HOME | grep alpine:3.6 | wc -l); \
  echo $?
```
