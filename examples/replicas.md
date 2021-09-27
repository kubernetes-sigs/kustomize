# Demo: change replicas


Define a place to work:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Make a `kustomization` containing a deployment resource with replicas

<!-- @createKustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- deployment.yaml
replicas:
- name: myapp
  count: 3
EOF
```

Declare the deployment resource

<!-- @createDeployment @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myapp
  name: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: myapp-container
        image: busybox:1.29.0
EOF
```

The `myapp` resource declares an deployment managing a pod running a busybox container.


Now build this `kustomization`
<!-- @kustomizeBuild @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME
```

Confirm that this replaces replicas for myapp:

<!-- @confirmImages @testAgainstLatestRelease -->
```
test 3 = \
  $(kustomize build $DEMO_HOME | grep replicas | awk '{print $2}'); \
  echo $?
```
