// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestImageTagTransformerNewTag(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  newTag: v2
fieldSpecs:
- path: spec/template/spec/containers[]/image
- path: spec/template/spec/initContainers[]/image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)

	th.AssertActualEqualsExpected(rm, `
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
        name: nginx-tagged
      - image: nginx:v2
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx:v2
        name: nginx-notag
      - image: nginx:v2
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)
}
func TestImageTagTransformerNewImage(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  newName: busybox
fieldSpecs:
- path: spec/template/spec/containers[]/image
- path: spec/template/spec/initContainers[]/image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: busybox:1.7.9
        name: nginx-tagged
      - image: busybox:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: busybox
        name: nginx-notag
      - image: busybox@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)
}

func TestImageTagTransformerNewImageAndTag(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  newName: busybox
  newTag: v2
fieldSpecs:
- path: spec/template/spec/containers[]/image
- path: spec/template/spec/initContainers[]/image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: busybox:v2
        name: nginx-tagged
      - image: busybox:v2
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: busybox:v2
        name: nginx-notag
      - image: busybox:v2
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)
}

func TestImageTagTransformerNewDigest(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  Digest: sha256:222222222222222222
fieldSpecs:
- path: spec/template/spec/containers[]/image
- path: spec/template/spec/initContainers[]/image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx@sha256:222222222222222222
        name: nginx-tagged
      - image: nginx@sha256:222222222222222222
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx@sha256:222222222222222222
        name: nginx-notag
      - image: nginx@sha256:222222222222222222
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)
}

func TestImageTagTransformerNewImageAndDigest(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  newName: busybox
  Digest: sha256:222222222222222222
fieldSpecs:
- path: spec/template/spec/containers[]/image
- path: spec/template/spec/initContainers[]/image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: busybox@sha256:222222222222222222
        name: nginx-tagged
      - image: busybox@sha256:222222222222222222
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: busybox@sha256:222222222222222222
        name: nginx-notag
      - image: busybox@sha256:222222222222222222
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)
}

func TestImageTagTransformerEmptyContainers(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: nginx
  newTag: v2
fieldSpecs:
- path: spec/template/spec/containers[]/image
  create: true
- path: spec/template/spec/initContainers[]/image
  create: true
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      initContainers:
`)
	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers: []
      initContainers: []
`)
}

func TestImageTagTransformerTagWithBraces(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: some.registry.io/my-image
  newTag: "my-fixed-tag"
fieldSpecs:
- path: spec/template/spec/containers[]/image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: some.registry.io/my-image:{GENERATED_TAG}
        name: my-image
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: some.registry.io/my-image:my-fixed-tag
        name: my-image
`)
}

func TestImageTagTransformerArbitraryPath(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageTagTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ImageTagTransformer
metadata:
  name: notImportantHere
imageTag:
  name: some.registry.io/my-image
  newTag: "my-fixed-tag"
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: some.registry.io/my-image:old-tag
        name: my-image
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: some.registry.io/my-image:my-fixed-tag
        name: my-image
`)
}

func TestImageTagTransformerInKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
resources:
- resources.yaml
images:
- name: old-image-name
  newName: new-image-name
  newTag: new-tag
`)

	th.WriteF("/app/resources.yaml", `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  containers:
  - image: old-image-name
    name: my-image
  initContainers:
  - image: old-image-name
    name: my-image
  template:
    spec:
      containers:
      - image: old-image-name
        name: my-image
      initContainers:
      - image: old-image-name
        name: my-image
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  containers:
  - image: new-image-name:new-tag
    name: my-image
  initContainers:
  - image: new-image-name:new-tag
    name: my-image
  template:
    spec:
      containers:
      - image: new-image-name:new-tag
        name: my-image
      initContainers:
      - image: new-image-name:new-tag
        name: my-image
`)
}
