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
		expectedErr  string
	}{
		"simple": {
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
- source:
    kind: Deployment
    name: deploy
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
      name: deploy
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:1.7.9
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
			expectedErr: "more than one match for source ~G_~V_Deployment",
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
			expectedErr: "replacements must specify a source and at least one target",
		},
		"field paths with key-value pairs": {
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
    fieldPath: spec.template.spec.containers.[name=nginx-tagged].image
  targets:
  - select:
      kind: Deployment
      name: deploy1
    fieldPaths: 
    - spec.template.spec.containers.[name=postgresdb].image
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
		"select by group and version": {
			input: `apiVersion: my-group-1/v1
kind: Custom
metadata:
  name: my-name
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group-2/v2
kind: Custom
metadata:
  name: my-name
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group-3/v3
kind: Custom
metadata:
  name: my-name
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
    group: my-group-2
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      version: v3
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: my-group-1/v1
kind: Custom
metadata:
  name: my-name
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group-2/v2
kind: Custom
metadata:
  name: my-name
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group-3/v3
kind: Custom
metadata:
  name: my-name
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:1.7.9
        name: postgresdb
`,
		},
		// regression test for missing metadata handling
		"missing metadata": {
			input: `spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group/v1
kind: Custom
metadata:
  name: my-name-1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group/v1
kind: Custom
metadata:
  name: my-name-2
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
    name: my-name-1
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      name: my-name-2
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group/v1
kind: Custom
metadata:
  name: my-name-1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: postgres:1.8.0
        name: postgresdb
---
apiVersion: my-group/v1
kind: Custom
metadata:
  name: my-name-2
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:1.7.9
        name: postgresdb
`,
		},
		"reject 1": {
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
    name: deploy2
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
    reject:
    - name: deploy2
    - name: deploy3
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
		},
		"reject 2": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: my-name
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
kind: StatefulSet
metadata:
  name: my-name
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
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      version: v1
    reject:
    - kind: Deployment
      name: my-name
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: my-name
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
kind: StatefulSet
metadata:
  name: my-name
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:1.7.9
        name: postgresdb
`,
		},
		// the only difference in the inputs between this and the previous test
		// is the dash before `name: my-name` on line 733
		"reject 3": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: my-name
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
kind: StatefulSet
metadata:
  name: my-name
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
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      version: v1
    reject:
    - kind: Deployment
    - name: my-name
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: my-name
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
kind: StatefulSet
metadata:
  name: my-name
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
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			f := Filter{}
			err := yaml.Unmarshal([]byte(tc.replacements), &f)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			actual, err := filtertest.RunFilterE(t, tc.input, f)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("unexpected error: %s\n", err.Error())
					t.FailNow()
				}
				if !assert.Equal(t, tc.expectedErr, err.Error()) {
					t.FailNow()
				}
			}
			if !assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}
