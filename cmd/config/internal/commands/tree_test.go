// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

func TestTreeCommandDefaultCurDir_files(t *testing.T) {
	d := t.TempDir()
	cwd, err := os.Getwd()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, os.Chdir(d)) {
		return
	}

	t.Cleanup(func() {
		if !assert.NoError(t, os.Chdir(cwd)) {
			t.FailNow()
		}
	})

	err = os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, `.
├── [f1.yaml]  Deployment foo
├── [f1.yaml]  Service foo
└── [f2.yaml]  Deployment bar
`, b.String()) {
		return
	}
}

// TestCmd_files verifies fmt reads the files and filters them
func TestTreeCommand_files(t *testing.T) {
	d := t.TempDir()

	err := os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{d})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, fmt.Sprintf(`%s
├── [f1.yaml]  Deployment foo
├── [f1.yaml]  Service foo
└── [f2.yaml]  Deployment bar
`, d), b.String()) {
		return
	}
}

func TestTreeCommand_Kustomization(t *testing.T) {
	d := t.TempDir()

	err := os.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	err = os.WriteFile(filepath.Join(d, "Kustomization"), []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- f2.yaml
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{d})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, fmt.Sprintf(`%s
└── [f2.yaml]  Deployment bar
`, d), b.String()) {
		return
	}
}

func TestTreeCommand_subpkgs(t *testing.T) {
	d := t.TempDir()

	err := os.MkdirAll(filepath.Join(d, "subpkg"), 0700)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "subpkg", "f2.yaml"), []byte(`kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	err = os.WriteFile(filepath.Join(d, "Krmfile"), []byte(`apiVersion: kpt.dev/v1alpha1
kind: Krmfile
metadata:
  name: mainpkg
openAPI:
  definitions:
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "subpkg", "Krmfile"), []byte(`apiVersion: kpt.dev/v1alpha1
kind: Krmfile
metadata:
  name: subpkg
openAPI:
  definitions:
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{d})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, fmt.Sprintf(`%s
├── [f1.yaml]  Deployment foo
├── [f1.yaml]  Service foo
└── Pkg: subpkg
    └── [f2.yaml]  Deployment bar
`, d), b.String()) {
		return
	}
}

func TestTreeCommand_stdin(t *testing.T) {
	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{"-"})
	r.Command.SetIn(bytes.NewBufferString(`apiVersion: extensions/v1
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo3
  namespace: default
  annotations:
    app: nginx2
    config.kubernetes.io/path: f1.yaml
spec:
  replicas: 1
---
apiVersion: extensions/v1
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo3
  namespace: default
  annotations:
    app: nginx2
    config.kubernetes.io/path: f1.yaml
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo3
  namespace: default
  annotations:
    app: nginx2
    config.kubernetes.io/path: f1.yaml
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo2
  namespace: default2
  annotations:
    app: nginx2
    config.kubernetes.io/path: f1.yaml
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx3
  name: foo
  namespace: default
  annotations:
    app: nginx3
    config.kubernetes.io/path: f1.yaml
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
    config.kubernetes.io/path: bar-package/f2.yaml
  name: bar
spec:
  replicas: 3
---
kind: Service
metadata:
  name: foo
  namespace: default
  annotations:
    app: nginx
    config.kubernetes.io/path: f1.yaml
spec:
  selector:
    app: nginx
`))
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, `.
├── [f1.yaml]  Deployment default/foo
├── [f1.yaml]  Service default/foo
├── [f1.yaml]  Deployment default/foo3
├── [f1.yaml]  Deployment default/foo3
├── [f1.yaml]  Deployment default/foo3
├── [f1.yaml]  Deployment default2/foo2
└── bar-package
    └── [f2.yaml]  Deployment bar
`, b.String()) {
		return
	}
}

func TestTreeCommand_includeReconcilers(t *testing.T) {
	d := t.TempDir()

	err := os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: gcr.io/example/reconciler:v1
kind: Abstraction
metadata:
  name: foo
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{d, "--include-local"})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, fmt.Sprintf(`%s
├── [f1.yaml]  Deployment foo
├── [f1.yaml]  Service foo
├── [f2.yaml]  Deployment bar
└── [f2.yaml]  Abstraction foo
`, d), b.String()) {
		return
	}
}

func TestTreeCommand_excludeNonReconcilers(t *testing.T) {
	d := t.TempDir()

	err := os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{d, "--include-local", "--exclude-non-local"})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, fmt.Sprintf(`%s
└── [f2.yaml]  Abstraction foo
`, d), b.String()) {
		return
	}
}

// TestTreeCommand_images tests that the image flag returns the image specified
// by various workloads.
func TestTreeCommand_images(t *testing.T) {
	d := t.TempDir()
	cwd, err := os.Getwd()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, os.Chdir(d)) {
		return
	}

	t.Cleanup(func() {
		if !assert.NoError(t, os.Chdir(cwd)) {
			t.FailNow()
		}
	})

	err = os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
        - name: test-deployment
          image: docker.io/bash:alpine3.18
---
kind: StatefulSet
metadata:
  name: statefulset
spec:
  template:
    spec:
      containers:
        - name: test-statefulset
          image: gcr.io/distroless/static-debian12:nonroot
---
kind: Job
metadata:
  name: job
spec:
  template:
    spec:
      containers:
        - name: job
          image: tagless
---
kind: CronJob
metadata:
  name: cronjob
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: test-cronjob
              image: local-image:v1.0.0
---
kind: Service
metadata:
  name: service
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetTreeRunner("")
	r.Command.SetArgs([]string{"--image"})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, `.
├── [f1.yaml]  CronJob cronjob
│   └── spec.jobTemplate.spec.template.spec.containers
│       └── 0
│           └── image: local-image:v1.0.0
├── [f1.yaml]  Deployment deployment
│   └── spec.template.spec.containers
│       └── 0
│           └── image: docker.io/bash:alpine3.18
├── [f1.yaml]  Job job
│   └── spec.template.spec.containers
│       └── 0
│           └── image: tagless
├── [f1.yaml]  Service service
└── [f1.yaml]  StatefulSet statefulset
    └── spec.template.spec.containers
        └── 0
            └── image: gcr.io/distroless/static-debian12:nonroot
`, b.String()) {
		return
	}
}
