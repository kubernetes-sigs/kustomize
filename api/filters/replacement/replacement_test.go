// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replacement

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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
        - "8080"
      - name: busybox
        image: busybox:latest
        args:
        - echo
        - example.com
        - "8080"
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
		"one replacements target has multiple value": {
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: deployment-name
          value: XXXXX
        - name: foo
          value: bar
      - image: nginx
        name: sidecar
        env:
        - name: deployment-name
          value: YYYYY
`,
			replacements: `replacements:
- source:
    kind: Deployment
    name: sample-deploy
    fieldPath: metadata.name
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.[image=nginx].env.[name=deployment-name].value
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: deployment-name
          value: sample-deploy
        - name: foo
          value: bar
      - image: nginx
        name: sidecar
        env:
        - name: deployment-name
          value: sample-deploy`,
		},
		"one replacements target should create multiple values": {
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: other
        name: do-not-modify-me
        env:
        - name: foo
          value: bar
      - image: nginx
        name: main
        env:
        - name: foo
          value: bar
      - image: nginx
        name: sidecar
`,
			replacements: `replacements:
- source:
    kind: Deployment
    name: sample-deploy
    fieldPath: metadata.name
  targets:
  - select:
      kind: Deployment
    options:
      create: true
    fieldPaths:
    - spec.template.spec.containers.[image=nginx].env.[name=deployment-name].value
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: other
        name: do-not-modify-me
        env:
        - name: foo
          value: bar
      - image: nginx
        name: main
        env:
        - name: foo
          value: bar
        - name: deployment-name
          value: sample-deploy
      - image: nginx
        name: sidecar
        env:
        - name: deployment-name
          value: sample-deploy`,
		},
		"index contains '*' character": {
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: deployment-name
          value: XXXXX
`,
			replacements: `replacements:
- source:
    kind: Deployment
    name: sample-deploy
    fieldPath: metadata.name
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.*.env.[name=deployment-name].value
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: deployment-name
          value: sample-deploy`,
		},
		"list index contains '*' character": {
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: deployment-name
          value: XXXXX
        - name: foo
          value: bar
      - image: nginx
        name: sidecar
        env:
        - name: deployment-name
          value: YYYYY
`,
			replacements: `replacements:
- source:
    kind: Deployment
    name: sample-deploy
    fieldPath: metadata.name
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.*.env.[name=deployment-name].value
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: deployment-name
          value: sample-deploy
        - name: foo
          value: bar
      - image: nginx
        name: sidecar
        env:
        - name: deployment-name
          value: sample-deploy`,
		},
		"index contains '*' character and create options": {
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: other-env
          value: YYYYY
`,
			replacements: `replacements:
- source:
    kind: Deployment
    name: sample-deploy
    fieldPath: metadata.name
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.*.env.[name=deployment-name].value
    options:
      create: true
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  template:
    metadata:
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: nginx
        name: main
        env:
        - name: other-env
          value: YYYYY
        - name: deployment-name
          value: sample-deploy
`,
		},
		"Issue 1493: wildcard to create or replace field in all containers in all workloads": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: policy
data:
  restart: OnFailure
---
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - image: nginx
    name: main
  - image: nginx
    name: sidecar
    imagePullPolicy: Always
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  containers:
  - image: nginx
    name: main
    imagePullPolicy: Always
  - image: nginx
    name: sidecar
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: policy
    fieldPath: data.restart
  targets:
  - select:
      kind: Pod
    fieldPaths:
    - spec.containers.*.imagePullPolicy
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: policy
data:
  restart: OnFailure
---
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - image: nginx
    name: main
    imagePullPolicy: OnFailure
  - image: nginx
    name: sidecar
    imagePullPolicy: OnFailure
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  containers:
  - image: nginx
    name: main
    imagePullPolicy: OnFailure
  - image: nginx
    name: sidecar
    imagePullPolicy: OnFailure
`,
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
		"string source -> integer target": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  PORT: "8080"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - containerPort: 80
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: config
    fieldPath: data.PORT
  targets:
  - select:
      kind: Pod
    fieldPaths:
    - spec.containers.0.ports.0.containerPort
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  PORT: "8080"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - containerPort: 8080
`,
		},
		"string source -> boolean target": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  MOUNT_TOKEN: "true"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    automountServiceAccountToken: false
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: config
    fieldPath: data.MOUNT_TOKEN
  targets:
  - select:
      kind: Pod
    fieldPaths:
    - spec.containers.0.automountServiceAccountToken
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  MOUNT_TOKEN: "true"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    automountServiceAccountToken: true
`,
		},
		// TODO: This is inconsistent with expectations; creating a numerical string would be
		// expected, unless we had knowledge of the intended type of the field to be
		// created.
		"numerical string source -> integer creation": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  PORT: "8080"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - protocol: TCP
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: config
    fieldPath: data.PORT
  targets:
  - select:
      kind: Pod
    fieldPaths:
    - spec.containers.0.ports.0.containerPort
    - spec.containers.0.ports.0.name
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  PORT: "8080"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - protocol: TCP
      containerPort: 8080
      name: 8080
`,
		},
		"integer source -> string target": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  PORT: "8080"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - containerPort: 80
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers.0.ports.0.containerPort
  targets:
  - select:
      kind: ConfigMap
    fieldPaths:
    - data.PORT
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  PORT: "80"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - containerPort: 80
`,
		},
		"boolean source -> string target": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  MOUNT_TOKEN: "true"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    automountServiceAccountToken: false
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers.0.automountServiceAccountToken
  targets:
  - select:
      kind: ConfigMap
    fieldPaths:
    - data.MOUNT_TOKEN
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  MOUNT_TOKEN: "false"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    automountServiceAccountToken: false
`,
		},
		// TODO: This result is expected, but we should create a string and not an
		// integer if we could know that the target type should be one. As a result,
		// the actual ConfigMap produces here cannot be applied.
		"integer source -> integer creation": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  FOO: "Bar"
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - containerPort: 80
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers.0.ports.0.containerPort
  targets:
  - select:
      kind: ConfigMap
    fieldPaths:
    - data.PORT
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  FOO: "Bar"
  PORT: 80
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
    ports:
    - containerPort: 80
`,
		},
		"replace an existing mapping value": {
			input: `apiVersion: batch/v1
kind: Job
metadata:
  name: hello
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
      restartPolicy: old
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - image: busybox
      name: myapp-container
  restartPolicy: new
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: my-pod
    fieldPath: spec
  targets:
  - select:
      name: hello
      kind: Job
    fieldPaths:
     - spec.template.spec
    options:
      create: true
`,
			expected: `apiVersion: batch/v1
kind: Job
metadata:
  name: hello
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
      restartPolicy: new
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: busybox
    name: myapp-container
  restartPolicy: new
`,
		},
		"issue4761 - path not in target with create: true": {
			input: `
---
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: request-id
spec:
  configPatches:
    - applyTo: NETWORK_FILTER
    - applyTo: NETWORK_FILTER
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-version
  annotations:
    config.kubernetes.io/local-config: true
data:
  ISTIO_REGEX: '^1\.14.*'
`,
			replacements: `
replacements:
  - source:
      kind: ConfigMap
      name: istio-version
      fieldPath: data.ISTIO_REGEX
    targets:
      - select:
          kind: EnvoyFilter
        fieldPaths:
          - spec.configPatches.0.match.proxy.proxyVersion
          - spec.configPatches.1.match.proxy.proxyVersion
          - spec.configPatches.2.match.proxy.proxyVersion
          - spec.configPatches.3.match.proxy.proxyVersion
        options:
          create: true
`,
			expected: `
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: request-id
spec:
  configPatches:
  - applyTo: NETWORK_FILTER
    match:
      proxy:
        proxyVersion: ^1\.14.*
  - applyTo: NETWORK_FILTER
    match:
      proxy:
        proxyVersion: ^1\.14.*
  - match:
      proxy:
        proxyVersion: ^1\.14.*
  - match:
      proxy:
        proxyVersion: ^1\.14.*
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-version
  annotations:
    config.kubernetes.io/local-config: true
data:
  ISTIO_REGEX: '^1\.14.*'
`,
		},
		"issue4761 - path not in target with create: false": {
			input: `
---
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: request-id
spec:
  configPatches:
    - applyTo: NETWORK_FILTER
    - applyTo: NETWORK_FILTER
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-version
  annotations:
    config.kubernetes.io/local-config: true
data:
  ISTIO_REGEX: '^1\.14.*'
`,
			replacements: `
replacements:
  - source:
      kind: ConfigMap
      name: istio-version
      fieldPath: data.ISTIO_REGEX
    targets:
      - select:
          kind: EnvoyFilter
        fieldPaths:
          - spec.configPatches.0.match.proxy.proxyVersion
          - spec.configPatches.1.match.proxy.proxyVersion
          - spec.configPatches.2.match.proxy.proxyVersion
          - spec.configPatches.3.match.proxy.proxyVersion
        options:
          create: false
`,
			expectedErr: "unable to find field \"spec.configPatches.0.match.proxy.proxyVersion\" in replacement target",
		},
		"issue4761 - wildcard solution": {
			input: `
---
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: request-id
spec:
  configPatches:
    - applyTo: NETWORK_FILTER
    - applyTo: NETWORK_FILTER
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-version
  annotations:
    config.kubernetes.io/local-config: true
data:
  ISTIO_REGEX: '^1\.14.*'
`,
			replacements: `
replacements:
  - source:
      kind: ConfigMap
      name: istio-version
      fieldPath: data.ISTIO_REGEX
    targets:
      - select:
          kind: EnvoyFilter
        fieldPaths:
          - spec.configPatches.*.match.proxy.proxyVersion
        options:
          create: true
`,
			expected: `
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: request-id
spec:
  configPatches:
  - applyTo: NETWORK_FILTER
    match:
      proxy:
        proxyVersion: ^1\.14.*
  - applyTo: NETWORK_FILTER
    match:
      proxy:
        proxyVersion: ^1\.14.*
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-version
  annotations:
    config.kubernetes.io/local-config: true
data:
  ISTIO_REGEX: '^1\.14.*'
`,
		},
		"append to sequence using index": {
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myingress
spec:
  rules:
    - host: myingress.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-svc
                port:
                  number: 80
`,
			replacements: `replacements:
  - source:
      kind: Ingress
      name: myingress
      fieldPath: spec.rules.0.host
    targets:
      - select:
          kind: Ingress
          name: myingress
        fieldPaths:
          - spec.tls.0.hosts.0
          - spec.tls.0.secretName          
        options:
          create: true
`,
			expected: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myingress
spec:
  rules:
  - host: myingress.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-svc
            port:
              number: 80
  tls:
  - hosts:
    - myingress.example.com
    secretName: myingress.example.com`,
		},
		"fail to append to sequence using a distant index": {
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myingress
spec:
  rules:
    - host: myingress.example.com
`,
			replacements: `replacements:
  - source:
      kind: Ingress
      name: myingress
      fieldPath: spec.rules.0.host
    targets:
      - select:
          kind: Ingress
          name: myingress
        fieldPaths:
          - spec.tls.5.hosts.5
          - spec.tls.5.secretName
        options:
          create: true
`,
			expectedErr: "unable to find or create field \"spec.tls.5.hosts.5\" in replacement target: index 5 specified but only 0 elements found",
		}, "replace with static value": {
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
- sourceValue: custom/postgres:1.9.0
  targets:
  - select:
      kind: Deployment
      name: deploy
    fieldPaths:
    - spec.template.spec.containers.[name=postgresdb].image
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
      - image: custom/postgres:1.9.0
        name: postgresdb
`,
		}, "source value and selector error": {
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: deploy
spec:
  replicas: 1
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
  sourceValue: 2
  targets:
  - select:
      kind: Deployment
      name: deploy
    fieldPaths:
    - spec.replicas
`,
			expectedErr: "value and resource selectors are mutually exclusive",
		},

		// -- regex --
		"select 1 from many using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  PORT: 8080
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
        args:
        - PORT
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - PORT
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
        args:
        - PORT
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.PORT
  targets:
  - select:
      name: d.*2
    fieldPaths:
    - spec.template.spec.containers.0.args.0
`,
			expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  PORT: 8080
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
        args:
        - PORT
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - "8080"
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
        args:
        - PORT
`,
		},

		"select 2 among 3 using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  PORT: 8080
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
        args:
        - PORT
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - PORT
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
        args:
        - PORT
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.PORT
  targets:
  - select:
      name: d.*[1-2]
    fieldPaths:
    - spec.template.spec.containers.0.args.0
`,
			expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  PORT: 8080
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
        args:
        - "8080"
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - "8080"
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
        args:
        - PORT
`,
		},

		"select 2 among 3 and create one field using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "server.svc"
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
        args:
        - HOST
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - HOST
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      name: d.*[2-3]
    fieldPaths:
    - spec.template.spec.containers.0.args.0
    options:
      create: true
`,
			expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "server.svc"
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
        args:
        - HOST
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - server.svc
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
        args:
        - server.svc
`,
		},

		"select 2 of 3 and create all the fields using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "server.svc"
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      name: d.*[2-3]
    fieldPaths:
    - spec.template.spec.containers.0.args.0
    options:
      create: true
`,
			expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "server.svc"
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1
        name: nginx1
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: nginx:2
        name: nginx1
        args:
        - server.svc
---
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: nginx:3
        name: nginx1
        args:
        - server.svc
`,
		},

		"select multiple targets and create multiple fields using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "server.svc"
  OPT: "--debug=true"
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: tomcat:1
        name: tomcat1
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: tomcat:2
        name: tomcat2
        args:
        - tomcat.svc
---
kind: StatefulSet
metadata:
  name: statefulset1
spec:
  template:
    spec:
      containers:
      - image: pg:1
        name: pg1
---
kind: State
metadata:
  name: state2
spec:
  template:
    spec:
      containers:
      - image: pg:2
        name: pg2
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  # create multiple fields, multiple targets, in multiple places
  - select:
      # should match deploy1 deploy2
      name: d.*[1-9]
      kind: Deployment|State
    fieldPaths:
    - spec.template.spec.containers.0.args.0
    - spec.template.spec.containers.1.name
    options:
      create: true
  - select:
      # should match state2, should not match statefulset1
      name: s.*[1-9]
      kind: Deployment|State
    fieldPaths:
    - metadata.annotations.service
    - spec.template.spec.containers.0.args.0
    options:
      create: true

- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.OPT
  targets:
  - select:
      # should match State/state2, should not match other kinds
      name: s.*[0-9]
      kind: State|fulSet|Set|ful|Dep.*
    fieldPaths:
    - spec.template.spec.containers.*.args.1
    options:
      create: true
`,
			expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "server.svc"
  OPT: "--debug=true"
---
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: tomcat:1
        name: tomcat1
        args:
        - server.svc
      - name: server.svc
---
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: tomcat:2
        name: tomcat2
        args:
        - server.svc
      - name: server.svc
---
kind: StatefulSet
metadata:
  name: statefulset1
spec:
  template:
    spec:
      containers:
      - image: pg:1
        name: pg1
---
kind: State
metadata:
  name: state2
  annotations:
    service: server.svc
spec:
  template:
    spec:
      containers:
      - image: pg:2
        name: pg2
        args:
        - server.svc
        - --debug=true
`,
		},

		"name not matching suffixed and prefixed names using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: prefix-deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: deploy1-suffix
  annotations:
    service: SERVICE
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      # should match deploy1-suffix
      name: dep.*[1-9]-suffix
      kind: Deployment|State
    fieldPaths:
    - metadata.annotations.service
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
    options:
      delimiter: '.'
  targets:
  - select:
      # should match prefix-deploy1
      name: pref.*-dep.*
    fieldPaths:
    - metadata.annotations.service
    options:
      delimiter: '.'
      index: -1
      create: true
`,
			expected: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: prefix-deploy1
  annotations:
    service: tomcat.SERVICE
---
kind: Deployment
metadata:
  name: deploy1-suffix
  annotations:
    service: tomcat.svc
`,
		},

		"not matching source error using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: prefix-deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: deploy1-suffix
  annotations:
    service: SERVICE
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm.*
    fieldPath: data.HOST
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - metadata.annotations.service
`,
			expectedErr: `nothing selected by ConfigMap.[noVer].[noGrp]/cm.*.[noNs]:data.HOST`,
		},

		"not matching target no error using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: prefix-deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: deploy1-suffix
  annotations:
    service: SERVICE
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      name: .*z.*
      kind: Deployment|State
    fieldPaths:
    - metadata.annotations.service
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: prefix-deploy1
  annotations:
    service: SERVICE
---
kind: Deployment
metadata:
  name: deploy1-suffix
  annotations:
    service: SERVICE
`,
		},

		"reduce selection with matching label using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  labels:
    target: selectme
---
kind: Deployment
metadata:
  name: deploy2
---
kind: Deployment
metadata:
  name: deploy3
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      # should match deploy1
      name: .*depl.*
      labelSelector:
        target=selectme
    fieldPaths:
    - metadata.annotations.service
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  labels:
    target: selectme
  annotations:
    service: tomcat.svc
---
kind: Deployment
metadata:
  name: deploy2
---
kind: Deployment
metadata:
  name: deploy3
`,
		},

		"select label not matching name using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  labels:
    target: selectme
---
kind: Deployment
metadata:
  name: deploy2
---
kind: Deployment
metadata:
  name: deploy3
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      name: .*notmatching.*
      labelSelector:
        target=selectme
    fieldPaths:
    - metadata.annotations.service
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
  labels:
    target: selectme
---
kind: Deployment
metadata:
  name: deploy2
---
kind: Deployment
metadata:
  name: deploy3
`,
		},

		"not matching substring using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
---
kind: Deployment
metadata:
  name: deploy2
---
kind: Deployment
metadata:
  name: mydeploy3
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      # should not match deploy1 deploy2 mydeploy3
      name: deploy
    fieldPaths:
    - metadata.annotations.service
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
kind: Deployment
metadata:
  name: deploy1
---
kind: Deployment
metadata:
  name: deploy2
---
kind: Deployment
metadata:
  name: mydeploy3
`,
		},

		"select by group using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
apiVersion: g1/v1
kind: Deployment
metadata:
  name: deploy1
---
apiVersion: g2/v1
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: g3/v1
kind: Deployment
metadata:
  name: deploy3
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      # should match deploy1 and deploy2
      name: .*depl.*
      group: g[13]
    fieldPaths:
    - metadata.annotations.service
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
apiVersion: g1/v1
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: tomcat.svc
---
apiVersion: g2/v1
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: g3/v1
kind: Deployment
metadata:
  name: deploy3
  annotations:
    service: tomcat.svc
`,
		},

		"select by version using regex": {
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
apiVersion: g1/v1
kind: Deployment
metadata:
  name: deploy1
---
apiVersion: g1/v2
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: g1/v3
kind: Deployment
metadata:
  name: deploy3
`,
			replacements: `replacements:
- source:
    kind: ConfigMap
    name: cm
    fieldPath: data.HOST
  targets:
  - select:
      # should match deploy1 and deploy2
      name: .*depl.*
      version: v[13]
    fieldPaths:
    - metadata.annotations.service
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOST: "tomcat.svc"
---
apiVersion: g1/v1
kind: Deployment
metadata:
  name: deploy1
  annotations:
    service: tomcat.svc
---
apiVersion: g1/v2
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: g1/v3
kind: Deployment
metadata:
  name: deploy3
  annotations:
    service: tomcat.svc
`,
		},

		"reject 1 with regex": {
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
      kind: Deploy.*
    reject:
    - name: .*ploy2|deploy3
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
		"reject 2 with regex": {
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
    - kind: Deplo.*
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
		// is the dash before `name: my-name`
		"reject 3 with regex": {
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
    - name: my.n.me
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
		"simple with regex": {
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
---
kind: StatefulSet
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
      kind: Deployment|StatefulSet
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
---
kind: StatefulSet
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
		"create using regex": {
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
---
apiVersion: apps/v1
kind: NotDeployment
metadata:
  name: notdeploy
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers
  targets:
  - select:
      name: depl.*
    fieldPaths:
    - spec.template.spec.containers
    options:
      create: true
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
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
---
apiVersion: apps/v1
kind: NotDeployment
metadata:
  name: notdeploy

`,
		},
		"create in source using regex": {
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
---
apiVersion: apps/v1
kind: NotDeployment
metadata:
  name: notdeploy
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers
  targets:
  - select:
      name: depl.*
    fieldPaths:
    - spec.template.spec.containers
    options:
      create: true
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
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
---
apiVersion: apps/v1
kind: NotDeployment
metadata:
  name: notdeploy

`,
		},
		"delimiter using regex": {
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod
  labels:
    number: zero-one
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
  labels:
    number: one
---
apiVersion: apps/v1
kind: NotDeployment
metadata:
  name: notdeploy
`,
			replacements: `replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: metadata.labels.number
    options:
      delimiter: "-"
      index: 1
  targets:
  - select:
      name: depl.*
    fieldPaths:
    - metadata.labels.number
    options:
      create: true
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: pod
  labels:
    number: zero-one
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
  labels:
    number: one
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
  labels:
    number: one
---
apiVersion: apps/v1
kind: NotDeployment
metadata:
  name: notdeploy
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
