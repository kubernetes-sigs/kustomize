[hello]: https://github.com/monopole/hello
[kind]: https://github.com/kubernetes-sigs/kind
[helloWorld]: https://github.com/kubernetes-sigs/kustomize/tree/master/examples/helloWorld

# Demo: hello app

This demo helps you to deploy an example hello app end-to-end using kustomize.

Steps:
1. Create the resources files.
2. Kustomize them.
3. Spin-up kubernetes cluster on local using [kind].
4. Deploy the app using kustomize and verify the status.

First define a place to work:

<!-- @makeWorkplace @testE2EAgainstLatestRelease-->
```
DEMO_HOME=$(mktemp -d)
```

Alternatively, use

> ```
> DEMO_HOME=~/hello
> ```

## Establish the base

Let's run the [hello] service.

<!-- @createBase @testE2EAgainstLatestRelease-->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE
OUTPUT=$DEMO_HOME/output
mkdir -p $OUTPUT
```

Now lets add a simple config map resource to the `base`

<!-- @createConfigMapYaml @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/configMap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
EOF
```

Create `deployment.yaml` with any image and with desired number of replicas

<!-- @createDeploymentYaml @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 3
  template:
    metadata:
      labels:
        deployment: hello
    spec:
      containers:
      - name: the-container
        image: monopole/hello:1
        command: ["/hello",
                  "--port=8080",
                  "--enableRiskyFeature=\$(ENABLE_RISKY)"]
        ports:
        - containerPort: 8080
        env:
        - name: ALT_GREETING
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: altGreeting
        - name: ENABLE_RISKY
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: enableRisky
EOF
```

Create `service.yaml` pointing to the deployment created above

<!-- @createServiceYaml @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: the-service
spec:
  selector:
    deployment: hello
  type: LoadBalancer
  ports:
  - protocol: TCP
    port: 8666
    targetPort: 8080
EOF
```

Create a `grouping.yaml` resource. By this, you are defining the grouping of the current directory, `base`. Kustomize uses the unique label in this file to track any future state changes made to this directory. Make sure the label key is `kustomize.config.k8s.io/inventory-id` and give any unique label value and DO NOT change it in future.

<!-- @createGroupingYaml @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/grouping.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: inventory-map
  labels:
    kustomize.config.k8s.io/inventory-id: hello-app
EOF
```

Now, create `kustomization.yaml` add all your resources.

<!-- @createKustomizationYaml @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/kustomization.yaml
commonLabels:
  app: hello

resources:
- deployment.yaml
- service.yaml
- configMap.yaml
- grouping.yaml
EOF
```

### The Base Kustomization

The `base` directory has a kustomization file:

<!-- @showKustomization @testE2EAgainstLatestRelease -->
```
more $BASE/kustomization.yaml
```

### Customize the base

A simple customization step could be to change the _app
label_ applied to all resources:

<!-- @addLabel @testE2EAgainstLatestRelease -->
```
sed -i.bak 's/app: hello/app: my-hello/' \
    $BASE/kustomization.yaml
```

The following requires installation of [kind].

Delete any existing kind cluster and create a new one. By default the name of the cluster is "kind"
<!-- @deleteAndCreateKindCluster @testE2EAgainstLatestRelease -->
```
kind delete cluster
kind create cluster
```

Use the kustomize binary in MYGOBIN to apply a deployment, fetch the status and verify the status.
<!-- @runHelloApp @testE2EAgainstLatestRelease -->
```
export KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true

kustomize resources apply $BASE --status;

kustomize status fetch $BASE > $OUTPUT/status

test 1 == \
  $(grep "the-deployment" $OUTPUT/status | grep "Deployment is available. Replicas: 3" | wc -l); \
  echo $?

test 1 == \
  $(grep "the-map" $OUTPUT/status | grep "Resource is always ready" | wc -l); \
  echo $?

test 1 == \
  $(grep "the-service" $OUTPUT/status | grep "Service is ready" | wc -l); \
  echo $?
```

Now let's replace the configMap with configMap2 apply the config, fetch and verify the status. This should delete the-map from deployment and add the-map2.
<!-- @replaceConfigMapInHello @testE2EAgainstLatestRelease -->
```
cat <<EOF >$BASE/configMap2.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map2
data:
  altGreeting: "Good Evening!"
  enableRisky: "false"
EOF

rm $BASE/configMap.yaml

sed -i.bak 's/configMap/configMap2/' \
    $BASE/kustomization.yaml

sed -i.bak 's/the-map/the-map2/' \
    $BASE/deployment.yaml

kustomize resources apply $BASE --status;

status=$(mktemp);
kustomize status fetch $BASE > $OUTPUT/status

test 1 == \
  $(grep "the-deployment" $OUTPUT/status | grep "Deployment is available. Replicas: 3" | wc -l); \
  echo $?

test 1 == \
  $(grep "the-map2" $OUTPUT/status | grep "Resource is always ready" | wc -l); \
  echo $?

test 1 == \
  $(grep "the-service" $OUTPUT/status | grep "Service is ready" | wc -l); \
  echo $?
```

Clean-up the cluster 
<!-- @deleteKindCluster @testE2EAgainstLatestRelease -->
```
kind delete cluster
```

### Next Exercise
Create overlays as described in the [helloWorld] section and verify the results.
