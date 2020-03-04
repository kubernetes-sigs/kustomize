// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

//nolint:lll
var elementTestCases = []testCase{
	//
	// Test Case
	//
	{description: `Add an element to an existing list`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:1
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:1
      - name: baz
        image: baz:2
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:1
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
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
apiVersion: apps/v1
kind: Deployment`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		local: `
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
      - image: foo:bar
        name: foo
`},

	{description: `Add an element to a non-existing list, existing in dest`,
		origin: `
apiVersion: apps/v1
kind: Deployment`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: baz
        image: baz:bar
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
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
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command:
        - run.sh
`,
		local: `
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
      - command:
        - run.sh
        name: foo
`},

	//
	// Test Case
	//
	{description: `Update a field on the elem, element missing from the dest`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command:
        - run.sh
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command:
        - run2.sh
`,
		local: `
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
      - command:
        - run2.sh
        name: foo
`},

	//
	// Test Case
	//
	{description: `Update a field on the elem, element present in the dest`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run.sh']
`,
		update: `
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
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
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
        image: foo:bar
        command: ['run2.sh']
`},

	//
	// Test Case
	//
	{description: `Add a field on the elem, element present in the dest`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
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
		local: `
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
`},

	//
	// Test Case
	//
	{description: `Add a field on the elem, element and field present in the dest`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
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
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
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
        image: foo:bar
        command: ['run2.sh']
`},

	//
	// Test Case
	//
	{description: `Ignore an element`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: {}
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: {}
`,
		local: `
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
`},

	//
	// Test Case
	//
	{description: `Leave deleted`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
apiVersion: apps/v1
kind: Deployment
`,
		local: `
apiVersion: apps/v1
kind: Deployment
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
`},

	//
	// Test Case
	//
	{description: `Remove an element -- matching`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`,
		local: `
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
    spec: {}
`},

	//
	// Test Case
	//
	{description: `Remove an element -- field missing from update`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`},

	//
	// Test Case
	//
	{description: `Remove an element -- element missing`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
      - name: baz
        image: baz:bar
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run.sh']
      - name: baz
        image: baz:bar
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
        command: ['run.sh']
`},

	//
	// Test Case
	//
	{description: `Remove an element -- empty containers`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: {}
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`},

	//
	// Test Case
	//
	{description: `Remove an element -- missing list field`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
        command: ['run.sh']
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec: {}
`},

	//
	// Test Case
	//
	{description: `infer merge keys merge'`,
		origin: `
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
`,
		update: `
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		local: `
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        image: foo:bar
`,
		expected: `
kind: Deployment
spec:
  template:
    spec:
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
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		local: `
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
		origin: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
`,
		update: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		local: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
      - name: foo # hell ow
        image: foo:bar
`,
		expected: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
      - name: foo # hell ow
        image: foo:bar
        command: ['run2.sh']
`,
		infer: false,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema as head comment'`,
		origin: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
`,
		update: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        command: ['run2.sh']
`,
		local: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
      containers:
      - name: foo # hell ow
        image: foo:bar
`,
		expected: `
apiVersion: custom
kind: Deployment
spec:
  template:
    spec:
      # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
      containers:
      - name: foo # hell ow
        image: foo:bar
        command: ['run2.sh']
`,
		infer: false,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema to parent field'`,
		origin: `
apiVersion: custom
kind: Deployment
spec:
  containers:
  - name: foo
`,
		update: `
apiVersion: custom
kind: Deployment
spec:
  containers:
  - name: foo
    command: ['run2.sh']
`,
		local: `
apiVersion: custom
kind: Deployment
spec: # {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
  containers:
  - name: foo # hell ow
    image: foo:bar
`,
		expected: `
apiVersion: custom
kind: Deployment
spec: # {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
  containers:
  - name: foo # hell ow
    image: foo:bar
    command: ['run2.sh']
`,
		infer: false,
	},

	//
	// Test Case
	//
	{description: `no infer merge keys merge using explicit schema to parent field header'`,
		origin: `
apiVersion: custom
kind: Deployment
spec:
  containers:
  - name: foo
`,
		update: `
apiVersion: custom
kind: Deployment
spec:
  containers:
  - name: foo
    command: ['run2.sh']
`,
		local: `
apiVersion: custom
kind: Deployment
# {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
spec:
  containers:
  - name: foo # hell ow
    image: foo:bar
`,
		expected: `
apiVersion: custom
kind: Deployment
# {"$ref":"#/definitions/io.k8s.api.core.v1.PodSpec"}
spec:
  containers:
  - name: foo # hell ow
    image: foo:bar
    command: ['run2.sh']
`,
		infer: false,
	},
}
