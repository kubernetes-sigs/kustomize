// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replacement

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/yaml"
)

func TestFilter(t *testing.T) {
	testCases := map[string]struct {
		input        string
		replacements string
		expected     string
		expectedErr  bool
	}{
		"simple": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
`,
			replacements: `replacements:
- source:
    kind: Deployment
    name: deploy2
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
      name: deploy1
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:1.7.9
        name: postgresdb
---
apiVersion: v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
`,
		},
		"complex type": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers: {}
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers
  targets:
  - select:
      kind: Deployment
    fieldPaths: 
    - spec.template.spec.containers
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
`,
		},
		"from ConfigMap": {
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  labels:
    foo: bar
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: command-demo-container
          image: debian
          command: ["printenv"]
          args:
            - HOSTNAME
            - PORT
        - name: busybox
          image: busybox:latest
          args:
            - echo
            - HOSTNAME
            - PORT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOSTNAME: example.com
  PORT: 8080
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOSTNAME
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.0.args.0
    - spec.template.spec.containers.1.args.1
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.PORT
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.0.args.1
    - spec.template.spec.containers.1.args.2
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  labels:
    foo: bar
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
      - name: command-demo-container
        image: debian
        command: ["printenv"]
        args:
        - example.com
        - 8080
      - name: busybox
        image: busybox:latest
        args:
        - echo
        - example.com
        - 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOSTNAME: example.com
  PORT: 8080
`,
		},
		"multiple matches for source select": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
`,
			replacements: `replacements:
- source:
    kind: Deployment
  targets:
  - select:
      kind: Deployment
`,
			expectedErr: true,
		},
		"replacement has no source": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
`,
			replacements: `replacements:
- targets:
  - select:
      kind: Deployment
`,
			expectedErr: true,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			f := Filter{}
			err := yaml.Unmarshal([]byte(tc.replacements), &f)
			assert.NoError(t, err)
			actual, err := filtertest.RunFilterE(t, tc.input, f)
			assert.Equal(t, tc.expectedErr, err != nil)
			if !tc.expectedErr &&
				!assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}
