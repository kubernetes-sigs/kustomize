# Test CRD Register kubeflow-kfdef


This folder the kubeflow-kfdef Kustomize CRD Register

## Setup the workspace

First, define a place to work:

<!-- @makeWorkplace @test -->
```bash
DEMO_HOME=$(mktemp -d)
```

## Preparation

<!-- @makeDirectories @test -->
```bash
mkdir -p ${DEMO_HOME}//home/kubedge/src/sigs.k8s.io/kustomize/examples/crds/kubeflow-kfdef
mkdir -p ${DEMO_HOME}/base
mkdir -p ${DEMO_HOME}/overlay
```

### Preparation Step KustomizationFile0

<!-- @createKustomizationFile0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/kustomization.yaml
resources:
- ./resources.yaml

transformers:
- ./transformer.yaml
EOF
```


### Preparation Step KustomizationFile1

<!-- @createKustomizationFile1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/kustomization.yaml
resources:
- ../base

patchesStrategicMerge:
- ./patch.yaml
EOF
```


### Preparation Step Resource0

<!-- @createResource0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/resources.yaml
# This is the config to install Kubeflow on an existing k8s cluster.
# If the cluster already has istio, comment out the istio install part below.

apiVersion: kfdef.apps.kubeflow.org/v1beta1
kind: KfDef
metadata:
  name: k8s-istio
  namespace: kubeflow
spec:
  repos:
  - name: manifests
    root: manifests-master
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  - name: kubeflow
    root: kubeflow-master
    uri: https://github.com/kubeflow/kubeflow/archive/master.tar.gz
  applications:
  # Istio install. If not needed, comment out istio-crds and istio-install.
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  # This component is the istio resources for Kubeflow (e.g. gateway), not about installing istio.
  - kustomizeConfig:
      parameters:
        - name: clusterRbacConfig
          value: "OFF"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: metacontroller
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-db
    name: katib-db
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-manager
    name: katib-manager
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-ui
    name: katib-ui # Issue: https://github.com/kubeflow/manifests/issues/151
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/metrics-collector
    name: metrics-collector
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/suggestion
    name: suggestion
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      parameters:
      - initRequired: true
        name: usageId
        value: <randomly-generated-id>
      - initRequired: true
        name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - istio
      parameters:
      - initRequired: true
        name: admin
        value: johnDoe@acme.com
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  enableApplications: true
  packageManager: kustomize
  skipInitProject: true
  useBasicAuth: false
  useIstio: true
  version: master
EOF
```


### Preparation Step Resource1

<!-- @createResource1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/transformer.yaml
apiVersion: kfdef.kubeflow.org/v1beta1
kind: KfDefCRDRegister
metadata:
  name: kfdefcrdregister
EOF
```


### Preparation Step Resource2

<!-- @createResource2 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/patch.yaml
apiVersion: kfdef.apps.kubeflow.org/v1beta1
kind: KfDef
metadata:
  name: k8s-istio
  namespace: kubeflow
spec:
  applications:
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: patchednamespace
    name: istio-install
  - kustomizeConfig:
      repoRef:
        name: patchedname
        path: patchedpath
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - addbeforeistio
      - istio
    name: centraldashboard
  - kustomizeConfig:
      overlays:
      - patchedoverlay1
      - patchedoverlay2
    name: webhook
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
cat <<'EOF' >${DEMO_HOME}/expected/kfdef.apps.kubeflow.org_v1beta1_kfdef_k8s-istio.yaml
apiVersion: kfdef.apps.kubeflow.org/v1beta1
kind: KfDef
metadata:
  name: k8s-istio
  namespace: kubeflow
spec:
  applications:
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: istio-system
      repoRef:
        name: manifests
        path: istio/istio-crds
    name: istio-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: patchednamespace
      repoRef:
        name: manifests
        path: istio/istio-install
    name: istio-install
  - kustomizeConfig:
      parameters:
      - name: clusterRbacConfig
        value: "OFF"
      repoRef:
        name: manifests
        path: istio/istio
    name: istio
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: application/application-crds
    name: application-crds
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: application/application
    name: application
  - kustomizeConfig:
      repoRef:
        name: patchedname
        path: patchedpath
    name: metacontroller
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: argo
    name: argo
  - kustomizeConfig:
      overlays:
      - addbeforeistio
      - istio
      repoRef:
        name: manifests
        path: common/centraldashboard
    name: centraldashboard
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: admission-webhook/bootstrap
    name: bootstrap
  - kustomizeConfig:
      overlays:
      - patchedoverlay1
      - patchedoverlay2
      repoRef:
        name: manifests
        path: admission-webhook/webhook
    name: webhook
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/jupyter-web-app
    name: jupyter-web-app
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-db
    name: katib-db
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-manager
    name: katib-manager
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-controller
    name: katib-controller
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: katib-v1alpha2/katib-ui
    name: katib-ui
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: metadata
    name: metadata
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/metrics-collector
    name: metrics-collector
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: katib-v1alpha2/suggestion
    name: suggestion
  - kustomizeConfig:
      overlays:
      - istio
      - application
      repoRef:
        name: manifests
        path: jupyter/notebook-controller
    name: notebook-controller
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-job-crds
    name: pytorch-job-crds
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pytorch-job/pytorch-operator
    name: pytorch-operator
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-crds
    name: knative-crds
  - kustomizeConfig:
      parameters:
      - name: namespace
        value: knative-serving
      repoRef:
        name: manifests
        path: knative/knative-serving-install
    name: knative-install
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kfserving/kfserving-crds
    name: kfserving-crds
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: kfserving/kfserving-install
    name: kfserving-install
  - kustomizeConfig:
      parameters:
      - initRequired: true
        name: usageId
        value: <randomly-generated-id>
      - initRequired: true
        name: reportUsage
        value: "true"
      repoRef:
        name: manifests
        path: common/spartakus
    name: spartakus
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tensorboard
    name: tensorboard
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: tf-training/tf-job-operator
    name: tf-job-operator
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/api-service
    name: api-service
  - kustomizeConfig:
      parameters:
      - name: minioPvcName
        value: minio-pv-claim
      repoRef:
        name: manifests
        path: pipeline/minio
    name: minio
  - kustomizeConfig:
      parameters:
      - name: mysqlPvcName
        value: mysql-pv-claim
      repoRef:
        name: manifests
        path: pipeline/mysql
    name: mysql
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/persistent-agent
    name: persistent-agent
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/pipelines-runner
    name: pipelines-runner
  - kustomizeConfig:
      overlays:
      - istio
      repoRef:
        name: manifests
        path: pipeline/pipelines-ui
    name: pipelines-ui
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/pipelines-viewer
    name: pipelines-viewer
  - kustomizeConfig:
      repoRef:
        name: manifests
        path: pipeline/scheduledworkflow
    name: scheduledworkflow
  - kustomizeConfig:
      overlays:
      - istio
      parameters:
      - initRequired: true
        name: admin
        value: johnDoe@acme.com
      repoRef:
        name: manifests
        path: profiles
    name: profiles
  - kustomizeConfig:
      overlays:
      - application
      repoRef:
        name: manifests
        path: seldon/seldon-core-operator
    name: seldon-core-operator
  enableApplications: true
  packageManager: kustomize
  repos:
  - name: manifests
    root: manifests-master
    uri: https://github.com/kubeflow/manifests/archive/master.tar.gz
  - name: kubeflow
    root: kubeflow-master
    uri: https://github.com/kubeflow/kubeflow/archive/master.tar.gz
  skipInitProject: true
  useBasicAuth: false
  useIstio: true
  version: master
EOF
```


<!-- @compareActualToExpected @test -->
```bash
test 0 == \
$(diff -r $DEMO_HOME/actual $DEMO_HOME/expected | wc -l); \
echo $?
```

