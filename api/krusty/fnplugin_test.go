package krusty_test

import (
	"os/exec"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestFnExecGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
resources:
- short_secret.yaml
generators:
- gener.yaml
`)

	// Create some additional resource just to make sure everything is added
	th.WriteF("/app/short_secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)

	th.WriteF("/app/gener.yaml", `
kind: executable
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./fnplugin_test/fnexectest.sh
spec:
`)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true
	m := th.Run("/app", o)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/path: deployment_nginx.yaml
    tshirt-size: small
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func skipIfNoDocker(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping because docker binary wasn't found in PATH")
	}
}

func TestFnContainerGenerator(t *testing.T) {
	skipIfNoDocker(t)

	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
resources:
- short_secret.yaml
generators:
- gener.yaml
`)
	// Create generator config
	th.WriteF("/app/gener.yaml", `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: CockroachDB
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
spec:
  replicas: 3
`)
	// Create some additional resource just to make sure everything is added
	th.WriteF("/app/short_secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  annotations:
    config.kubernetes.io/path: config/demo-budget_poddisruptionbudget.yaml
  labels:
    app: cockroachdb
    name: demo
  name: demo-budget
spec:
  minAvailable: 67%
  selector:
    matchLabels:
      app: cockroachdb
      name: demo
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/path: config/demo-public_service.yaml
  labels:
    app: cockroachdb
    name: demo
  name: demo-public
spec:
  ports:
  - name: grpc
    port: 26257
    targetPort: 26257
  - name: http
    port: 8080
    targetPort: 8080
  selector:
    app: cockroachdb
    name: demo
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/path: config/demo_service.yaml
    prometheus.io/path: _status/vars
    prometheus.io/port: "8080"
    prometheus.io/scrape: "true"
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
  labels:
    app: cockroachdb
    name: demo
  name: demo
spec:
  clusterIP: None
  ports:
  - name: grpc
    port: 26257
    targetPort: 26257
  - name: http
    port: 8080
    targetPort: 8080
  selector:
    app: cockroachdb
    name: demo
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  annotations:
    config.kubernetes.io/path: config/demo_statefulset.yaml
  labels:
    app: cockroachdb
    name: demo
  name: demo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cockroachdb
      name: demo
  serviceName: demo
  template:
    metadata:
      labels:
        app: cockroachdb
        name: demo
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - cockroachdb
              topologyKey: kubernetes.io/hostname
            weight: 100
      containers:
      - command:
        - /bin/bash
        - -ecx
        - |
          # The use of qualified `+"`hostname -f`"+` is crucial:
          # Other nodes aren't able to look up the unqualified hostname.
          CRARGS=("start" "--logtostderr" "--insecure" "--host" "$(hostname -f)" "--http-host" "0.0.0.0")
          # We only want to initialize a new cluster (by omitting the join flag)
          # if we're sure that we're the first node (i.e. index 0) and that
          # there aren't any other nodes running as part of the cluster that
          # this is supposed to be a part of (which indicates that a cluster
          # already exists and we should make sure not to create a new one).
          # It's fine to run without --join on a restart if there aren't any
          # other nodes.
          if [ ! "$(hostname)" == "cockroachdb-0" ] ||              [ -e "/cockroach/cockroach-data/cluster_exists_marker" ]
          then
            # We don't join cockroachdb in order to avoid a node attempting
            # to join itself, which currently doesn't work
            # (https://github.com/cockroachdb/cockroach/issues/9625).
            CRARGS+=("--join" "cockroachdb-public")
          fi
          exec /cockroach/cockroach ${CRARGS[*]}
        image: cockroachdb/cockroach:v1.1.0
        imagePullPolicy: IfNotPresent
        name: demo
        ports:
        - containerPort: 26257
          name: grpc
        - containerPort: 8080
          name: http
        volumeMounts:
        - mountPath: /cockroach/cockroach-data
          name: datadir
      initContainers:
      - args:
        - -on-start=/on-start.sh
        - -service=cockroachdb
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: cockroachdb/cockroach-k8s-init:0.1
        imagePullPolicy: IfNotPresent
        name: bootstrap
        volumeMounts:
        - mountPath: /cockroach/cockroach-data
          name: datadir
      terminationGracePeriodSeconds: 60
      volumes:
      - name: datadir
        persistentVolumeClaim:
          claimName: datadir
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
`)
}

func TestFnContainerTransformer(t *testing.T) {
	skipIfNoDocker(t)

	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
resources:
- data.yaml
transformers:
- transf1.yaml
- transf2.yaml
`)

	th.WriteF("/app/data.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
  annotations:
    tshirt-size: small # this injects the resource reservations
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
`)
	// This transformer should add resource reservations based on annotation in data.yaml
	// See https://github.com/kubernetes-sigs/kustomize/tree/master/functions/examples/injection-tshirt-sizes
	th.WriteF("/app/transf1.yaml", `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: Validator
metadata:
  name: valid
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: gcr.io/kustomize-functions/example-tshirt:v0.2.0
`)
	// This transformer will check resources without and won't do any changes
	// See https://github.com/kubernetes-sigs/kustomize/tree/master/functions/examples/validator-kubeval
	th.WriteF("/app/transf2.yaml", `
apiVersion: examples.config.kubernetes.io/v1beta1
kind: Kubeval
metadata:
  name: validate
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/example-validator-kubeval:v0.1.0
spec:
  strict: true
  ignoreMissingSchemas: true

  # TODO: Update this to use network/volumes features.
  # Relevant issues:
  #   - https://github.com/kubernetes-sigs/kustomize/issues/1901
  #   - https://github.com/kubernetes-sigs/kustomize/issues/1902
  kubernetesVersion: "1.16.0"
  schemaLocation: "file:///schemas"
`)
	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/path: deployment_nginx.yaml
    tshirt-size: small
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        resources:
          requests:
            cpu: 200m
            memory: 50M
`)
}

func TestFnContainerTransformerWithConfig(t *testing.T) {
	skipIfNoDocker(t)

	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
resources:
- data1.yaml
- data2.yaml
transformers:
- label_namespace.yaml
`)

	th.WriteF("/app/data1.yaml", `apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
`)
	th.WriteF("/app/data2.yaml", `apiVersion: v1
kind: Namespace
metadata:
  name: another-namespace
`)

	th.WriteF("/app/label_namespace.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  name: label_namespace
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: gcr.io/kpt-functions/label-namespace@sha256:4f030738d6d25a207641ca517916431517578bd0eb8d98a8bde04e3bb9315dcd
data:
  label_name: my-ns-name
  label_value: function-test
`)

	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    config.kubernetes.io/path: namespace_my-namespace.yaml
  labels:
    my-ns-name: function-test
  name: my-namespace
---
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    config.kubernetes.io/path: namespace_another-namespace.yaml
  labels:
    my-ns-name: function-test
  name: another-namespace
`)
}

func TestFnContainerEnvVars(t *testing.T) {
	skipIfNoDocker(t)

	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
generators:
- gener.yaml
`)

	// TODO: cheange image to gcr.io/kpt-functions/templater:stable
	// when https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/103
	// is merged
	th.WriteF("/app/gener.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/kpt-templater:0.0.1
        envs:
        - TESTTEMPLATE=value
data:
  template: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: env
    data:
      value: '{{ env "TESTTEMPLATE" }}'
`)
	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  value: value
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/path: configmap_env.yaml
  name: env
`)
}
