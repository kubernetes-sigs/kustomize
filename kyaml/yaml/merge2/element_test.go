// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var elementTestCases = []testCase{
	{description: `merge Element -- keep field in dest`,
		source: `
kind: Deployment
items:
- name: foo
  image: foo:v1
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v0
  command: ['run.sh']
`,
		expected: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
	},

	{description: `merge Element -- add field to dest`,
		source: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v0
`,
		expected: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
	},

	{description: `merge Element -- add list, empty in dest`,
		source: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
		dest: `
kind: Deployment
items: []
`,
		expected: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
	},

	{description: `merge Element -- add list, missing from dest`,
		source: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
		dest: `
kind: Deployment
`,
		expected: `
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
	},

	{description: `merge Element -- add Element first`,
		source: `
kind: Deployment
items:
- name: bar
  image: bar:v1
  command: ['run2.sh']
- name: foo
  image: foo:v1
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v0
`,
		expected: `
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
	},

	{description: `merge Element -- add Element second`,
		source: `
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v0
`,
		expected: `
kind: Deployment
items:
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
kind: Deployment
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		expected: `
kind: Deployment
items:
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
kind: Deployment
items:
- name: foo
  image: foo:v1
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v0
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		expected: `
kind: Deployment
items:
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
kind: Deployment
items: {}
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		expected: `
kind: Deployment
items:
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
kind: Deployment
items: null
`,
		dest: `
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		expected: `
kind: Deployment
`,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys no merge'`,
		source: `
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		dest: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		expected: `
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		noInfer: true,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using schema`,
		source: `
kind: Deployment
apiVersion: apps/v1
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		dest: `
kind: Deployment
apiVersion: apps/v1
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		expected: `
kind: Deployment
apiVersion: apps/v1
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run2.sh']
`,
		noInfer: true,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema as line comment'`,
		source: `
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		dest: `
kind: Deployment
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: foo # hell ow
  image: foo:bar
`,
		expected: `
kind: Deployment
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		noInfer: true,
	},
}
