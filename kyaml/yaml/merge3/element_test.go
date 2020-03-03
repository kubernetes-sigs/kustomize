// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var elementTestCases = []testCase{
	//
	// Test Case
	//
	{description: `Add an element to an existing list`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:1
`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:1
- name: baz
  image: baz:2
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:1
`,
		expected: `
kind: Deployment
containers:
- name: foo
  image: foo:1
- image: baz:2
  name: baz
  
`},

	//
	// Test Case
	//
	{description: `Add an element to a non-existing list`,
		origin: `
kind: Deployment`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		local: `
kind: Deployment
`,
		expected: `
kind: Deployment
containers:
- image: foo:bar
  name: foo
`},

	{description: `Add an element to a non-existing list, existing in dest`,
		origin: `
kind: Deployment`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		local: `
kind: Deployment
containers:
- name: baz
  image: baz:bar
`,
		expected: `
kind: Deployment
containers:
- name: baz
  image: baz:bar
- image: foo:bar
  name: foo
`},

	//
	// Test Case
	// TODO(pwittrock): Figure out if there is something better we can do here
	// This element is missing from the destination -- only the new fields are added
	{description: `Add a field to the element, element missing from dest`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command:
  - run.sh
`,
		local: `
kind: Deployment
`,
		expected: `
kind: Deployment
containers:
- command:
  - run.sh
  name: foo
`},

	//
	// Test Case
	//
	{description: `Update a field on the elem, element missing from the dest`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command:
  - run.sh
`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: 
  - run2.sh
`,
		local: `
kind: Deployment
`,
		expected: `
kind: Deployment
containers:
- command:
  - run2.sh
  name: foo
`},

	//
	// Test Case
	//
	{description: `Update a field on the elem, element present in the dest`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		expected: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`},

	//
	// Test Case
	//
	{description: `Add a field on the elem, element present in the dest`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		expected: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`},

	//
	// Test Case
	//
	{description: `Add a field on the elem, element and field present in the dest`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		expected: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`},

	//
	// Test Case
	//
	{description: `Ignore an element`,
		origin: `
kind: Deployment
containers: {}
`,
		update: `
kind: Deployment
containers: {}
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		expected: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`},

	//
	// Test Case
	//
	{description: `Leave deleted`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
`,
		local: `
kind: Deployment
`,
		expected: `
kind: Deployment
`},

	//
	// Test Case
	//
	{description: `Remove an element -- matching`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		expected: `
kind: Deployment
`},

	//
	// Test Case
	//
	{description: `Remove an element -- field missing from update`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		expected: `
kind: Deployment
`},

	//
	// Test Case
	//
	{description: `Remove an element -- element missing`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
- name: baz
  image: baz:bar
`,
		update: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
- name: baz
  image: baz:bar
`,
		expected: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`},

	//
	// Test Case
	//
	{description: `Remove an element -- empty containers`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
containers: {}
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		expected: `
kind: Deployment
`},

	//
	// Test Case
	//
	{description: `Remove an element -- missing list field`,
		origin: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		update: `
kind: Deployment
`,
		local: `
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		expected: `
kind: Deployment
`},

	//
	// Test Case
	//
	{description: `no infer merge keys no merge'`,
		origin: `
kind: Deployment
containers:
- name: foo
`,
		update: `
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		local: `
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
		origin: `
kind: Deployment
apiVersion: apps/v1
spec:
  template:
    spec:
      containers:
      - name: foo
`,
		update: `
kind: Deployment
apiVersion: apps/v1
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		local: `
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
		origin: `
kind: Deployment
containers:
- name: foo
`,
		update: `
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		local: `
kind: Deployment
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: foo # hell ow
  image: foo:bar
`,
		expected: `
kind: Deployment
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: foo # hell ow
  image: foo:bar
  command: ['run2.sh']
`,
		noInfer: true,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema as head comment'`,
		origin: `
kind: Deployment
containers:
- name: foo
`,
		update: `
kind: Deployment
containers:
- name: foo
  command: ['run2.sh']
`,
		local: `
kind: Deployment
# {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
containers:
- name: foo # hell ow
  image: foo:bar
`,
		expected: `
kind: Deployment
# {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
containers:
- name: foo # hell ow
  image: foo:bar
  command: ['run2.sh']
`,
		noInfer: true,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema to parent field'`,
		origin: `
kind: Deployment
spec:
  containers:
  - name: foo
`,
		update: `
kind: Deployment
spec:
  containers:
  - name: foo
    command: ['run2.sh']
`,
		local: `
kind: Deployment
spec: # {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
  containers:
  - name: foo # hell ow
    image: foo:bar
`,
		expected: `
kind: Deployment
spec: # {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
  containers:
  - name: foo # hell ow
    image: foo:bar
    command: ['run2.sh']
`,
		noInfer: true,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema to parent field header'`,
		origin: `
kind: Deployment
spec:
  containers:
  - name: foo
`,
		update: `
kind: Deployment
spec:
  containers:
  - name: foo
    command: ['run2.sh']
`,
		local: `
kind: Deployment
# {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
spec:
  containers:
  - name: foo # hell ow
    image: foo:bar
`,
		expected: `
kind: Deployment
# {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
spec:
  containers:
  - name: foo # hell ow
    image: foo:bar
    command: ['run2.sh']
`,
		noInfer: true,
	},
}
