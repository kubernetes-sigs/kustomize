# Demo: SpringBoot

This example illustrates the organization of K8s configuration files with
environment-specific overlays to customize the configuration for a SpringBoot
application.

In this example, we have organized the files as follows:

> ```
> ├── base                           # Base configs which will be further customized.
> │   ├── application.properties     # Used by the configMapGenerator in kustomization.yaml.
> │   ├── deployment.yaml            # K8s Deployment definition.
> │   ├── kustomization.yaml
> │   └── service.yaml               # K8s Service definition.
> │
> └── overlays
>     │
>     ├── staging                    # Staging variant of the base configs.
>     │   ├── application.properties # Customization of the base ConfigMap for staging.
>     │   ├── kustomization.yaml
>     │   └── staging.env            # Additional ConfigMap items.
>     │
>     └── production                 # Production variant of the base configs.
>         ├── application.properties
>         ├── healthcheck_patch.yaml # Adds liveness and readiness probes to the deployment.
>         ├── kustomization.yaml
>         ├── memorylimit_patch.yaml # Memory limits for containers in the deployment.
>         └── patch.yaml             # Customization of the base ConfigMap for production.
> ```


## Base Configuration
Note that the `configMapGenerator` element in `kustomization.yaml` allows externalizing the
contents of a `ConfigMap` item in a separate file, in this case in `application.properties`.
This capability makes it convenient to maintain the contents in a file format that is
appropriate for the content rather than having to inline the contents in the YAML definition
of the `ConfigMap`.

To build the base K8s configuration from these files:
<!-- @springBootBase @test -->
```
# Change current directory to the springboot example.
[[ -d examples/springboot ]] && cd examples/springboot

# Generate the K8s manifest.
kustomize build base
```
The output of `kustomize build base` can be piped to `kubectl apply -f -`
to apply the final configuration to the K8s cluster.

## Staging Variant
In the staging environment we want to:
- Customize the application configuration to allow accessing the staging database.
- Add a `ConfigMap` property `environment=staging` based on definition in an external file.
- Specify the K8s resource names to have `staging-` prefix.

To build the staging K8s configuration from these files:
<!-- @springBootStaging @test -->
```
# Change current directory to the springboot example.
[[ -d examples/springboot ]] && cd examples/springboot

kustomize build overlays/staging
```

## Production Variant
In the production environment we want to:
- Customize the application configuration to allow accessing the production database.
- Specify the K8s resource names to have `prod-` prefix and `env: prod` label.
- Specify container memory limits in the K8s configuration.
- Configure liveness and readiness probes.


We can verify the final result of all the above configurations by generating the
K8s manifest:
<!-- @springBootProd @test -->
```
# Change current directory to the springboot example.
[[ -d examples/springboot ]] && cd examples/springboot

kustomize build overlays/production
```

The individual configuration customizations in the `production` overlay are 
explained in further detail below.

### Customize ConfigMap
We want to add database credentials for the prod environment. In general, these credentials
can be put into the file `application.properties`.
However, for some cases, we want to keep the credentials in a different file and keep
application specific configs in `application.properties`. With this clear separation,
the credentials and application specific things can be managed and maintained flexibly
by different teams.
For example, application developers only tune the application configs in `application.properties`
and operation teams or SREs only care about the credentials.

For Spring Boot application, we can set an active profile through the environment variable
`spring.profiles.active`.
Then the application will pick up an extra `application-<profile>.properties` file.
With this, we can customize the configMap in two steps. Add an environment variable through
the patch and add a file to the configMap.

### Resource Name and Label Customization
We want resources in production environment to have names prefixed with `prod-` as well as
certain labels so that we can query them by label selector.
This is specified by the following element in the `kustomization.yaml`:
> ```yaml
> commonLabels:
>   env: prod
> ```

To verify that the name-prefix was applied correctly to all the resources:
<!-- @springBootProd @test -->
```
kustomize build overlays/production | grep -C3 -E '^  name:'
```

To verify that the `env: prod` label was applied correctly to all the resources:
<!-- @springBootProd @test -->
```
kustomize build overlays/production | grep -C3 -E '^    env:'
```

### Configuring JVM Memory Limit
When a Spring Boot application is deployed in a K8s cluster, the JVM is running inside a container.
We want to set memory limit for the container and make sure the JVM is aware of that limit. In K8s
deployment, we can set the resource limits for containers and inject these limits to some environment
variables by downward API. When the container starts to run, it can pick up the environment variables
and set JVM options accordingly.

The `memorylimit_patch.yaml` defines the container resource limits and then injects then into
environment variable named `MEM_TOTAL_MB`:
> ```yaml
>           env:
>           - name: MEM_TOTAL_MB
>             valueFrom:
>               resourceFieldRef:
>                 resource: limits.memory
> ```

### Liveness and Readiness Probes
We also want to add liveness check and readiness check in the production environment.
Spring Boot application has end-points such as `/actuator/health` for this. We can
customize the K8s deployment resource to talk to Spring Boot end point.
