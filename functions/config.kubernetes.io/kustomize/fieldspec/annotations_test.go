// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"testing"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/testutil"
)

func TestAnnotationFilter(t *testing.T) {
	testutil.RunTestCases(t, annotationTestCases)
}

var bar = "bar"

var annotationTestCases = []testutil.FieldSpecTestCase{

	// Test Case
	{
		Name: "add-annotations-crd",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  annotations:
    # keep this annotation
    a: b
spec:
  template:
    metadata:
      annotations:
        c: d
---
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-2
  annotations: {}
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  annotations:
    foo: bar
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  annotations:
    # keep this annotation
    a: b
    foo: bar
spec:
  template:
    metadata:
      annotations:
        c: d
        foo: bar
---
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance-2
  annotations:
    foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "update-annotation-crd",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  annotations:
    foo: baz
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  annotations:
    foo: baz
    a: b
spec:
  template:
    metadata:
      annotations:
        c: d
        foo: baz
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  annotations:
    foo: bar
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  annotations:
    foo: bar
    a: b
spec:
  template:
    metadata:
      annotations:
        c: d
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-replication-controller",
		Input: `
apiVersion: v1
kind: ReplicationController
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: v1
kind: ReplicationController
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  template:
    metadata:
      annotations:
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-deployment",
		Input: `
apiVersion: example.com/v1alpha17
kind: Deployment
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: example.com/v1alpha17
kind: Deployment
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  template:
    metadata:
      annotations:
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-replica-set",
		Input: `
apiVersion: example.com/v1alpha17
kind: ReplicaSet
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: example.com/v1alpha17
kind: ReplicaSet
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  template:
    metadata:
      annotations:
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-stateful-set",
		Input: `
apiVersion: example.com/v1alpha17
kind: StatefulSet
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: example.com/v1alpha17
kind: StatefulSet
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  template:
    metadata:
      annotations:
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-daemon-set",
		Input: `
apiVersion: example.com/v1alpha17
kind: DaemonSet
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: example.com/v1alpha17
kind: DaemonSet
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  template:
    metadata:
      annotations:
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-job",
		Input: `
apiVersion: batch/v1alpha17
kind: Job
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: batch/v1alpha17
kind: Job
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  template:
    metadata:
      annotations:
        foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "add-annotation-job",
		Input: `
apiVersion: batch/v1alpha17
kind: CronJob
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: batch/v1alpha17
kind: CronJob
metadata:
  Name: Instance
  annotations:
    foo: bar
spec:
  jobTemplate:
    metadata:
      annotations:
        foo: bar
    spec:
      template:
        metadata:
          annotations:
            foo: bar
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}},
	},

	// Test Case
	{
		Name: "owner",
		Input: `
apiVersion: batch/v1alpha17
kind: CronJob
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: batch/v1alpha17
kind: CronJob
metadata:
  Name: Instance
  annotations:
    foo: bar # {"ownedBy":"owner"}
spec:
  jobTemplate:
    metadata:
      annotations:
        foo: bar # {"ownedBy":"owner"}
    spec:
      template:
        metadata:
          annotations:
            foo: bar # {"ownedBy":"owner"}
`,
		Instance: KustomizeAnnotationsFilter{Annotations: map[string]*string{"foo": &bar}, kustomizeName: "owner"},
	},
}
