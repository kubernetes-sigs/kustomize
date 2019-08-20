# Demo: SpringBoot

In this tutorial, you will learn - how to use `kustomize` to customize a basic Spring Boot application's
k8s configuration for production use cases.

In the production environment we want to customize the following:

- add application specific configuration for this Spring Boot application
- configure prod DB access configuration
- resource names to be prefixed by 'prod-'.
- resources to have 'env: prod' labels.
- JVM memory to be properly set.
- health check and readiness check.

First make a place to work:
<!-- @makeDemoHome @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### Download resources

To keep this document shorter, the base resources
needed to run springboot on a k8s cluster are off in a
supplemental data directory rather than declared here
as HERE documents.

Download them:

<!-- @downloadResources @testAgainstLatestRelease -->
```
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/springboot"

curl -s -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/base/{deployment,service}.yaml"
```

### Initialize kustomization.yaml

The `kustomize` program gets its instructions from
a file called `kustomization.yaml`.

Start this file:

<!-- @kustomizeYaml @testAgainstLatestRelease -->
```
touch $DEMO_HOME/kustomization.yaml
```

### Add the resources

<!-- @addResources @testAgainstLatestRelease -->
```
cd $DEMO_HOME

kustomize edit add resource service.yaml
kustomize edit add resource deployment.yaml

cat kustomization.yaml
```

`kustomization.yaml`'s resources section should contain:

> ```
> resources:
> - service.yaml
> - deployment.yaml
> ```

### Add configMap generator

<!-- @addConfigMap @testAgainstLatestRelease -->
```
echo "app.name=Kustomize Demo" >$DEMO_HOME/application.properties

kustomize edit add configmap demo-configmap \
  --from-file application.properties

cat kustomization.yaml
```

`kustomization.yaml`'s configMapGenerator section should contain:

> ```
> configMapGenerator:
> - files:
>   - application.properties
>   name: demo-configmap
> ```

### Customize configMap

We want to add database credentials for the prod environment. In general, these credentials can be put into the file `application.properties`.
However, for some cases, we want to keep the credentials in a different file and keep application specific configs in `application.properties`.
 With this clear separation, the credentials and application specific things can be managed and maintained flexibly by different teams.
For example, application developers only tune the application configs in `application.properties` and operation teams or SREs
only care about the credentials.

For Spring Boot application, we can set an active profile through the environment variable `spring.profiles.active`. Then
the application will pick up an extra `application-<profile>.properties` file. With this, we can customize the configMap in two
steps. Add an environment variable through the patch and add a file to the configMap.

<!-- @customizeConfigMap @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/patch.yaml
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: sbdemo
spec:
  template:
    spec:
      containers:
        - name: sbdemo
          env:
          - name: spring.profiles.active
            value: prod
EOF

kustomize edit add patch patch.yaml

cat <<EOF >$DEMO_HOME/application-prod.properties
spring.jpa.hibernate.ddl-auto=update
spring.datasource.url=jdbc:mysql://<prod_database_host>:3306/db_example
spring.datasource.username=root
spring.datasource.password=admin
EOF

kustomize edit add configmap \
  demo-configmap --from-file application-prod.properties

cat kustomization.yaml
```

`kustomization.yaml`'s configMapGenerator section should contain:
> ```
> configMapGenerator:
> - files:
>   - application.properties
>   - application-prod.properties
>   name: demo-configmap
> ```

### Name Customization

Arrange for the resources to begin with prefix
_prod-_ (since they are meant for the _production_
environment):

<!-- @customizeLabel @testAgainstLatestRelease -->
```
cd $DEMO_HOME
kustomize edit set nameprefix 'prod-'
```

`kustomization.yaml` should have updated value of namePrefix field:

> ```
> namePrefix: prod-
> ```

This `namePrefix` directive adds _prod-_ to all
resource names, as can be seen by building the
resources:

<!-- @build1 @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME | grep prod-
```

### Label Customization

We want resources in production environment to have
certain labels so that we can query them by label
selector.

`kustomize` does not have `edit set label` command to
add a label, but one can always edit
`kustomization.yaml` directly:

<!-- @customizeLabels @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
commonLabels:
  env: prod
EOF
```

Confirm that the resources now all have names prefixed
by `prod-` and the label tuple `env:prod`:

<!-- @build2 @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME | grep -C 3 env
```

### Download Patch for JVM memory

When a Spring Boot application is deployed in a k8s cluster, the JVM is running inside a container. We want to set memory limit for the container and make sure
the JVM is aware of that limit. In K8s deployment, we can set the resource limits for containers and inject these limits to
some environment variables by downward API. When the container starts to run, it can pick up the environment variables and
set JVM options accordingly.

Download the patch `memorylimit_patch.yaml`. It contains the memory limits setup.

<!-- @downloadPatch @testAgainstLatestRelease -->
```
curl -s  -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/overlays/production/{memorylimit_patch}.yaml"

cat $DEMO_HOME/memorylimit_patch.yaml
```

The output contains

> ```
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   name: sbdemo
> spec:
>   template:
>     spec:
>       containers:
>         - name: sbdemo
>           resources:
>             limits:
>               memory: 1250Mi
>             requests:
>               memory: 1250Mi
>           env:
>           - name: MEM_TOTAL_MB
>             valueFrom:
>               resourceFieldRef:
>                 resource: limits.memory
> ```

### Download Patch for health check
We also want to add liveness check and readiness check in the production environment. Spring Boot application
has end points such as `/actuator/health` for this. We can customize the k8s deployment resource to talk to Spring Boot end point.

Download the patch `healthcheck_patch.yaml`. It contains the liveness probes and readyness probes.

<!-- @downloadPatch @testAgainstLatestRelease -->
```
curl -s  -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/overlays/production/{healthcheck_patch}.yaml"

cat $DEMO_HOME/healthcheck_patch.yaml
```

The output contains

> ```
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   name: sbdemo
> spec:
>   template:
>     spec:
>       containers:
>         - name: sbdemo
>           livenessProbe:
>             httpGet:
>               path: /actuator/health
>               port: 8080
>             initialDelaySeconds: 10
>             periodSeconds: 3
>           readinessProbe:
>             initialDelaySeconds: 20
>             periodSeconds: 10
>             httpGet:
>               path: /actuator/info
>               port: 8080
> ```

### Add patches

Add these patches to the kustomization:

<!-- @addPatch @testAgainstLatestRelease -->
```
cd $DEMO_HOME
kustomize edit add patch memorylimit_patch.yaml
kustomize edit add patch healthcheck_patch.yaml
```

`kustomization.yaml` should have patches field:

> ```
> patchesStrategicMerge:
> - patch.yaml
> - memorylimit_patch.yaml
> - healthcheck_patch.yaml
> ```

The output of the following command can now be applied
to the cluster (i.e. piped to `kubectl apply`) to
create the production environment.

<!-- @finalBuild @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME  # | kubectl apply -f -
```
