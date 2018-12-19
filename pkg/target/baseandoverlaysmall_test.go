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

package target

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
)

func writeSmallBase(t *testing.T, ldr loadertest.FakeLoader) {
	writeK(t, ldr, "/app/base", `
namePrefix: a-
commonLabels:
  app: myApp
resources:
- deployment.yaml
- service.yaml
`)
	writeF(t, ldr, "/app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    backend: bungie
  ports:
    - port: 7002
`)
	writeF(t, ldr, "/app/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
}

func TestSmallBase(t *testing.T) {
	ldr := loadertest.NewFakeLoader("/app/base")
	writeSmallBase(t, ldr)
	m, err := makeKustTarget(t, ldr).MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	if m == nil {
		t.Fatalf("Empty map.")
	}
	s, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	assertExpectedEqualsActual(t, s, `apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
  name: a-myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    backend: bungie
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
  name: a-myDeployment
spec:
  selector:
    matchLabels:
      app: myApp
  template:
    metadata:
      labels:
        app: myApp
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
`)
}

func TestSmallOverlay(t *testing.T) {
	ldr := loadertest.NewFakeLoader("/app/overlay")
	writeSmallBase(t, ldr)
	writeK(t, ldr, "/app/overlay", `
namePrefix: b-
commonLabels:
  env: prod
bases:
- ../base
patchesStrategicMerge:
- deployment/deployment.yaml
imageTags:
- name: whatever
  newTag: 1.8.0`)

	writeF(t, ldr, "/app/overlay/configmap/app.env", `
DB_USERNAME=admin
DB_PASSWORD=somepw
`)
	writeF(t, ldr, "/app/overlay/configmap/app-init.ini", `
FOO=bar
BAR=baz
`)
	writeF(t, ldr, "/app/overlay/deployment/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 1000
`)
	m, err := makeKustTarget(t, ldr).MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	s, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	assertExpectedEqualsActual(t, s, `apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
    env: prod
  name: b-a-myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    backend: bungie
    env: prod
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
    env: prod
  name: b-a-myDeployment
spec:
  replicas: 1000
  selector:
    matchLabels:
      app: myApp
      env: prod
  template:
    metadata:
      labels:
        app: myApp
        backend: awesome
        env: prod
    spec:
      containers:
      - image: whatever:1.8.0
        name: whatever
`)
}
