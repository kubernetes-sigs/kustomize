// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestBasicVariableRef(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
namePrefix: base-
resources:
- pod.yaml
vars:
- name: POD_NAME
  objref:
    apiVersion: v1
    kind: Pod
    name: clown
  fieldref:
    fieldpath: metadata.name
`)

	th.WriteF("/app/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: clown
spec:
  containers:
  - name: frown
    image: frown
    command:
    - echo
    - "$(POD_NAME)"
    env:
      - name: FOO
        value: "$(POD_NAME)"
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Pod
metadata:
  name: base-clown
spec:
  containers:
  - command:
    - echo
    - base-clown
    env:
    - name: FOO
      value: base-clown
    image: frown
    name: frown
`)
}

func TestBasicVarCollision(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base1", `
namePrefix: base1-
resources:
- pod.yaml
vars:
- name: POD_NAME
  objref:
    apiVersion: v1
    kind: Pod
    name: kelley
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base1/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: kelley
spec:
  containers:
  - name: smile
    image: smile
    command:
    - echo
    - "$(POD_NAME)"
    env:
      - name: FOO
        value: "$(POD_NAME)"
`)

	th.WriteK("/app/base2", `
namePrefix: base2-
resources:
- pod.yaml
vars:
- name: POD_NAME
  objref:
    apiVersion: v1
    kind: Pod
    name: grimaldi
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base2/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: grimaldi
spec:
  containers:
  - name: dance
    image: dance
    command:
    - echo
    - "$(POD_NAME)"
    env:
      - name: FOO
        value: "$(POD_NAME)"
`)

	th.WriteK("/app/overlay", `
resources:
- ../base1
- ../base2
`)
	err := th.RunWithErr("/app/overlay", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("should have an error")
	}
	if !strings.Contains(err.Error(), "var 'POD_NAME' already encountered") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestVarPropagatesUp(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base1", `
namePrefix: base1-
resources:
- pod.yaml
vars:
- name: POD_NAME1
  objref:
    apiVersion: v1
    kind: Pod
    name: kelley
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base1/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: kelley
spec:
  containers:
  - name: smile
    image: smile
    command:
    - echo
    - "$(POD_NAME1)"
    env:
      - name: FOO
        value: "$(POD_NAME1)"
`)

	th.WriteK("/app/base2", `
namePrefix: base2-
resources:
- pod.yaml
vars:
- name: POD_NAME2
  objref:
    apiVersion: v1
    kind: Pod
    name: grimaldi
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base2/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: grimaldi
spec:
  containers:
  - name: dance
    image: dance
    command:
    - echo
    - "$(POD_NAME2)"
    env:
      - name: FOO
        value: "$(POD_NAME2)"
`)

	th.WriteK("/app/overlay", `
resources:
- pod.yaml
- ../base1
- ../base2
`)
	th.WriteF("/app/overlay/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: circus
spec:
  containers:
  - name: ring
    image: ring
    command:
    - echo
    - "$(POD_NAME1)"
    - "$(POD_NAME2)"
    env:
      - name: P1
        value: "$(POD_NAME1)"
      - name: P2
        value: "$(POD_NAME2)"
`)
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Pod
metadata:
  name: circus
spec:
  containers:
  - command:
    - echo
    - base1-kelley
    - base2-grimaldi
    env:
    - name: P1
      value: base1-kelley
    - name: P2
      value: base2-grimaldi
    image: ring
    name: ring
---
apiVersion: v1
kind: Pod
metadata:
  name: base1-kelley
spec:
  containers:
  - command:
    - echo
    - base1-kelley
    env:
    - name: FOO
      value: base1-kelley
    image: smile
    name: smile
---
apiVersion: v1
kind: Pod
metadata:
  name: base2-grimaldi
spec:
  containers:
  - command:
    - echo
    - base2-grimaldi
    env:
    - name: FOO
      value: base2-grimaldi
    image: dance
    name: dance
`)
}

// Not so much a bug as a desire for local variables
// with less than global scope.  Currently all variables
// are global.  So if a base with a variable is included
// twice, it's a collision, so it's denied.
func TestBug506(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
namePrefix: base-
resources:
- pod.yaml
vars:
- name: POD_NAME
  objref:
    apiVersion: v1
    kind: Pod
    name: myServerPod
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: myServerPod
spec:
  containers:
    - name: myServer
      image: whatever
      env:
        - name: POD_NAME
          value: $(POD_NAME)
`)
	th.WriteK("/app/o1", `
nameprefix: p1-
resources:
- ../base
`)
	th.WriteK("/app/o2", `
nameprefix: p2-
resources:
- ../base
`)
	th.WriteK("/app/top", `
resources:
- ../o1
- ../o2
`)

	/*
	   	const presumablyDesired = `
	   apiVersion: v1
	   kind: Pod
	   metadata:
	     name: p1-base-myServerPod
	   spec:
	     containers:
	     - env:
	       - name: POD_NAME
	         value: p1-base-myServerPod
	       image: whatever
	       name: myServer
	   ---
	   apiVersion: v1
	   kind: Pod
	   metadata:
	     name: p2-base-myServerPod
	   spec:
	     containers:
	     - env:
	       - name: POD_NAME
	         value: p2-base-myServerPod
	       image: whatever
	       name: myServer
	   `
	*/
	err := th.RunWithErr("/app/top", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("should have an error")
	}
	if !strings.Contains(err.Error(), "var 'POD_NAME' already encountered") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestVarRefBig(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
namePrefix: base-
resources:
- role-stuff.yaml
- services.yaml
- statefulset.yaml
- cronjob.yaml
- pdb.yaml
configMapGenerator:
- name: test-config-map
  literals:
  - foo=bar
  - baz=qux
vars:
 - name: CDB_PUBLIC_SVC
   objref:
        kind: Service
        name: cockroachdb-public
        apiVersion: v1
   fieldref:
        fieldpath: metadata.name
 - name: CDB_STATEFULSET_NAME
   objref:
        kind: StatefulSet
        name: cockroachdb
        apiVersion: apps/v1beta1
   fieldref:
        fieldpath: metadata.name
 - name: CDB_HTTP_PORT
   objref:
        kind: StatefulSet
        name: cockroachdb
        apiVersion: apps/v1beta1
   fieldref:
        fieldpath: spec.template.spec.containers[0].ports[1].containerPort
 - name: CDB_STATEFULSET_SVC
   objref:
        kind: Service
        name: cockroachdb
        apiVersion: v1
   fieldref:
        fieldpath: metadata.name

 - name: TEST_CONFIG_MAP
   objref:
        kind: ConfigMap
        name: test-config-map
        apiVersion: v1
   fieldref:
        fieldpath: metadata.name`)
	th.WriteF("/app/base/cronjob.yaml", `
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: cronjob-example
spec:
  schedule: "*/1 * * * *"
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cronjob-example
            image: cockroachdb/cockroach:v1.1.5
            command:
            - echo
            - "$(CDB_STATEFULSET_NAME)"
            - "$(TEST_CONFIG_MAP)"
            env:
              - name: CDB_PUBLIC_SVC
                value: "$(CDB_PUBLIC_SVC)"
`)
	th.WriteF("/app/base/services.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: cockroachdb
  labels:
    app: cockroachdb
  annotations:
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
    # Enable automatic monitoring of all instances when Prometheus is running in the cluster.
    prometheus.io/scrape: "true"
    prometheus.io/path: "_status/vars"
    prometheus.io/port: "8080"
spec:
  ports:
  - port: 26257
    targetPort: 26257
    name: grpc
  - port: $(CDB_HTTP_PORT)
    targetPort: $(CDB_HTTP_PORT)
    name: http
  clusterIP: None
  selector:
    app: cockroachdb
---
apiVersion: v1
kind: Service
metadata:
  # This service is meant to be used by clients of the database. It exposes a ClusterIP that will
  # automatically load balance connections to the different database pods.
  name: cockroachdb-public
  labels:
    app: cockroachdb
spec:
  ports:
  # The main port, served by gRPC, serves Postgres-flavor SQL, internode
  # traffic and the cli.
  - port: 26257
    targetPort: 26257
    name: grpc
  # The secondary port serves the UI as well as health and debug endpoints.
  - port: $(CDB_HTTP_PORT)
    targetPort: $(CDB_HTTP_PORT)
    name: http
  selector:
    app: cockroachdb
`)
	th.WriteF("/app/base/role-stuff.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cockroachdb
  labels:
    app: cockroachdb
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: cockroachdb
  labels:
    app: cockroachdb
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - get
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cockroachdb
  labels:
    app: cockroachdb
rules:
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests
  verbs:
  - create
  - get
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: cockroachdb
  labels:
    app: cockroachdb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cockroachdb
subjects:
- kind: ServiceAccount
  name: cockroachdb
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cockroachdb
  labels:
    app: cockroachdb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cockroachdb
subjects:
- kind: ServiceAccount
  name: cockroachdb
  namespace: default
`)
	th.WriteF("/app/base/pdb.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: cockroachdb-budget
  labels:
    app: cockroachdb
spec:
  selector:
    matchLabels:
      app: cockroachdb
  maxUnavailable: 1
`)
	th.WriteF("/app/base/statefulset.yaml", `
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: cockroachdb
spec:
  serviceName: "cockroachdb"
  replicas: 3
  template:
    metadata:
      labels:
        app: cockroachdb
    spec:
      serviceAccountName: cockroachdb
      # Init containers are run only once in the lifetime of a pod, before
      # it's started up for the first time. It has to exit successfully
      # before the pod's main containers are allowed to start.
      initContainers:
      # The init-certs container sends a certificate signing request to the
      # kubernetes cluster.
      # You can see pending requests using: kubectl get csr
      # CSRs can be approved using:         kubectl certificate approve <csr name>
      #
      # All addresses used to contact a node must be specified in the --addresses arg.
      #
      # In addition to the node certificate and key, the init-certs entrypoint will symlink
      # the cluster CA to the certs directory.
      - name: init-certs
        image: cockroachdb/cockroach-k8s-request-cert:0.2
        imagePullPolicy: IfNotPresent
        command:
        - "/bin/ash"
        - "-ecx"
        - "/request-cert"
        - -namespace=${POD_NAMESPACE}
        - -certs-dir=/cockroach-certs
        - -type=node
        - -addresses=localhost,127.0.0.1,${POD_IP},$(hostname -f),$(hostname -f|cut -f 1-2 -d '.'),$(CDB_PUBLIC_SVC)
        - -symlink-ca-from=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - name: certs
          mountPath: /cockroach-certs

      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - cockroachdb
              topologyKey: kubernetes.io/hostname
      containers:
      - name: cockroachdb
        image: cockroachdb/cockroach:v1.1.5
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 26257
          name: grpc
        - containerPort: 8080
          name: http
        volumeMounts:
        - name: datadir
          mountPath: /cockroach/cockroach-data
        - name: certs
          mountPath: /cockroach/cockroach-certs
        command:
          - "/bin/bash"
          - "-ecx"
          - "exec /cockroach/cockroach start --logtostderr"
          - --certs-dir /cockroach/cockroach-certs
          - --host $(hostname -f)
          - --http-host 0.0.0.0
          - --join $(CDB_STATEFULSET_NAME)-0.$(CDB_STATEFULSET_SVC),$(CDB_STATEFULSET_NAME)-1.$(CDB_STATEFULSET_SVC),$(CDB_STATEFULSET_NAME)-2.$(CDB_STATEFULSET_SVC)
          - --cache 25%
          - --max-sql-memory 25%
      # No pre-stop hook is required, a SIGTERM plus some time is all that's
      # needed for graceful shutdown of a node.
      terminationGracePeriodSeconds: 60
      volumes:
      - name: datadir
        persistentVolumeClaim:
          claimName: datadir
      - name: certs
        emptyDir: {}
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes:
        - "ReadWriteOnce"
      resources:
        requests:
          storage: 1Gi
`)
	th.WriteK("/app/overlay/staging", `
namePrefix: dev-
resources:
- ../../base
`)
	m := th.Run("/app/overlay/staging", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - get
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb
rules:
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests
  verbs:
  - create
  - get
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: dev-base-cockroachdb
subjects:
- kind: ServiceAccount
  name: dev-base-cockroachdb
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: dev-base-cockroachdb
subjects:
- kind: ServiceAccount
  name: dev-base-cockroachdb
  namespace: default
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: _status/vars
    prometheus.io/port: "8080"
    prometheus.io/scrape: "true"
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb
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
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb-public
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
---
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: dev-base-cockroachdb
spec:
  replicas: 3
  serviceName: dev-base-cockroachdb
  template:
    metadata:
      labels:
        app: cockroachdb
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
        - exec /cockroach/cockroach start --logtostderr
        - --certs-dir /cockroach/cockroach-certs
        - --host $(hostname -f)
        - --http-host 0.0.0.0
        - --join dev-base-cockroachdb-0.dev-base-cockroachdb,dev-base-cockroachdb-1.dev-base-cockroachdb,dev-base-cockroachdb-2.dev-base-cockroachdb
        - --cache 25%
        - --max-sql-memory 25%
        image: cockroachdb/cockroach:v1.1.5
        imagePullPolicy: IfNotPresent
        name: cockroachdb
        ports:
        - containerPort: 26257
          name: grpc
        - containerPort: 8080
          name: http
        volumeMounts:
        - mountPath: /cockroach/cockroach-data
          name: datadir
        - mountPath: /cockroach/cockroach-certs
          name: certs
      initContainers:
      - command:
        - /bin/ash
        - -ecx
        - /request-cert
        - -namespace=${POD_NAMESPACE}
        - -certs-dir=/cockroach-certs
        - -type=node
        - -addresses=localhost,127.0.0.1,${POD_IP},$(hostname -f),$(hostname -f|cut -f 1-2 -d '.'),dev-base-cockroachdb-public
        - -symlink-ca-from=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: cockroachdb/cockroach-k8s-request-cert:0.2
        imagePullPolicy: IfNotPresent
        name: init-certs
        volumeMounts:
        - mountPath: /cockroach-certs
          name: certs
      serviceAccountName: dev-base-cockroachdb
      terminationGracePeriodSeconds: 60
      volumes:
      - name: datadir
        persistentVolumeClaim:
          claimName: datadir
      - emptyDir: {}
        name: certs
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
---
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: dev-base-cronjob-example
spec:
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - command:
            - echo
            - dev-base-cockroachdb
            - dev-base-test-config-map-6b85g79g7g
            env:
            - name: CDB_PUBLIC_SVC
              value: dev-base-cockroachdb-public
            image: cockroachdb/cockroach:v1.1.5
            name: cronjob-example
  schedule: '*/1 * * * *'
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: cockroachdb
  name: dev-base-cockroachdb-budget
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: cockroachdb
---
apiVersion: v1
data:
  baz: qux
  foo: bar
kind: ConfigMap
metadata:
  name: dev-base-test-config-map-6b85g79g7g
`)
}

func TestVariableRefIngress(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
resources:
- service.yaml
- deployment.yaml
- ingress.yaml

vars:
- name: DEPLOYMENT_NAME
  objref:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app.kubernetes.io/component: nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/component: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.15.7-alpine
        ports:
        - name: http
          containerPort: 80
`)
	th.WriteF("/app/base/ingress.yaml", `
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: nginx
  labels:
    app.kubernetes.io/component: nginx
spec:
  rules:
  - host: $(DEPLOYMENT_NAME).example.com
    http:
      paths:
      - backend:
          serviceName: nginx
          servicePort: 80
        path: /
  tls:
  - hosts:
    - $(DEPLOYMENT_NAME).example.com
    secretName: $(DEPLOYMENT_NAME).example.com-tls
`)
	th.WriteF("/app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app.kubernetes.io/component: nginx
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: http
`)
	th.WriteK("/app/overlay", `
nameprefix: kustomized-
resources:
- ../base
`)
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: nginx
  name: kustomized-nginx
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: http
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: nginx
  name: kustomized-nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/component: nginx
    spec:
      containers:
      - image: nginx:1.15.7-alpine
        name: nginx
        ports:
        - containerPort: 80
          name: http
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  labels:
    app.kubernetes.io/component: nginx
  name: kustomized-nginx
spec:
  rules:
  - host: kustomized-nginx.example.com
    http:
      paths:
      - backend:
          serviceName: kustomized-nginx
          servicePort: 80
        path: /
  tls:
  - hosts:
    - kustomized-nginx.example.com
    secretName: kustomized-nginx.example.com-tls
`)
}

func TestVariableRefMountPath(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
resources:
- deployment.yaml
- namespace.yaml

vars:
- name: NAMESPACE
  objref:
    apiVersion: v1
    kind: Namespace
    name: my-namespace

`)
	th.WriteF("/app/base/deployment.yaml", `
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: my-deployment
  spec:
    template:
      spec:
        containers:
        - name: app
          image: busybox
          volumeMounts:
          - name: my-volume
            mountPath: "/$(NAMESPACE)"
          env:
          - name: NAMESPACE
            value: $(NAMESPACE)
        volumes:
        - name: my-volume
          emptyDir: {}
`)
	th.WriteF("/app/base/namespace.yaml", `
  apiVersion: v1
  kind: Namespace
  metadata:
    name: my-namespace
`)

	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - env:
        - name: NAMESPACE
          value: my-namespace
        image: busybox
        name: app
        volumeMounts:
        - mountPath: /my-namespace
          name: my-volume
      volumes:
      - emptyDir: {}
        name: my-volume
---
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
`)
}

func TestVariableRefMaps(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
resources:
- deployment.yaml
- namespace.yaml
vars:
- name: NAMESPACE
  objref:
    apiVersion: v1
    kind: Namespace
    name: my-namespace
`)
	th.WriteF("/app/base/deployment.yaml", `
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: my-deployment
    labels:
      my-label: $(NAMESPACE)
    annotations:
      my-annotation: $(NAMESPACE)
  spec:
    template:
      spec:
        containers:
        - name: app
          image: busybox
`)
	th.WriteF("/app/base/namespace.yaml", `
  apiVersion: v1
  kind: Namespace
  metadata:
    name: my-namespace
`)

	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    my-annotation: my-namespace
  labels:
    my-label: my-namespace
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: app
---
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
`)
}

func TestVaribaleRefDifferentPrefix(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
namePrefix: base-
resources:
- dev
- test
`)

	th.WriteK("/app/base/dev", `
namePrefix: dev-
resources:
- elasticsearch-dev-service.yml
vars:
- name: elasticsearch-dev-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name

`)
	th.WriteF("/app/base/dev/elasticsearch-dev-service.yml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
spec:
  template:
    spec:
      containers:
        - name: elasticsearch
          env:
            - name: DISCOVERY_SERVICE
              value: "$(elasticsearch-dev-service-name).monitoring.svc.cluster.local"
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
spec:
  ports:
    - name: transport
      port: 9300
      protocol: TCP
  clusterIP: None
`)

	th.WriteK("/app/base/test", `
namePrefix: test-
resources:
- elasticsearch-test-service.yml
vars:
- name: elasticsearch-test-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base/test/elasticsearch-test-service.yml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
spec:
  template:
    spec:
      containers:
        - name: elasticsearch
          env:
            - name: DISCOVERY_SERVICE
              value: "$(elasticsearch-test-service-name).monitoring.svc.cluster.local"
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
spec:
  ports:
    - name: transport
      port: 9300
      protocol: TCP
  clusterIP: None
`)

	m := th.Run("/app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: base-dev-elasticsearch
spec:
  template:
    spec:
      containers:
      - env:
        - name: DISCOVERY_SERVICE
          value: base-dev-elasticsearch.monitoring.svc.cluster.local
        name: elasticsearch
---
apiVersion: v1
kind: Service
metadata:
  name: base-dev-elasticsearch
spec:
  clusterIP: None
  ports:
  - name: transport
    port: 9300
    protocol: TCP
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: base-test-elasticsearch
spec:
  template:
    spec:
      containers:
      - env:
        - name: DISCOVERY_SERVICE
          value: base-test-elasticsearch.monitoring.svc.cluster.local
        name: elasticsearch
---
apiVersion: v1
kind: Service
metadata:
  name: base-test-elasticsearch
spec:
  clusterIP: None
  ports:
  - name: transport
    port: 9300
    protocol: TCP
`)
}

func TestVariableRefNFSServer(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
resources:
- pv_pvc.yaml
- nfs_deployment.yaml
- nfs_service.yaml
- Deployment.yaml
- CronJob.yaml
- DaemonSet.yaml
- ReplicaSet.yaml
- StatefulSet.yaml
- Pod.yaml
- Job.yaml
- nfs_pv.yaml

vars:
- name: NFS_SERVER_SERVICE_NAME
  objref:
    kind: Service
    name: nfs-server-service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
`)
	th.WriteF("/app/base/pv_pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: shared-volume-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
	th.WriteF("/app/base/nfs_deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nfs-server
spec:
  replicas: 1
  template:
    spec:
      metadata:
        labels:
          role: nfs-server
      containers:
        - name: nfs-server
          image: gcr.io/google_containers/volume-nfs:0.8
          ports:
            - name: nfs
              containerPort: 2049
            - name: mountd
              containerPort: 20048
            - name: rpcbind
              containerPort: 111
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /exports
              name: shared-files
      volumes:
        - name: shared-files
          persistentVolumeClaim:
            claimName: shared-volume-claim
`)
	th.WriteF("/app/base/nfs_service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: nfs-server-service
spec:
  ports:
    - name: nfs
      port: 2049
    - name: mountd
      port: 20048
    - name: rpcbind
      port: 111
  selector:
    role: nfs-server
`)
	th.WriteF("/app/base/Deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app.kubernetes.io/component: nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/component: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.15.7-alpine
        ports:
        - name: http
          containerPort: 80
        volumeMounts:
          - mountPath: /app/shared-files
            name: nfs-files-vol
      volumes:
        - name: nfs-files-vol
          nfs:
            server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
            path: /
            readOnly: false
`)
	th.WriteF("/app/base/CronJob.yaml", `
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
          volumeMounts:
          - mountPath: /app/shared-files
            name: nfs-files-vol
        volumes:
        - name: nfs-files-vol
          nfs:
            server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
            path: /
            readOnly: false
`)
	th.WriteF("/app/base/DaemonSet.yaml", `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd-elasticsearch
        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - mountPath: /app/shared-files
          name: nfs-files-vol
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: nfs-files-vol
        nfs:
          server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
          path: /
          readOnly: false
`)
	th.WriteF("/app/base/ReplicaSet.yaml", `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: frontend
  labels:
    app: guestbook
    tier: frontend
spec:
  # modify replicas according to your case
  replicas: 3
  selector:
    matchLabels:
      tier: frontend
  template:
    metadata:
      labels:
        tier: frontend
    spec:
      containers:
      - name: php-redis
        image: gcr.io/google_samples/gb-frontend:v3
        volumeMounts:
        - mountPath: /app/shared-files
          name: nfs-files-vol
      volumes:
      - name: nfs-files-vol
        nfs:
          server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
          path: /
          readOnly: false
`)

	th.WriteF("/app/base/Job.yaml", `
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
        volumeMounts:
        - mountPath: /app/shared-files
          name: nfs-files-vol
      restartPolicy: Never
      volumes:
      - name: nfs-files-vol
        nfs:
          server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
          path: /
          readOnly: false
  backoffLimit: 4
`)
	th.WriteF("/app/base/StatefulSet.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteMany" ]
      nfs:
        server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
        path: /
        readOnly: false
`)
	th.WriteF("/app/base/Pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: nginx
    image: nginx:1.15.7-alpine
    ports:
    - name: http
      containerPort: 80
  volumeMounts:
  - name: nfs-files-vol
    mountPath: /app/shared-files
  volumes:
  - name: nfs-files-vol
    nfs:
      server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
      path: /
      readOnly: false
`)
	th.WriteF("/app/base/nfs_pv.yaml", `
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-files-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  nfs:
    server: $(NFS_SERVER_SERVICE_NAME).default.srv.cluster.local
    path: /
    readOnly: false
`)
	th.WriteK("/app/overlay", `
nameprefix: kustomized-
resources:
- ../base
`)
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: kustomized-shared-volume-claim
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kustomized-nfs-server
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: gcr.io/google_containers/volume-nfs:0.8
        name: nfs-server
        ports:
        - containerPort: 2049
          name: nfs
        - containerPort: 20048
          name: mountd
        - containerPort: 111
          name: rpcbind
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /exports
          name: shared-files
      metadata:
        labels:
          role: nfs-server
      volumes:
      - name: shared-files
        persistentVolumeClaim:
          claimName: kustomized-shared-volume-claim
---
apiVersion: v1
kind: Service
metadata:
  name: kustomized-nfs-server-service
spec:
  ports:
  - name: nfs
    port: 2049
  - name: mountd
    port: 20048
  - name: rpcbind
    port: 111
  selector:
    role: nfs-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: nginx
  name: kustomized-nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/component: nginx
    spec:
      containers:
      - image: nginx:1.15.7-alpine
        name: nginx
        ports:
        - containerPort: 80
          name: http
        volumeMounts:
        - mountPath: /app/shared-files
          name: nfs-files-vol
      volumes:
      - name: nfs-files-vol
        nfs:
          path: /
          readOnly: false
          server: kustomized-nfs-server-service.default.srv.cluster.local
---
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: kustomized-hello
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
            image: busybox
            name: hello
          restartPolicy: OnFailure
          volumeMounts:
          - mountPath: /app/shared-files
            name: nfs-files-vol
        volumes:
        - name: nfs-files-vol
          nfs:
            path: /
            readOnly: false
            server: kustomized-nfs-server-service.default.srv.cluster.local
  schedule: '*/1 * * * *'
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: fluentd-logging
  name: kustomized-fluentd-elasticsearch
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      containers:
      - image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
        name: fluentd-elasticsearch
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - mountPath: /var/log
          name: varlog
        - mountPath: /var/lib/docker/containers
          name: varlibdockercontainers
          readOnly: true
        - mountPath: /app/shared-files
          name: nfs-files-vol
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      volumes:
      - hostPath:
          path: /var/log
        name: varlog
      - hostPath:
          path: /var/lib/docker/containers
        name: varlibdockercontainers
      - name: nfs-files-vol
        nfs:
          path: /
          readOnly: false
          server: kustomized-nfs-server-service.default.srv.cluster.local
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  labels:
    app: guestbook
    tier: frontend
  name: kustomized-frontend
spec:
  replicas: 3
  selector:
    matchLabels:
      tier: frontend
  template:
    metadata:
      labels:
        tier: frontend
    spec:
      containers:
      - image: gcr.io/google_samples/gb-frontend:v3
        name: php-redis
        volumeMounts:
        - mountPath: /app/shared-files
          name: nfs-files-vol
      volumes:
      - name: nfs-files-vol
        nfs:
          path: /
          readOnly: false
          server: kustomized-nfs-server-service.default.srv.cluster.local
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: kustomized-web
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  serviceName: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: k8s.gcr.io/nginx-slim:0.8
        name: nginx
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - mountPath: /usr/share/nginx/html
          name: www
      terminationGracePeriodSeconds: 10
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes:
      - ReadWriteMany
      nfs:
        path: /
        readOnly: false
        server: kustomized-nfs-server-service.default.srv.cluster.local
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
  name: kustomized-myapp-pod
spec:
  containers:
  - image: nginx:1.15.7-alpine
    name: nginx
    ports:
    - containerPort: 80
      name: http
  volumeMounts:
  - mountPath: /app/shared-files
    name: nfs-files-vol
  volumes:
  - name: nfs-files-vol
    nfs:
      path: /
      readOnly: false
      server: kustomized-nfs-server-service.default.srv.cluster.local
---
apiVersion: batch/v1
kind: Job
metadata:
  name: kustomized-pi
spec:
  backoffLimit: 4
  template:
    spec:
      containers:
      - command:
        - perl
        - -Mbignum=bpi
        - -wle
        - print bpi(2000)
        image: perl
        name: pi
        volumeMounts:
        - mountPath: /app/shared-files
          name: nfs-files-vol
      restartPolicy: Never
      volumes:
      - name: nfs-files-vol
        nfs:
          path: /
          readOnly: false
          server: kustomized-nfs-server-service.default.srv.cluster.local
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: kustomized-nfs-files-pv
spec:
  accessModes:
  - ReadWriteMany
  capacity:
    storage: 10Gi
  nfs:
    path: /
    readOnly: false
    server: kustomized-nfs-server-service.default.srv.cluster.local
`)
}

func TestDeploymentAnnotations(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

configMapGenerator:
  - name: testConfigMap
    envs:
      - test.properties

vars:
  - name: FOO
    objref:
      kind: ConfigMap
      name: testConfigMap
      apiVersion: v1
    fieldref:
      fieldpath: data.foo

commonAnnotations:
  foo: $(FOO)

resources:
  - deployment.yaml
`)

	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
spec:
  template:
    spec:
      containers:
        - name: test
`)
	th.WriteF("/app/test.properties", `foo=bar`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    foo: bar
  name: test
spec:
  template:
    metadata:
      annotations:
        foo: bar
    spec:
      containers:
      - name: test
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    foo: bar
  name: testConfigMap-798k5k7g9f
`)
}
