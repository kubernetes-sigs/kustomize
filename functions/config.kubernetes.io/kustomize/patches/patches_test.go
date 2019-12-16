// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patches

import (
	"testing"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/testutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestPatchesFilter(t *testing.T) {
	testutil.RunTestCases(t, patchTestCases)
}

var patchTestCases = []testutil.FieldSpecTestCase{
	// Test Case
	{
		Name: "add-namespace",
		Input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  labels:
    a: b
spec:
  selector:
    a: b
  template:
    metadata:
      labels:
        a: b
    spec:
      containers:
      - name: nginx
        image: nginx:v1.0.0
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  labels:
    a: c
spec:
  selector:
    a: c
  template:
    metadata:
      labels:
        a: c
    spec:
      containers:
      - name: nginx
        image: nginx:v1.0.1
`,
		Expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  labels:
    a: b
spec:
  selector:
    a: b
  template:
    metadata:
      labels:
        a: b
    spec:
      containers:
      - name: nginx
        image: nginx:v1.0.0
        args:
        - d # {"ownedBy":"owner"}
        - e # {"ownedBy":"owner"}
        - f # {"ownedBy":"owner"}
  replicas: 5 # {"ownedBy":"owner"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  labels:
    a: c
spec:
  selector:
    a: c
  template:
    metadata:
      labels:
        a: c
    spec:
      containers:
      - name: nginx
        image: nginx:v1.0.1
`,
		Instance: &PatchFilter{
			Patches: []Patch{
				{
					Patch: *yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  replicas: 5
  template:
    spec:
      containers:
      - name: nginx
        args: ['d', 'e', 'f']
`).YNode(),
					Targets: Targets{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
						Name:       "foo",
					},
				},
			},
			setterName: "owner",
		},
	},
}
