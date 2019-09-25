# Test CRD Register argo-rollout


This folder the argo-rollout Kustomize CRD Register

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

resources:
- ./rollout.yaml

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
cat <<'EOF' >${DEMO_HOME}/base/rollout.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: rollout
spec:
  template:
    spec:
      containers:
      - name: guestbook
        image: gcr.io/heptio-images/ks-guestbook-demo:0.1
        env:
        - name: test
          value: foo
EOF
```


### Preparation Step Resource1

<!-- @createResource1 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/base/transformer.yaml
apiVersion: rollouts.argoproj.io/v1alpha1
kind: RolloutCRDRegister
metadata:
  name: rolloutcrdregister
EOF
```


### Preparation Step Resource2

<!-- @createResource2 @test -->
```bash
cat <<'EOF' >${DEMO_HOME}/overlay/patch.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: rollout
spec:
  template:
    spec:
      containers:
      - name: guestbook
        env:
        - name: test
          value: bar
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
cat <<'EOF' >${DEMO_HOME}/expected/argoproj.io_v1alpha1_rollout_rollout.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: rollout
spec:
  template:
    spec:
      containers:
      - env:
        - name: test
          value: bar
        image: gcr.io/heptio-images/ks-guestbook-demo:0.1
        name: guestbook
EOF
```


<!-- @compareActualToExpected @test -->
```bash
test 0 == \
$(diff -r $DEMO_HOME/actual $DEMO_HOME/expected | wc -l); \
echo $?
```

