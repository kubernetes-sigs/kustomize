// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

func TestReplicaCountTransformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
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
apiVersion: apps/v1
kind: Service
metadata:
  name: myapp
spec:
  ports:
  - port: 1111
  targetport: 1111
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otherapp
spec:
  replicas: 5
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 5
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: app
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: app
---
apiVersion: apps/v1
kind: ReplicationController
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: app
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Service
metadata:
  name: myapp
spec:
  ports:
  - port: 1111
  targetport: 1111
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otherapp
spec:
  replicas: 5
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 23
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: myapp
spec:
  replicas: 23
  selector:
    matchLabels:
      app: app
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: myapp
spec:
  replicas: 23
  selector:
    matchLabels:
      app: app
---
apiVersion: apps/v1
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

func TestMatchesCurrentID(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin("builtin", "", "PrefixSuffixTransformer")
	tc.BuildGoPlugin("builtin", "", "ReplicaCountTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: notImportantHere
suffix: -test
fieldSpecs:
  - path: metadata/name
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment`)

	rm, _ = th.RunTransformerFromResMap(`
apiVersion: builtin
kind: ReplicaCountTransformer
metadata:
  name: notImportantHere

replica:
  name: deployment-test
  count: 23
fieldSpecs:
- path: spec/replicas
  create: true
  kind: Deployment`, rm)
	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-test
spec:
  replicas: 23
`)
}

func TestNoMatch(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin("builtin", "", "ReplicaCountTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	err := th.ErrorFromLoadAndRunTransformer(`
apiVersion: builtin
kind: ReplicaCountTransformer
metadata:
  name: notImportantHere
replica:
  name: service
  count: 3
fieldSpecs:
- path: spec/replicas
  create: true
  kind: Deployment
`, `
kind: Service
metadata:
  name: service
spec:
`)

	if err == nil {
		t.Fatalf("No match should return an error")
	}
	if err.Error() !=
		"Resource with name service does not match a config with the following GVK [~G_~V_Deployment]" {
		t.Fatalf("Unexpected error: %v", err)
	}
}
