// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var elementTestCases = []testCase{
	{description: `merge Element -- keep field in dest`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v0
        command: ['run.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
	},

	{description: `merge Element -- add field to dest`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v0
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
	},

	{description: `merge Element -- add list, empty in dest`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: []
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
	},

	{description: `merge Element -- add list, missing from dest`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
        command: ['run.sh']
`,
	},

	{description: `merge Element -- add Element first`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: bar
        image: bar:v1
        command: ['run2.sh']
      - name: foo
        image: foo:v1
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v0
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
	},

	{description: `merge Element -- add Element second`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v0
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
	},

	//
	// Test Case
	//
	{description: `keep list -- list missing from src`,
		source: `
apiVersion: apps/v1
kind: Deployment
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
	},

	//
	// Test Case
	//
	{description: `keep Element -- element missing in src`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v0
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
	},

	//
	// Test Case
	//
	{description: `keep element -- empty list in src`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: {}
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
	},

	//
	// Test Case
	//
	{description: `remove Element -- null in src`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: null
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:v1
      - name: bar
        image: bar:v1
        command: ['run2.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`,
	},

	//
	// Test Case
	//
	{description: `infer merge keys merge'`,
		source: `
apiVersion: custom
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		dest: `
apiVersion: custom
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		expected: `
apiVersion: custom
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		infer: true,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using schema`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run2.sh']
`,
		infer: false,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema as line comment'`,
		source: `
apiVersion: custom
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		dest: `
apiVersion: custom
kind: Deployment
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: foo # hell ow
  image: foo:bar
`,
		expected: `
apiVersion: custom
kind: Deployment
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		infer: false,
	},
}
