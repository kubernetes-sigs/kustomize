// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const expectedHelm = `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: test-minecraft
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: test-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
`

func TestHelmChartInflationGeneratorOld(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmChartInflationGenerator:
- chartName: minecraft
  chartRepoUrl: https://itzg.github.io/minecraft-server-charts
  chartVersion: 3.1.3
  releaseName: test
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelm)
}

func TestHelmChartInflationGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmCharts:
- name: minecraft
  repo: https://itzg.github.io/minecraft-server-charts
  version: 3.1.3
  releaseName: test
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelm)
}

// Last mile helm - show how kustomize puts helm charts into different
// namespaces with different customizations.
func TestHelmChartProdVsDev(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	dirBase := th.MkDir("base")
	dirProd := th.MkDir("prod")
	dirDev := th.MkDir("dev")
	dirBoth := th.MkDir("both")

	th.WriteK(dirBase, `
helmCharts:
- name: minecraft
  repo: https://itzg.github.io/minecraft-server-charts
  version: 3.1.3
  releaseName: test
`)
	th.WriteK(dirProd, `
namespace: prod
namePrefix: myProd-
resources:
- ../base
`)
	th.WriteK(dirDev, `
namespace: dev
namePrefix: myDev-
resources:
- ../base
`)
	th.WriteK(dirBoth, `
resources:
- ../dev
- ../prod
`)

	// Base unchanged
	m := th.Run(dirBase, th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, expectedHelm)

	// Prod has a "prod" namespace and a prefix.
	m = th.Run(dirProd, th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
`)

	// Both has two namespaces.
	m = th.Run(dirBoth, th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myDev-test-minecraft
  namespace: dev
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myDev-test-minecraft
  namespace: dev
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
---
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: myProd-test-minecraft
  namespace: prod
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
`)
}

func TestHelmChartInflationGeneratorWithNamespace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t)
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	th.WriteK(th.GetRoot(), `
helmChartInflationGenerator:
- chartName: nats
  chartVersion: v0.13.1
  chartRepoUrl: https://nats-io.github.io/k8s/helm/charts
  releaseName: nats
  releaseNamespace: test
`)

	m := th.Run(th.GetRoot(), th.MakeOptionsPluginsEnabled())

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  nats.conf: |
    # NATS Clients Port
    port: 4222

    # PID file shared with configuration reloader.
    pid_file: "/var/run/nats/nats.pid"

    ###############
    #             #
    # Monitoring  #
    #             #
    ###############
    http: 8222
    server_name:$POD_NAME
    lame_duck_duration: 120s
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/instance: nats
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nats
    app.kubernetes.io/version: 2.7.2
    helm.sh/chart: nats-0.13.1
  name: nats-config
  namespace: test
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/instance: nats
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nats
    app.kubernetes.io/version: 2.7.2
    helm.sh/chart: nats-0.13.1
  name: nats
  namespace: test
spec:
  clusterIP: None
  ports:
  - name: client
    port: 4222
  - name: cluster
    port: 6222
  - name: monitor
    port: 8222
  - name: metrics
    port: 7777
  - name: leafnodes
    port: 7422
  - name: gateways
    port: 7522
  selector:
    app.kubernetes.io/instance: nats
    app.kubernetes.io/name: nats
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nats-box
    chart: nats-0.13.1
  name: nats-box
  namespace: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nats-box
  template:
    metadata:
      labels:
        app: nats-box
    spec:
      containers:
      - command:
        - tail
        - -f
        - /dev/null
        env:
        - name: NATS_URL
          value: nats
        image: natsio/nats-box:0.8.1
        imagePullPolicy: IfNotPresent
        name: nats-box
        resources: null
        volumeMounts: null
      volumes: null
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app.kubernetes.io/instance: nats
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nats
    app.kubernetes.io/version: 2.7.2
    helm.sh/chart: nats-0.13.1
  name: nats
  namespace: test
spec:
  podManagementPolicy: Parallel
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: nats
      app.kubernetes.io/name: nats
  serviceName: nats
  template:
    metadata:
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "7777"
        prometheus.io/scrape: "true"
      labels:
        app.kubernetes.io/instance: nats
        app.kubernetes.io/name: nats
    spec:
      containers:
      - command:
        - nats-server
        - --config
        - /etc/nats-config/nats.conf
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: SERVER_NAME
          value: $(POD_NAME)
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: CLUSTER_ADVERTISE
          value: $(POD_NAME).nats.$(POD_NAMESPACE).svc.cluster.local
        image: nats:2.7.2-alpine
        imagePullPolicy: IfNotPresent
        lifecycle:
          preStop:
            exec:
              command:
              - /bin/sh
              - -c
              - nats-server -sl=ldm=/var/run/nats/nats.pid
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: 8222
          initialDelaySeconds: 10
          periodSeconds: 60
          successThreshold: 1
          timeoutSeconds: 5
        name: nats
        ports:
        - containerPort: 4222
          name: client
        - containerPort: 7422
          name: leafnodes
        - containerPort: 7522
          name: gateways
        - containerPort: 6222
          name: cluster
        - containerPort: 8222
          name: monitor
        - containerPort: 7777
          name: metrics
        resources: {}
        startupProbe:
          failureThreshold: 30
          httpGet:
            path: /
            port: 8222
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
        volumeMounts:
        - mountPath: /etc/nats-config
          name: config-volume
        - mountPath: /var/run/nats
          name: pid
      - command:
        - nats-server-config-reloader
        - -pid
        - /var/run/nats/nats.pid
        - -config
        - /etc/nats-config/nats.conf
        image: natsio/nats-server-config-reloader:0.6.2
        imagePullPolicy: IfNotPresent
        name: reloader
        resources: null
        volumeMounts:
        - mountPath: /etc/nats-config
          name: config-volume
        - mountPath: /var/run/nats
          name: pid
      - args:
        - -connz
        - -routez
        - -subz
        - -varz
        - -prefix=nats
        - -use_internal_server_id
        - http://localhost:8222/
        image: natsio/prometheus-nats-exporter:0.9.1
        imagePullPolicy: IfNotPresent
        name: metrics
        ports:
        - containerPort: 7777
          name: metrics
        resources: {}
      shareProcessNamespace: true
      terminationGracePeriodSeconds: 120
      volumes:
      - configMap:
          name: nats-config
        name: config-volume
      - emptyDir: {}
        name: pid
  volumeClaimTemplates: null
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    helm.sh/hook: test
  labels:
    app.kubernetes.io/instance: nats
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nats
    app.kubernetes.io/version: 2.7.2
    helm.sh/chart: nats-0.13.1
  name: nats-test-request-reply
spec:
  containers:
  - command:
    - /bin/sh
    - -ec
    - |
      nats reply -s nats://$NATS_HOST:4222 'name.>' --command "echo 1" &
    - |
      "&&"
    - |
      name=$(nats request -s nats://$NATS_HOST:4222 name.test '' 2>/dev/null)
    - |
      "&&"
    - |
      [ $name = test ]
    env:
    - name: NATS_HOST
      value: nats
    image: synadia/nats-box
    name: nats-box
  restartPolicy: Never
`)
}
