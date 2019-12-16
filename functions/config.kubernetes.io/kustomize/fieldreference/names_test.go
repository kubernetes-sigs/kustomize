// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldreference_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/fieldreference"
)

func TestKustomizeNameFilter_Filter(t *testing.T) {
	doTestCases(t, nameTestCases)
}

var nameTestCases = []fieldReferenceTestCase{
	// Test Case
	{
		name: "add-name-prefix",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: prefix-instance-1
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: prefix-instance-2
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NamePrefix:        "prefix-",
		},
	},

	// Test Case
	{
		name: "update-name-prefix",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: prefix-instance-1
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: prefix-instance-2
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: new-instance-1
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: new-instance-2
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NamePrefix:        "new-",
		},
	},

	// Test Case
	{
		name: "add-name-suffix",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1-suffix
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2-suffix
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NameSuffix:        "-suffix",
		},
	},

	// Test Case
	{
		name: "update-name-suffix",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1-suffix
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2-suffix
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1-new
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2-new
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NameSuffix:        "-new",
		},
	},

	// Test Case
	{
		name: "name-prefix-to-suffix",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: prefix-instance-1
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: prefix-instance-2
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1-suffix
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2-suffix
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NameSuffix:        "-suffix",
		},
	},

	// Test Case
	{
		name: "name-suffix-to-prefix",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance-1-suffix
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-2-suffix
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: new-instance-1
  annotations:
    kustomize.io/original-name/test: instance-1
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: new-instance-2
  annotations:
    kustomize.io/original-name/test: instance-2
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NamePrefix:        "new-",
		},
	},

	// Test Case
	// Verify the Deployment Reference to a ConfigMap is updated
	{
		name: "pod-configmap-reference",
		input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: prefix-cm-instance-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-kust1234567890
    kustomize.io/reference-name/test: cm-instance
---
apiVersion: v1
kind: Pod
metadata:
  name: prefix-pod-instance
  annotations:
    kustomize.io/original-name/test: pod-instance
spec:
  volumes:
    configMap:
      name: prefix-cm-instance-kustabcdef
`,
		expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: new-cm-instance-suffix-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-kust1234567890
    kustomize.io/reference-name/test: cm-instance
---
apiVersion: v1
kind: Pod
metadata:
  name: new-pod-instance-suffix
  annotations:
    kustomize.io/original-name/test: pod-instance
spec:
  volumes:
    configMap:
      name: new-cm-instance-suffix-kust1234567890
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NamePrefix:        "new-",
			NameSuffix:        "-suffix",
		},
	},

	// Test Case
	// Verify the Deployment Reference to a ConfigMap is updated
	{
		name: "deployment-configmap-reference",
		input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: prefix-cm-instance-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-kust1234567890
    kustomize.io/reference-name/test: cm-instance
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prefix-deploy-instance
  annotations:
    kustomize.io/original-name/test: deploy-instance
spec:
  template:
    spec:
      volumes:
        configMap:
          name: prefix-cm-instance-kustabcdef
`,
		expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: new-cm-instance-suffix-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-kust1234567890
    kustomize.io/reference-name/test: cm-instance
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-deploy-instance-suffix
  annotations:
    kustomize.io/original-name/test: deploy-instance
spec:
  template:
    spec:
      volumes:
        configMap:
          name: new-cm-instance-suffix-kust1234567890
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NamePrefix:        "new-",
			NameSuffix:        "-suffix",
		},
	},

	// Test Case
	// Verify the Deployment Reference to a ConfigMap is updated
	{
		name: "deployment-containers-configmap-reference",
		input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: prefix-cm-instance-1-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-1-kust1234567890
    kustomize.io/reference-name/test: cm-instance-1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: prefix-cm-instance-2-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-2-kust1234567890
    kustomize.io/reference-name/test: cm-instance-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prefix-deploy-instance
  annotations:
    kustomize.io/original-name/test: deploy-instance
spec:
  template:
    spec:
      containers:
      - name: one
        env:
          valueFrom:
            configMapKeyRef:
              name: prefix-cm-instance-1-kustabcdef
      - name: two
        env:
          valueFrom:
            configMapKeyRef:
              name: prefix-cm-instance-2-kusthijglm
`,
		expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: new-cm-instance-1-suffix-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-1-kust1234567890
    kustomize.io/reference-name/test: cm-instance-1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: new-cm-instance-2-suffix-kust1234567890
  annotations:
    kustomize.io/original-name/test: cm-instance-2-kust1234567890
    kustomize.io/reference-name/test: cm-instance-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: new-deploy-instance-suffix
  annotations:
    kustomize.io/original-name/test: deploy-instance
spec:
  template:
    spec:
      containers:
      - name: one
        env:
          valueFrom:
            configMapKeyRef:
              name: new-cm-instance-1-suffix-kust1234567890
      - name: two
        env:
          valueFrom:
            configMapKeyRef:
              name: new-cm-instance-2-suffix-kust1234567890
`,
		instance: &KustomizeNameFilter{
			NameKustomizeName: "test",
			NamePrefix:        "new-",
			NameSuffix:        "-suffix",
		},
	},
}
