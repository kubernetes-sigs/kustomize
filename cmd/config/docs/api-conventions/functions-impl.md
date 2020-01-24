# Running Configuration Functions using kustomize CLI

Configuration functions can be implemented using any toolchain and invoked using any
container workflow orchestrator including Tekton, Cloud Build, or run directly using `docker run`.

Run `config help docs-fn-spec` to see the Configuration Functions Specification.

`kustomize config run` is an example orchestrator for invoking Configuration Functions. This
document describes how to implement and invoke an example function.

## Example Function Implementation

Following is an example for implementing an nginx abstraction using a configuration
function.

### `nginx-template.sh`

`nginx-template.sh` is a simple bash script which uses a _heredoc_ as a templating solution
for generating Resources from the functionConfig input fields.

The script wraps itself using `config run wrap -- $0` which will:

1. Parse the `ResourceList.functionConfig` (provided to the container stdin) into env vars
2. Merge the stdout into the original list of Resources
3. Defaults filenames for newly generated Resources (if they are not set as annotations)
   to `config/NAME_KIND.yaml`
4. Format the output

```bash
#!/bin/bash
# script must run wrapped by "kustomize config run wrap"
# for parsing input the functionConfig into env vars
if [ -z ${WRAPPED} ]; then
  export WRAPPED=true
  config run wrap -- $0
  exit $?
fi

cat <<End-of-message
apiVersion: v1
kind: Service
metadata:
  name: ${NAME}
  labels:
    app: nginx
    instance: ${NAME}
spec:
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: nginx
    instance: ${NAME}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${NAME}
  labels:
    app: nginx
    instance: ${NAME}
spec:
  replicas: ${REPLICAS}
  selector:
    matchLabels:
      app: nginx
      instance: ${NAME}
  template:
    metadata:
      labels:
        app: nginx
        instance: ${NAME}
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
End-of-message
```

### Dockerfile

`Dockerfile` installs `kustomize config` and copies the script into the container image.

```
FROM golang:1.13-stretch
RUN go get sigs.k8s.io/kustomize/cmd/config
RUN mv /go/bin/config /usr/bin/config
COPY nginx-template.sh /usr/bin/nginx-template.sh
CMD ["nginx-template.sh]
```

## Example Function Usage

Following is an example of running the `kustomize config run` using the preceding API.

When run by `kustomize config run`, functions are run in containers with the
following environment:

- Network: `none`
- User: `nobody`
- Security Options: `no-new-privileges`
- Volumes: the volume containing the `functionConfig` yaml is mounted under `/local` as `ro`

### Input

`dir/nginx.yaml` contains a reference to the Function. The contents of `nginx.yaml`
are passed to the Function through the `ResourceList.functionConfig` field.

```yaml
apiVersion: example.com/v1beta1
kind: Nginx
metadata:
  name: my-instance
  annotations:
    config.kubernetes.io/local-config: "true"
    config.k8s.io/function: |
      container:
        image: gcr.io/example-functions/nginx-template:v1.0.0
spec:
  replicas: 5
```

- `annotations[config.k8s.io/function].container.image`: the image to use for this API
- `annotations[config.kubernetes.io/local-config]`: mark this as not a Resource that should
  be applied

### Output

The function is invoked using byrunning `kustomize config run dir/`.

`dir/my-instance_deployment.yaml` contains the Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-instance
  labels:
    app: nginx
    instance: my-instance
spec:
  replicas: 5
  selector:
    matchLabels:
      app: nginx
      instance: my-instance
  template:
    metadata:
      labels:
        app: nginx
        instance: my-instance
    spec:
      containers:
        - name: nginx
          image: nginx:1.7.9
          ports:
            - containerPort: 80
```

`dir/my-instance_service.yaml` contains the Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-instance
  labels:
    app: nginx
    instance: my-instance
spec:
  ports:
    - port: 80
      targetPort: 80
      name: http
  selector:
    app: nginx
    instance: my-instance
```
