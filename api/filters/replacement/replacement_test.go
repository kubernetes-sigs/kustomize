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
			expectedErr: "multiple matches for selector Deployment.[noVer].[noGrp]/[noName].[noNs]",
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
		"partial string replacement - replace": {
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
    options:
      delimiter: ':'
  targets:
  - select:
      kind: Deployment
      name: deploy1
    fieldPaths: 
    - spec.template.spec.containers.1.image
    options:
      delimiter: ':'
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
      - image: nginx:1.8.0
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
		"partial string replacement - prefix": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group/config
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod1
    fieldPath: spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 0
  targets:
  - select:
      kind: Pod
      name: pod2
    fieldPaths: 
    - spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: -1
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group/config
`,
		},
		"partial string replacement - suffix": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group/config
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod2
    fieldPath: spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 1
  targets:
  - select:
      kind: Pod
      name: pod1
    fieldPaths: 
    - spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 2
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group/config
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group/config
`,
		},
		"partial string replacement - last element": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group1
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group2
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod2
    fieldPath: spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 0
  targets:
  - select:
      kind: Pod
      name: pod1
    fieldPaths: 
    - spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 1
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group2
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group2
`,
		},
		"partial string replacement - first element": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group1/config
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group2
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod2
    fieldPath: spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 0
  targets:
  - select:
      kind: Pod
      name: pod1
    fieldPaths: 
    - spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 0
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group2/config
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group2
`,
		},
		"options.index out of bounds": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: my/group1
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  volumes:
  - projected:
      sources:
      - configMap:
          name: myconfigmap
          items:
          - key: config
            path: group2
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod2
    fieldPath: spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: -1
  targets:
  - select:
      kind: Pod
      name: pod1
    fieldPaths: 
    - spec.volumes.0.projected.sources.0.configMap.items.0.path
    options:
      delimiter: '/'
      index: 1
`,
			expectedErr: "options.index -1 is out of bounds for value group2",
		},
		"create": {
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
  name: deploy1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers
  targets:
  - select:
      name: deploy1
    fieldPaths: 
    - spec.template.spec.containers
    options:
      create: true
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers
  targets:
  - select:
      name: deploy2
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
  name: deploy1
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
  name: deploy2
`,
		},
		"complex type with delimiter in source": {
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
    options: 
      delimiter: "/"
  targets:
  - select:
      kind: Deployment
    fieldPaths: 
    - spec.template.spec.containers
`,
			expectedErr: "delimiter option can only be used with scalar nodes",
		},
		"complex type with delimiter in target": {
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
    options: 
      delimiter: "/"
`,
			expectedErr: "delimiter option can only be used with scalar nodes",
		},
		"mapping value contains '.' character": {
			input: `apiVersion: v1
kind: Custom
metadata:
  name: custom
  annotations:
    a.b.c/d-e: source
    f.g.h/i-j: target
`,
			replacements: `replacements:
- source:
    name: custom
    fieldPath: metadata.annotations.[a.b.c/d-e]
  targets:
  - select:
      name: custom
    fieldPaths: 
    - metadata.annotations.[f.g.h/i-j]
`,
			expected: `apiVersion: v1
kind: Custom
metadata:
  name: custom
  annotations:
    a.b.c/d-e: source
    f.g.h/i-j: source
`,
		},
		"list index contains '.' character": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: source
data:
  value: example
---
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: some-secret
spec:
  backendType: secretsManager
  data:
    - key: some-prefix-replaceme
      name: .first
      version: latest
      property: first
    - key: some-prefix-replaceme
      name: second
      version: latest
      property: second
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    version: v1
    name: source
    fieldPath: data.value
  targets:
  - select:
      group: kubernetes-client.io
      version: v1
      kind: ExternalSecret
      name: some-secret
    fieldPaths:
    - spec.data.[name=.first].key
    - spec.data.[name=second].key
    options:
      delimiter: "-"
      index: 2
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: source
data:
  value: example
---
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: some-secret
spec:
  backendType: secretsManager
  data:
  - key: some-prefix-example
    name: .first
    version: latest
    property: first
  - key: some-prefix-example
    name: second
    version: latest
    property: second`,
		},
		"multiple field paths in target": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: source
data:
  value: example
---
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: some-secret
spec:
  backendType: secretsManager
  data:
    - key: some-prefix-replaceme
      name: first
      version: latest
      property: first
    - key: some-prefix-replaceme
      name: second
      version: latest
      property: second
    - key: some-prefix-replaceme
      name: third
      version: latest
      property: third
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    version: v1
    name: source
    fieldPath: data.value
  targets:
  - select:
      group: kubernetes-client.io
      version: v1
      kind: ExternalSecret
      name: some-secret
    fieldPaths:
    - spec.data.0.key
    - spec.data.1.key
    - spec.data.2.key
    options:
      delimiter: "-"
      index: 2
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: source
data:
  value: example
---
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: some-secret
spec:
  backendType: secretsManager
  data:
  - key: some-prefix-example
    name: first
    version: latest
    property: first
  - key: some-prefix-example
    name: second
    version: latest
    property: second
  - key: some-prefix-example
    name: third
    version: latest
    property: third
`,
		},
		"using a previous ID": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: pre-deploy
  annotations: 
    internal.config.kubernetes.io/previousNames: deploy,deploy
    internal.config.kubernetes.io/previousKinds: CronJob,Deployment
    internal.config.kubernetes.io/previousNamespaces: default,default
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
    kind: CronJob
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
  name: pre-deploy
  annotations:
    internal.config.kubernetes.io/previousNames: deploy,deploy
    internal.config.kubernetes.io/previousKinds: CronJob,Deployment
    internal.config.kubernetes.io/previousNamespaces: default,default
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

		"replacement source.fieldPath does not exist": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
data:
  grpcPort: 8081
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: ports-from
    fieldPath: data.httpPort
  targets:
  - select:
      kind: ConfigMap
      name: ports-to
    fieldPaths:
    - data.grpcPort
    options:
      create: true
`,
			expectedErr: "fieldPath `data.httpPort` is missing for replacement source ConfigMap.[noVer].[noGrp]/ports-from.[noNs]",
		},
		"annotationSelector": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy-1
  annotations:
    foo: bar-1
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
  name: deploy-2
  annotations:
    foo: bar-2
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
    name: deploy-1
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      annotationSelector: foo=bar-1
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy-1
  annotations:
    foo: bar-1
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
  name: deploy-2
  annotations:
    foo: bar-2
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
		"labelSelector": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy-1
  labels:
    foo: bar-1
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
  name: deploy-2
  labels:
    foo: bar-2
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
    name: deploy-1
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      labelSelector: foo=bar-1
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy-1
  labels:
    foo: bar-1
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
  name: deploy-2
  labels:
    foo: bar-2
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
		"reject via labelSelector": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy-1
  labels:
    foo: bar-1
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
  name: deploy-2
  labels:
    foo: bar-2
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
    name: deploy-1
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
    reject:
    - labelSelector: foo=bar-2
    fieldPaths: 
    - spec.template.spec.containers.1.image
`,
			expected: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy-1
  labels:
    foo: bar-1
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
  name: deploy-2
  labels:
    foo: bar-2
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
				if !assert.Contains(t, err.Error(), tc.expectedErr) {
					t.FailNow()
				}
			}
			if !assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}
