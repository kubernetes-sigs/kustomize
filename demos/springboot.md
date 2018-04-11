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

### Download resources

Download `deployment.yaml`, `service.yaml`. These are plain k8s resources files one
could add to a k8s cluster to run sbdemo.

<!-- @makeSpringBootDir @test -->
```

DEMO_HOME=$(mktemp -d)
cd $DEMO_HOME

CONTENT=https://raw.githubusercontent.com/kinflate

# Get SpringBoot configs
for f in service deployment; do \
  wget -q $CONTENT/example-springboot/master/$f.yaml ; \
done
```

### Initialize kustomize.yaml

The `kustomize` program gets its instructions from
a file called `kustomize.yaml`.

Start this file:

<!-- @kustomizeYaml @test -->
```
touch $DEMO_HOME/kustomize.yaml
```

### Add the resources

<!-- @addResources @test -->
```
cd $DEMO_HOME

kustomize edit add resource service.yaml
kustomize edit add resource deployment.yaml

cat kustomize.yaml
```

`kustomize.yaml`'s resources section should contain:

> ```
> resources:
> - service.yaml
> - deployment.yaml
> ```

### Add configmap generator

<!-- @addConfigMap @test -->
```
cd $DEMO_HOME
wget -q $CONTENT/example-springboot/master/application.properties
kustomize edit add configmap demo-configmap --from-file application.properties

cat kustomize.yaml
```

`kustomize.yaml`'s configMapGenerator section should contain:

> ```
> configMapGenerator:
> - files:
>   - application.properties
>   name: demo-configmap
> ```

### Customize configmap
We want to add database credentials for the prod environment. In general, these credentials can be put into the file `application.properties`.
However, for some cases, we want to keep the credentials in a different file and keep application specific configs in `application.properties`.
 With this clear separation, the credentials and application specific things can be managed and maintained flexibly by different teams.
For example, application developers only tune the application configs in `application.properties` and operation teams or SREs
only care about the credentials.

For Spring Boot application, we can set an active profile through the environment variable `spring.profiles.active`. Then
the application will pick up an extra `application-<profile>.properties` file. With this, we can customize the configmap in two
steps. Add an environment variable through the patch and add a file to the configmap.

<!-- @customizeConfigMap @test -->
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

cat <<EOF >>$DEMO_HOME/kustomize.yaml
patches:
- patch.yaml
EOF

cat <<EOF >$DEMO_HOME/application-prod.properties
spring.jpa.hibernate.ddl-auto=update
spring.datasource.url=jdbc:mysql://<prod_database_host>:3306/db_example
spring.datasource.username=root
spring.datasource.password=admin
EOF

kustomize edit add configmap demo-configmap --from-file application-prod.properties

cat kustomize.yaml
```

`kustomize.yaml`'s configMapGenerator section should contain:
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

<!-- @customizeLabel @test -->
```
cd $DEMO_HOME

kustomize edit set nameprefix 'prod-'

cat kustomize.yaml
```

`kustomize.yaml` should have updated value of namePrefix field:

> ```
> namePrefix: prod-
> objectAnnotations:
>  note: This is a example annotation
> ```

This `namePrefix` directive adds _prod-_ to all
resource names.

<!-- @genNamePrefixConfig @test -->
```
kustomize build $DEMO_HOME
```

The output should contain:

> ```
> apiVersion: v1
> data:
>   ....
> kind: ConfigMap
> metadata:
>   ....
>   name: prod-demo-configmap-7746248cmc
> ---
> apiVersion: v1
> kind: Service
> metadata:
>   ....
>   name: prod-sbdemo
> spec:
>   ....
> ---
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   ....
>   name: prod-sbdemo
> spec:
>   selector:
>     ....
> ```

### Label Customization

We want resources in production environment to have
certain labels so that we can query them by label
selector.

`kustomize` does not have `edit set label` command to add
label, but we can edit `kustomize.yaml` file under
`prod` directory and add the production labels under
`objectLabels` fields as highlighted below.

<!-- @customizeLabels @test -->
```
sed -i 's/app: helloworld/app: prod/' \
    $DEMO_HOME/kustomize.yaml
```

At this point, running `kustomize build` will
generate MySQL configs with name-prefix 'prod-' and
labels `env:prod`.


### Download Patch for JVM memory
When a Spring Boot application is deployed in a k8s cluster, the JVM is running inside a container. We want to set memory limit for the container and make sure
the JVM is aware of that limit. In K8s deployment, we can set the resource limits for containers and inject these limits to
some environment variables by downward API. When the container starts to run, it can pick up the environment variables and
set JVM options accordingly.

Download the patch `memorylimit_patch.yaml`. It contains the memory limits setup.

<!-- @downloadPatch @test -->
```
cd $DEMO_HOME
wget -q $CONTENT/example-springboot-instances/master/production/memorylimit_patch.yaml

cat memorylimit_patch.yaml
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

<!-- @downloadPatch @test -->
```
cd $DEMO_HOME
wget -q $CONTENT/example-springboot-instances/master/production/healthcheck_patch.yaml

cat healthcheck_patch.yaml
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

Currently `kustomize` doesn't provide a command to add a file as a patch, but we can edit the file `kustomize.yaml` to
include this patch.

<!-- @addPatch @test -->
```
mv $DEMO_HOME/kustomize.yaml $DEMO_HOME/tmp.yaml

sed '/patches:$/{N;s/- patch.yaml/- patch.yaml\n- memorylimit_patch.yaml\n- healthcheck_patch.yaml/}' \
    $DEMO_HOME/tmp.yaml >& $DEMO_HOME/kustomize.yaml
```

`kustomize.yaml` should have patches field:

> ```
> patches
> - patch.yaml
> - memorylimit_patch.yaml
> - healthcheck_patch.yaml
> ```

The output of the following command can now be applied
to the cluster (i.e. piped to `kubectl apply`) to
create the production environment.

<!-- @finalBuild @test -->
```
kustomize build $DEMO_HOME  # | kubectl apply -f -
```
