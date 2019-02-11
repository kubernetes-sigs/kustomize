/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package target_test

import (
	"testing"
)

// TODO(monopole): Add a feature test example covering secret generation.

// WARNING: These tests use a fake file system, and any attempt to use a
// feature that spawns shells will fail, because said shells expect a working
// directory corresponding to a real directory on disk - see
// these lines in secretfactory.go:
//   cmd := exec.CommandContext(ctx, commands[0], commands[1:]...)
//	 cmd.Dir = f.wd
// Worse, the fake directory might match a real directory on the your system,
// making the failure less obvious (and maybe hurting something if your secret
// generation technique writes data to disk).  So no use of secret generation
// in these particular tests.
// To eventually fix this, we could write the data to a real filesystem, and
// clean up after, or use some other trick compatible with exec.

func writeMediumBase(th *KustTestHarness) {
	th.writeK("/app/base", `
namePrefix: baseprefix-
commonLabels:
  foo: bar
commonAnnotations:
  baseAnno: This is a base annotation
resources:
- deployment/deployment.yaml
- service/service.yaml
`)
	th.writeF("/app/base/service/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: mungebot-service
  labels:
    app: mungebot
spec:
  ports:
    - port: 7002
  selector:
    app: mungebot
`)
	th.writeF("/app/base/deployment/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - name: nginx
        image: nginx
        env:
        - name: foo
          value: bar
        ports:
        - containerPort: 80
`)
}

func TestMediumBase(t *testing.T) {
	th := NewKustTestHarness(t, "/app/base")
	writeMediumBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: baseprefix-mungebot-service
spec:
  ports:
  - port: 7002
  selector:
    app: mungebot
    foo: bar
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: baseprefix-mungebot
spec:
  replicas: 1
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      annotations:
        baseAnno: This is a base annotation
      labels:
        app: mungebot
        foo: bar
    spec:
      containers:
      - env:
        - name: foo
          value: bar
        image: nginx
        name: nginx
        ports:
        - containerPort: 80
`)
}

func TestMediumOverlay(t *testing.T) {
	th := NewKustTestHarness(t, "/app/overlay")
	writeMediumBase(th)
	th.writeK("/app/overlay", `
namePrefix: test-infra-
commonLabels:
  app: mungebot
  org: kubernetes
  repo: test-infra
commonAnnotations:
  note: This is a test annotation
bases:
- ../base
patchesStrategicMerge:
- deployment/deployment.yaml
configMapGenerator:
- name: app-env
  env: configmap/app.env
- name: app-config
  files:
  - configmap/app-init.ini
images:
- name: nginx
  newTag: 1.8.0`)

	th.writeF("/app/overlay/configmap/app.env", `
DB_USERNAME=admin
DB_PASSWORD=somepw
`)
	th.writeF("/app/overlay/configmap/app-init.ini", `
FOO=bar
BAR=baz
`)
	th.writeF("/app/overlay/deployment/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: mungebot
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        env:
        - name: FOO
          valueFrom:
            configMapKeyRef:
              name: app-env
              key: somekey
      - name: busybox
        image: busybox
        envFrom:
        - configMapRef:
            name: someConfigMap
        - configMapRef:
            name: app-env
        volumeMounts:
        - mountPath: /tmp/env
          name: app-env
      volumes:
      - configMap:
          name: app-env
        name: app-env
`)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	// TODO(#669): The name of the patched Deployment is
	// test-infra-baseprefix-mungebot, retaining the base
	// prefix (example of correct behavior).
	th.assertActualEqualsExpected(m, `
apiVersion: v1
data:
  app-init.ini: |2

    FOO=bar
    BAR=baz
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mungebot
    org: kubernetes
    repo: test-infra
  name: test-infra-app-config-fd62mfc87h
---
apiVersion: v1
data:
  DB_PASSWORD: somepw
  DB_USERNAME: admin
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: mungebot
    org: kubernetes
    repo: test-infra
  name: test-infra-app-env-bh449c299k
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    baseAnno: This is a base annotation
    note: This is a test annotation
  labels:
    app: mungebot
    foo: bar
    org: kubernetes
    repo: test-infra
  name: test-infra-baseprefix-mungebot-service
spec:
  ports:
  - port: 7002
  selector:
    app: mungebot
    foo: bar
    org: kubernetes
    repo: test-infra
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  annotations:
    baseAnno: This is a base annotation
    note: This is a test annotation
  labels:
    app: mungebot
    foo: bar
    org: kubernetes
    repo: test-infra
  name: test-infra-baseprefix-mungebot
spec:
  replicas: 2
  selector:
    matchLabels:
      app: mungebot
      foo: bar
      org: kubernetes
      repo: test-infra
  template:
    metadata:
      annotations:
        baseAnno: This is a base annotation
        note: This is a test annotation
      labels:
        app: mungebot
        foo: bar
        org: kubernetes
        repo: test-infra
    spec:
      containers:
      - env:
        - name: FOO
          valueFrom:
            configMapKeyRef:
              key: somekey
              name: test-infra-app-env-bh449c299k
        - name: foo
          value: bar
        image: nginx:1.8.0
        name: nginx
        ports:
        - containerPort: 80
      - envFrom:
        - configMapRef:
            name: someConfigMap
        - configMapRef:
            name: test-infra-app-env-bh449c299k
        image: busybox
        name: busybox
        volumeMounts:
        - mountPath: /tmp/env
          name: app-env
      volumes:
      - configMap:
          name: test-infra-app-env-bh449c299k
        name: app-env
`)
}
