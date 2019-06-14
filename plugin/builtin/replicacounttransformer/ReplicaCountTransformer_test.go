// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/plugin"
)

func TestReplicaCountTransformer(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "ReplicaCountTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ReplicaCountTransformer
metadata:
  name: notImportantHere

replica:
  name: myapp
  count: 23
fieldSpecs:
- path: spec/replicas
  create: true
  kind: Deployment

- path: spec/replicas
  create: true
  kind: ReplicationController

- path: spec/replicas
  create: true
  kind: ReplicaSet

- path: spec/replicas
  create: true
  kind: StatefulSet
`, `
apiVersion: builtin
kind: Service
metadata:
  name: myapp
spec:
  ports:
  - port: 1111
  targetport: 1111
---
apiVersion: builtin
kind: Deployment
metadata:
  name: otherapp
spec:
  replicas: 5
---
apiVersion: builtin
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 5
---
apiVersion: builtin
kind: StatefulSet
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: app
---
apiVersion: builtin
kind: ReplicaSet
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: app
---
apiVersion: builtin
kind: ReplicationController
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: app
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: builtin
kind: Service
metadata:
  name: myapp
spec:
  ports:
  - port: 1111
  targetport: 1111
---
apiVersion: builtin
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 23
---
apiVersion: builtin
kind: Deployment
metadata:
  name: otherapp
spec:
  replicas: 5
---
apiVersion: builtin
kind: StatefulSet
metadata:
  name: myapp
spec:
  replicas: 23
  selector:
    matchLabels:
      app: app
---
apiVersion: builtin
kind: ReplicaSet
metadata:
  name: myapp
spec:
  replicas: 23
  selector:
    matchLabels:
      app: app
---
apiVersion: builtin
kind: ReplicationController
metadata:
  name: myapp
spec:
  replicas: 23
  selector:
    matchLabels:
      app: app
`)
}
