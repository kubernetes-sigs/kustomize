# Test CRD Register istio-networking


This folder the istio-networking Kustomize CRD Register

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

namespace: kubeflow

resources:
- ./resources.yaml

transformers:
- ./transformer.yaml

configurations:
- ./kustomizeconfig.yaml

configMapGenerator:
- name: istio-parameters
  env: params.env

vars:
- name: clusterRbacConfig
  objref:
    kind: ConfigMap
    name: istio-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterRbacConfig
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
# - ./patch.yaml
EOF
```


### Preparation Step KustomizeConfig0

<!-- @createKustomizeConfig0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/kustomizeconfig.yaml
varReference:
- path: spec/mode
  kind: ClusterRbacConfig
EOF
```


### Preparation Step Resource0

<!-- @createResource0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/resources.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kubeflow-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: grafana-vs
spec:
  hosts:
  - "*"
  gateways:
  - "kubeflow-gateway"
  http:
  - match:
    - uri:
        prefix: "/istio/grafana/"
      method:
        exact: "GET"
    rewrite:
      uri: "/"
    route:
    - destination:
        host: "grafana.istio-system.svc.cluster.local"
        port:
          number: 3000
---
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: google-api-entry
spec:
  hosts:
  - www.googleapis.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  resolution: DNS
  location: MESH_EXTERNAL
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: google-api-vs
spec:
  hosts:
  - www.googleapis.com
  tls:
  - match:
    - port: 443
      sni_hosts:
      - www.googleapis.com
    route:
    - destination:
        host: www.googleapis.com
        port:
          number: 443
      weight: 100
---
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: google-storage-api-entry
spec:
  hosts:
  - storage.googleapis.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  resolution: DNS
  location: MESH_EXTERNAL
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: google-storage-api-vs
spec:
  hosts:
  - storage.googleapis.com
  tls:
  - match:
    - port: 443
      sni_hosts:
      - storage.googleapis.com
    route:
    - destination:
        host: storage.googleapis.com
        port:
          number: 443
      weight: 100
---
apiVersion: rbac.istio.io/v1alpha1
kind: ClusterRbacConfig
metadata:
  name: default
spec:
  mode: $(clusterRbacConfig)
EOF
```


### Preparation Step Resource1

<!-- @createResource1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/transformer.yaml
apiVersion: networking.istio.io/v1alpha3
kind: NetworkingCRDRegister
metadata:
  name: networkingcrdregister
EOF
```


### Preparation Step Resource2

<!-- @createResource2 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/patch.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kubeflow-gateway
spec:
  selector:
    foo: bar
EOF
```


### Preparation Step Other0

<!-- @createOther0 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/params.env
clusterRbacConfig=ON
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
cat <<'EOF' >${DEMO_HOME}/expected/~g_v1_configmap_istio-parameters-9m7m4fhdb8.yaml
apiVersion: v1
data:
  clusterRbacConfig: "ON"
kind: ConfigMap
metadata:
  name: istio-parameters-9m7m4fhdb8
  namespace: kubeflow
EOF
```


### Verification Step Expected1

<!-- @createExpected1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/networking.istio.io_v1alpha3_gateway_kubeflow-gateway.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kubeflow-gateway
  namespace: kubeflow
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - '*'
    port:
      name: http
      number: 80
      protocol: HTTP
EOF
```


### Verification Step Expected2

<!-- @createExpected2 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/networking.istio.io_v1alpha3_serviceentry_google-api-entry.yaml
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: google-api-entry
  namespace: kubeflow
spec:
  hosts:
  - www.googleapis.com
  location: MESH_EXTERNAL
  ports:
  - name: https
    number: 443
    protocol: HTTPS
  resolution: DNS
EOF
```


### Verification Step Expected3

<!-- @createExpected3 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/networking.istio.io_v1alpha3_serviceentry_google-storage-api-entry.yaml
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: google-storage-api-entry
  namespace: kubeflow
spec:
  hosts:
  - storage.googleapis.com
  location: MESH_EXTERNAL
  ports:
  - name: https
    number: 443
    protocol: HTTPS
  resolution: DNS
EOF
```


### Verification Step Expected4

<!-- @createExpected4 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/networking.istio.io_v1alpha3_virtualservice_google-api-vs.yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: google-api-vs
  namespace: kubeflow
spec:
  hosts:
  - www.googleapis.com
  tls:
  - match:
    - port: 443
      sni_hosts:
      - www.googleapis.com
    route:
    - destination:
        host: www.googleapis.com
        port:
          number: 443
      weight: 100
EOF
```


### Verification Step Expected5

<!-- @createExpected5 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/networking.istio.io_v1alpha3_virtualservice_google-storage-api-vs.yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: google-storage-api-vs
  namespace: kubeflow
spec:
  hosts:
  - storage.googleapis.com
  tls:
  - match:
    - port: 443
      sni_hosts:
      - storage.googleapis.com
    route:
    - destination:
        host: storage.googleapis.com
        port:
          number: 443
      weight: 100
EOF
```


### Verification Step Expected6

<!-- @createExpected6 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/networking.istio.io_v1alpha3_virtualservice_grafana-vs.yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: grafana-vs
  namespace: kubeflow
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - method:
        exact: GET
      uri:
        prefix: /istio/grafana/
    rewrite:
      uri: /
    route:
    - destination:
        host: grafana.istio-system.svc.cluster.local
        port:
          number: 3000
EOF
```


### Verification Step Expected7

<!-- @createExpected7 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/expected/rbac.istio.io_v1alpha1_clusterrbacconfig_default.yaml
apiVersion: rbac.istio.io/v1alpha1
kind: ClusterRbacConfig
metadata:
  name: default
  namespace: kubeflow
spec:
  mode: "ON"
EOF
```


<!-- @compareActualToExpected @test -->
```bash
test 0 == \
$(diff -r $DEMO_HOME/actual $DEMO_HOME/expected | wc -l); \
echo $?
```

