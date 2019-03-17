/*
Copyright 2019 The Kubernetes Authors.

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

func makeTransfomersImageBase(th *KustTestHarness) {
	th.writeK("/app/base", `
resources:
- deploy1.yaml
- random.yaml
images:
- name: nginx
  newTag: v2
- name: my-nginx
  newTag: previous
- name: myprivaterepohostname:1234/my/image
  newTag: v1.0.1
- name: foobar
  digest: sha256:24a0c4b4
- name: alpine
  newName: myprivaterepohostname:1234/my/cool-alpine
- name: gcr.io:8080/my-project/my-cool-app
  newName: my-cool-app
- name: postgres
  newName: my-postgres
  newTag: v3
- name: docker
  newName: my-docker
  digest: sha256:25a0d4b4
`)
	th.writeF("/app/base/deploy1.yaml", `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      initContainers:
      - name: nginx2
        image: my-nginx:1.8.0
      - name: init-alpine
        image: alpine:1.8.0
      containers:
      - name: ngnix
        image: nginx:1.7.9
      - name: repliaced-with-digest
        image: foobar:1
      - name: postgresdb
        image: postgres:1.8.0
`)
	th.writeF("/app/base/random.yaml", `
kind: randomKind
metadata:
  name: random
spec:
  template:
    spec:
      containers:
      - name: ngnix1
        image: nginx
spec2:
  template:
    spec:
      containers:
      - name: nginx3
        image: nginx:v1
      - name: nginx4
        image: my-nginx:latest
spec3:
  template:
    spec:
      initContainers:
      - name: postgresdb
        image: postgres:alpine-9
      - name: init-docker
        image: docker:17-git
      - name: myImage
        image: myprivaterepohostname:1234/my/image:latest
      - name: myImage2
        image: myprivaterepohostname:1234/my/image
      - name: my-app
        image: my-app-image:v1
      - name: my-cool-app
        image: gcr.io:8080/my-project/my-cool-app:latest
`)
}

func TestTransfomersImageDefaultConfig(t *testing.T) {
	th := NewKustTestHarness(t, "/app/base")
	makeTransfomersImageBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:v2
        name: ngnix
      - image: foobar@sha256:24a0c4b4
        name: repliaced-with-digest
      - image: my-postgres:v3
        name: postgresdb
      initContainers:
      - image: my-nginx:previous
        name: nginx2
      - image: myprivaterepohostname:1234/my/cool-alpine:1.8.0
        name: init-alpine
---
kind: randomKind
metadata:
  name: random
spec:
  template:
    spec:
      containers:
      - image: nginx:v2
        name: ngnix1
spec2:
  template:
    spec:
      containers:
      - image: nginx:v2
        name: nginx3
      - image: my-nginx:previous
        name: nginx4
spec3:
  template:
    spec:
      initContainers:
      - image: my-postgres:v3
        name: postgresdb
      - image: my-docker@sha256:25a0d4b4
        name: init-docker
      - image: myprivaterepohostname:1234/my/image:v1.0.1
        name: myImage
      - image: myprivaterepohostname:1234/my/image:v1.0.1
        name: myImage2
      - image: my-app-image:v1
        name: my-app
      - image: my-cool-app:latest
        name: my-cool-app
`)
}
