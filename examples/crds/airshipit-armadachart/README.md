# Test CRD Register airshipit-armadachart


This folder the airshipit-armadachart Kustomize CRD Register

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
apiVersion: armada.airshipit.org/v1alpha1
kind: ArmadaChart
metadata:
  name: blog-1
spec:
  chart_name: blog-1
  release: blog-1
  namespace: default
  install:
    no_hooks: false
  upgrade:
    no_hooks: false
  values: {}
  source:
    type: local
    location: /opt/armada/helm-charts/testchart
    subpath: .
    reference: 87aad18f7d8c6a1a08f3adc8866efd33bee6aa52
  dependencies: []
  target_state: uninitialized
EOF
```


### Preparation Step Resource1

<!-- @createResource1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/transformer.yaml
apiVersion: armada.airshipit.org/v1alpha1
kind: ArmadaCRDRegister
metadata:
  name: armadacrdregister
EOF
```


### Preparation Step Resource2

<!-- @createResource2 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/patch.yaml
apiVersion: armada.airshipit.org/v1alpha1
kind: ArmadaChart
metadata:
  name: blog-1
spec:
  source:
    type: foobar
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
cat <<'EOF' >${DEMO_HOME}/expected/armada.airshipit.org_v1alpha1_armadachart_blog-1.yaml
apiVersion: armada.airshipit.org/v1alpha1
kind: ArmadaChart
metadata:
  name: blog-1
spec:
  chart_name: blog-1
  dependencies: []
  install:
    no_hooks: false
  namespace: default
  release: blog-1
  source:
    location: /opt/armada/helm-charts/testchart
    reference: 87aad18f7d8c6a1a08f3adc8866efd33bee6aa52
    subpath: .
    type: foobar
  target_state: uninitialized
  upgrade:
    no_hooks: false
  values: {}
EOF
```


<!-- @compareActualToExpected @test -->
```bash
test 0 == \
$(diff -r $DEMO_HOME/actual $DEMO_HOME/expected | wc -l); \
echo $?
```

