// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestPrinter_Write_Package_Structure(t *testing.T) {
	in := `kind: Deployment
metadata:
  labels:
    app: nginx3
  name: foo
  namespace: default
  annotations:
    app: nginx3
    config.kubernetes.io/package: foo-package/3
    config.kubernetes.io/path: f3.yaml
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  namespace: default
  annotations:
    app: nginx2
    config.kubernetes.io/package: foo-package
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
    config.kubernetes.io/package: bar-package
    config.kubernetes.io/path: f2.yaml
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
    config.kubernetes.io/package: foo-package
    config.kubernetes.io/path: f1.yaml
spec:
  selector:
    app: nginx
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out, Structure: TreeStructurePackage}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `
├── bar-package
│   └── [f2.yaml]  Deployment bar
└── foo-package
    ├── [f1.yaml]  Deployment default/foo
    ├── [f1.yaml]  Service default/foo
    └── foo-package/3
        └── [f3.yaml]  Deployment default/foo
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_Write_Package_Structure_base(t *testing.T) {
	in := `kind: Deployment
metadata:
  labels:
    app: nginx3
  name: foo
  namespace: default
  annotations:
    app: nginx3
    config.kubernetes.io/package: .
    config.kubernetes.io/path: f3.yaml
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  namespace: default
  annotations:
    app: nginx2
    config.kubernetes.io/package: foo-package
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
    config.kubernetes.io/package: bar-package
    config.kubernetes.io/path: f2.yaml
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
    config.kubernetes.io/package: .
    config.kubernetes.io/path: f1.yaml
spec:
  selector:
    app: nginx
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out, Structure: TreeStructurePackage}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `
├── [f1.yaml]  Service default/foo
├── [f3.yaml]  Deployment default/foo
├── bar-package
│   └── [f2.yaml]  Deployment bar
└── foo-package
    └── [f1.yaml]  Deployment default/foo
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_Write_Package_Structure_sort(t *testing.T) {
	in := `apiVersion: extensions/v1
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo3
  namespace: default
  annotations:
    app: nginx2
    config.kubernetes.io/package: .
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
    config.kubernetes.io/package: .
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
    config.kubernetes.io/package: .
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
    config.kubernetes.io/package: .
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
    config.kubernetes.io/package: .
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
    config.kubernetes.io/package: bar-package
    config.kubernetes.io/path: f2.yaml
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
    config.kubernetes.io/package: .
    config.kubernetes.io/path: f1.yaml
spec:
  selector:
    app: nginx
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out, Structure: TreeStructurePackage}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `
├── [f1.yaml]  Deployment default/foo
├── [f1.yaml]  Service default/foo
├── [f1.yaml]  Deployment default/foo3
├── [f1.yaml]  Deployment default/foo3
├── [f1.yaml]  Deployment default/foo3
├── [f1.yaml]  Deployment default2/foo2
└── bar-package
    └── [f2.yaml]  Deployment bar
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_metaError(t *testing.T) {
	out := &bytes.Buffer{}
	err := TreeWriter{Writer: out}.Write([]*yaml.RNode{{}})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, `
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_Write_Graph_Structure(t *testing.T) {
	in := `
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-0
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: apps/v1
    kind: StatefulSet
    name: cockroachdb
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-1
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: apps/v1
    kind: StatefulSet
    name: cockroachdb
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-2
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: apps/v1
    kind: StatefulSet
    name: cockroachdb
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.0
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cockroachdb
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: app.k8s.io/v1beta1
    kind: Application
    name: myapp
spec:
  replicas: 3
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Service
metadata:
  name: cockroachdb
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: app.k8s.io/v1beta1
    kind: Application
    name: myapp
---
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  labels:
    app.kubernetes.io/name: myapp
  name: myapp
  namespace: myapp-staging
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out, Structure: TreeStructureGraph}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `.
└── [Resource]  Application myapp-staging/myapp
    ├── [Resource]  Service myapp-staging/cockroachdb
    └── [Resource]  StatefulSet myapp-staging/cockroachdb
        ├── [Resource]  Pod myapp-staging/cockroachdb-0
        ├── [Resource]  Pod myapp-staging/cockroachdb-1
        └── [Resource]  Pod myapp-staging/cockroachdb-2
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_Write_Structure_Defaulting_when_ownerRefs_present(t *testing.T) {
	in := `
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-0
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: apps/v1
    kind: StatefulSet
    name: cockroachdb
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-1
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: apps/v1
    kind: StatefulSet
    name: cockroachdb
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-2
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: apps/v1
    kind: StatefulSet
    name: cockroachdb
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.0
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cockroachdb
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: app.k8s.io/v1beta1
    kind: Application
    name: myapp
spec:
  replicas: 3
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Service
metadata:
  name: cockroachdb
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: app.k8s.io/v1beta1
    kind: Application
    name: myapp
---
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  labels:
    app.kubernetes.io/name: myapp
  name: myapp
  namespace: myapp-staging
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out}}, // Structure unspecified
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `.
└── [Resource]  Application myapp-staging/myapp
    ├── [Resource]  Service myapp-staging/cockroachdb
    └── [Resource]  StatefulSet myapp-staging/cockroachdb
        ├── [Resource]  Pod myapp-staging/cockroachdb-0
        ├── [Resource]  Pod myapp-staging/cockroachdb-1
        └── [Resource]  Pod myapp-staging/cockroachdb-2
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_Write_Structure_Defaulting_when_ownerRefs_absent(t *testing.T) {
	in := `
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-0
  namespace: myapp-staging
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-1
  namespace: myapp-staging
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Pod
metadata:
  name: cockroachdb-2
  namespace: myapp-staging
spec:
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.0
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cockroachdb
  namespace: myapp-staging
spec:
  replicas: 3
  containers:
  - name: cockroachdb
    image: cockraochdb:1.1.1
---
apiVersion: v1
kind: Service
metadata:
  name: cockroachdb
  namespace: myapp-staging
---
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  labels:
    app.kubernetes.io/name: myapp
  name: myapp
  namespace: myapp-staging
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out}}, // Structure unspecified
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `
└── 
    ├── [.]  Service myapp-staging/cockroachdb
    ├── [.]  StatefulSet myapp-staging/cockroachdb
    ├── [.]  Pod myapp-staging/cockroachdb-0
    ├── [.]  Pod myapp-staging/cockroachdb-1
    ├── [.]  Pod myapp-staging/cockroachdb-2
    └── [.]  Application myapp-staging/myapp
`, out.String()) {
		t.FailNow()
	}
}

func TestPrinter_Write_error_when_owner_missing(t *testing.T) {
	in := `
---
apiVersion: v1
kind: Service
metadata:
  name: cockroachdb
  namespace: myapp-staging
  ownerReferences:
  - apiVersion: app.k8s.io/v1beta1
    kind: Application
    name: nginx
---
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  labels:
    app.kubernetes.io/name: myapp
  name: myapp
  namespace: myapp-staging
`
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: bytes.NewBufferString(in)}},
		Outputs: []Writer{TreeWriter{Writer: out}},
	}.Execute()
	assert.Error(t, err)
	assert.Equal(t, "owner 'Application myapp-staging/nginx' not found in input, but found as an owner of input objects", err.Error())
}
