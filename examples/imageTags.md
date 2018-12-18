# Demo: change image tags


Define a place to work:

<!-- @makeWorkplace @test -->
```
DEMO_HOME=$(mktemp -d)
```

Make a `kustomization` containing a pod resource

<!-- @createKustomization @test -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
apiVersion: v1
kind: Kustomization
resources:
- pod.yaml
EOF
```

Declare the pod resource

<!-- @createDeployment @test -->
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
The tag `1.29.0` can be changed by adding `imageTags` in `kustomization.yaml`.


Add `imageTags`:
<!-- @addImageTags @test -->
```
cd $DEMO_HOME
kustomize edit set imagetag busybox:1.29.1
```

The `kustomization.yaml` will be added following `imageTags`.
> ```
> imageTags:
> - name: busybox
>   newTag: 1.29.1
> ```

Now build this `kustomization`
<!-- @kustomizeBuild @test -->
```
kustomize build $DEMO_HOME
```

Confirm that this replaces _both_ busybox tags:

<!-- @confirmTags @test -->
```
test 2 == \
  $(kustomize build $DEMO_HOME | grep busybox:1.29.1 | wc -l); \
  echo $?
```
