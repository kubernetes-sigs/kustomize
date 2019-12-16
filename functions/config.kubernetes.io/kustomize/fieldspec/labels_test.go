// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"testing"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/testutil"
)

func TestLabelsFilter(t *testing.T) {
	testutil.RunTestCases(t, labelsTestCases)
}

var f = "f"

var labelsTestCases = []testutil.FieldSpecTestCase{

	// Test Case
	{
		Name: "crd",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-1
spec:
  template:
    metadata:
      labels: {}
---
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-2
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-1
  labels:
    e: f
spec:
  template:
    metadata:
      labels:
        e: f
---
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-2
  labels:
    e: f
`,
		Instance: KustomizeLabelsFilter{Labels: map[string]*string{"e": &f}},
	},

	// Test Case
	{
		Name: "builtin-spec-selector-matchLabels",
		Input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  Name: Instance-1
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  Name: Instance-2
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  Name: Instance-3
spec:
  volumeClaimTemplates:
  - metadata:
      Name: foo-1
  - metadata:
      Name: foo-2
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  Name: Instance-4
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  Name: Instance-5
---
apiVersion: apps/v1
kind: ReplicationController
metadata:
  Name: Instance-6
`,
		Expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  Name: Instance-1
  labels:
    e: f
spec:
  template:
    metadata:
      labels:
        e: f
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  Name: Instance-2
  labels:
    e: f
spec:
  template:
    metadata:
      labels:
        e: f
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  Name: Instance-3
  labels:
    e: f
spec:
  volumeClaimTemplates:
  - metadata:
      Name: foo-1
      labels:
        e: f
  - metadata:
      Name: foo-2
      labels:
        e: f
  template:
    metadata:
      labels:
        e: f
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  Name: Instance-4
  labels:
    e: f
spec:
  template:
    metadata:
      labels:
        e: f
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  Name: Instance-5
  labels:
    e: f
spec:
  template:
    metadata:
      labels:
        e: f
---
apiVersion: apps/v1
kind: ReplicationController
metadata:
  Name: Instance-6
  labels:
    e: f
spec:
  template:
    metadata:
      labels:
        e: f
`,
		Instance: KustomizeLabelsFilter{Labels: map[string]*string{"e": &f}},
	},

	{
		Name: "owner",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-1
spec:
  template:
    metadata:
      labels: {}
---
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-2
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-1
  labels:
    e: f # {"ownedBy":"owner"}
spec:
  template:
    metadata:
      labels:
        e: f # {"ownedBy":"owner"}
---
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-2
  labels:
    e: f # {"ownedBy":"owner"}
`,
		Instance: KustomizeLabelsFilter{Labels: map[string]*string{"e": &f}, kustomizeName: "owner"},
	},
}
