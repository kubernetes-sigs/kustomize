[Issue2479]: https://github.com/kubernetes-sigs/kustomize/issues/2479

# Investigate [Issue2479]


Make a place to work:

<!-- @makePlaceToWork @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

<!-- @defineManifest @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/manifest.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: controller-manager
  name: my-controller-manager
  namespace: my-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    metadata:
      labels:
        control-plane: controller-manager
    spec:
      containers:
      - args:
        - --secure-listen-address=0.0.0.0:8443
        - --upstream=http://127.0.0.1:8080/
        - --logtostderr=true
        - --v=10
        image: some-image
        name: some-image-deployment
        ports:
        - containerPort: 8443
EOF
```

<!-- @defineKustomization @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
commonLabels:
  some-label.site.com: "true"
resources:
- manifest.yaml
EOF
```

<!-- @definedExpectedOutput @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/out_expected.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: controller-manager
    some-label.site.com: "true"
  name: my-controller-manager
  namespace: my-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: controller-manager
      some-label.site.com: "true"
  template:
    metadata:
      labels:
        control-plane: controller-manager
        some-label.site.com: "true"
    spec:
      containers:
      - args:
        - --secure-listen-address=0.0.0.0:8443
        - --upstream=http://127.0.0.1:8080/
        - --logtostderr=true
        - --v=10
        image: some-image
        name: some-image-deployment
        ports:
        - containerPort: 8443
EOF
```

Run the build:

<!-- @runIt @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME >$DEMO_HOME/out_actual.yaml
```

and confirm that the actual output matches the expected output:

<!-- @diffShouldExitZero @testAgainstLatestRelease -->
```
diff $DEMO_HOME/out_actual.yaml $DEMO_HOME/out_expected.yaml
```
