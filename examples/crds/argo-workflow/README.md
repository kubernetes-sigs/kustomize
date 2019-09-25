# Test CRD Register argo-workflow


This folder the argo-workflow Kustomize CRD Register

## Setup the workspace

First, define a place to work:

<!-- @makeWorkplace @test -->
```bash
DEMO_HOME=$(mktemp -d)
```

## Preparation

<!-- @makeDirectories @test -->
```bash
mkdir -p ${DEMO_HOME}/
mkdir -p ${DEMO_HOME}/base
mkdir -p ${DEMO_HOME}/overlay
```

### Preparation Step KustomizationFile0

<!-- @createKustomizationFile0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namePrefix: daemon-

resources:
- ./workflow.yaml

transformers:
- transformer.yaml
EOF
```


### Preparation Step KustomizationFile1

<!-- @createKustomizationFile1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../base

patchesStrategicMerge:
- ./patch.yaml
EOF
```


### Preparation Step Resource0

<!-- @createResource0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/transformer.yaml
apiVersion: argo.argoproj.io/v1alpha1
kind: WorkflowCRDRegister
metadata:
  name: workflowcrdregister
EOF
```


### Preparation Step Resource1

<!-- @createResource1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/workflow.yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: nginx
spec:
  entrypoint: daemon-nginx-example

  templates:
  - name: daemon-nginx-example
    steps:
    - - name: nginx-server
        template: nginx-server
    - - name: nginx-client
        template: nginx-client
        arguments:
          parameters:
          - name: server-ip
            value: "{{steps.nginx-server.ip}}"

  - name: nginx-server
    daemon: true
    container:
      image: nginx:1.13
      readinessProbe:
        httpGet:
          path: /
          port: 80
        initialDelaySeconds: 2
        timeoutSeconds: 1

  - name: nginx-client
    inputs:
      parameters:
      - name: server-ip
    container:
      image: appropriate/curl:latest
      command: ["/bin/sh", "-c"]
      args: ["echo curl --silent -G http://{{inputs.parameters.server-ip}}:80/ && curl --silent -G http://{{inputs.parameters.server-ip}}:80/"]
EOF
```


### Preparation Step Resource2

<!-- @createResource2 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/patch.yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: nginx
spec:
  templates:
  - name: nginx-server
    container:
      image: nginx:latest
EOF
```

## Execution

<!-- @build @test -->
```bash
mkdir ${DEMO_HOME}/actual
kustomize build ${DEMO_HOME}/overlay -o ${DEMO_HOME}/actual --enable_alpha_plugins
```

## Verification

<!-- @createExpectedDir @test -->
```bash
mkdir ${DEMO_HOME}/expected
```


### Verification Step Expected0

<!-- @createExpected0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/argoproj.io_v1alpha1_workflow_daemon-nginx.yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: daemon-nginx
spec:
  entrypoint: daemon-nginx-example
  templates:
  - name: daemon-nginx-example
    steps:
    - - name: nginx-server
        template: nginx-server
    - - arguments:
          parameters:
          - name: server-ip
            value: '{{steps.nginx-server.ip}}'
        name: nginx-client
        template: nginx-client
  - container:
      image: nginx:latest
      readinessProbe:
        httpGet:
          path: /
          port: 80
        initialDelaySeconds: 2
        timeoutSeconds: 1
    daemon: true
    name: nginx-server
  - container:
      args:
      - echo curl --silent -G http://{{inputs.parameters.server-ip}}:80/ && curl --silent
        -G http://{{inputs.parameters.server-ip}}:80/
      command:
      - /bin/sh
      - -c
      image: appropriate/curl:latest
    inputs:
      parameters:
      - name: server-ip
    name: nginx-client
EOF
```


<!-- @compareActualToExpected @test -->
```bash
test 0 == \
$(diff -r $DEMO_HOME/actual $DEMO_HOME/expected | wc -l); \
echo $?
```

