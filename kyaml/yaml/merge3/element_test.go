// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

//nolint:lll
var elementTestCases = []testCase{
	//
	// Test Case
	//
	{
		description: `Add an element to an existing list`,
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
	{
		description: `Add an element to a non-existing list`,
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

	{
		description: `Add an element to a non-existing list, existing in dest`,
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
	{
		description: `Add a field to the element, element missing from dest`,
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
	{
		description: `Update a field on the elem, element missing from the dest`,
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
	{
		description: `Update a field on the elem, element present in the dest`,
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
	{
		description: `Add a field on the elem, element and field present in the dest`,
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
	{
		description: `Ignore an element`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: null
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers: null
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
	{
		description: `Leave deleted`,
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
	{
		description: `Remove an element -- matching`,
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
	{
		description: `Remove an element -- field missing from update`,
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
	{
		description: `Remove an element -- element missing`,
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
	{
		description: `Remove an element -- empty containers`,
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
      containers: null
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
	{
		description: `Remove an element -- missing list field`,
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
	{
		description: `infer merge keys merge'`,
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
	{
		description: `no infer merge keys merge using schema`,
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
	{
		description: `no infer merge keys merge using explicit schema as line comment'`,
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
	{
		description: `no infer merge keys merge using explicit schema as head comment'`,
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
	{
		description: `no infer merge keys merge using explicit schema to parent field'`,
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
	{
		description: `no infer merge keys merge using explicit schema to parent field header'`,
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

	// The following test cases are regression tests
	// that should not be broken as a result of
	// #3111, #3159

	//
	// Test Case
	//
	{
		description: `Add a containerPort to an existing list`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
        - containerPort: 80
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
        - containerPort: 80
`},

	//
	// Test Case
	//
	{
		description: `Add a containerPort to a non-existing list, existing in dest`,
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
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
        - containerPort: 8080
`},

	//
	// Test Case
	//
	{
		description: `Add a name to containerPort`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          name: 8080-port-update
        - containerPort: 80
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
        - containerPort: 80
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          name: 8080-port-update
        - containerPort: 80
`},
	//
	// Test Case
	//
	{
		description: `Update protocol for a port`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`},
	//
	// Test Case
	//
	{
		description: `Append container port`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
          protocol: HTTP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
          protocol: HTTP
        - containerPort: 8080
          protocol: TCP
`},

	{
		description: `Update container-port name`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: foo
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
          name: bar
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
          name: foo
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
          name: bar
`},

	//
	// Test Case
	//
	{
		description: `Add a containerPort with protocol to an existing list`,
		origin: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
        - containerPort: 8080
          protocol: TCP
`},

	//
	// Test Case
	//
	{
		description: `Add a containerPort with protocol to a non-existing list, existing in dest`,
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
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
        - containerPort: 8080
          protocol: UDP
`},

	//
	// Test Case
	//
	{
		description: `Merge with name for same container-port`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: TCP
          name: updated
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: HTTP
          name: local
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: HTTP
          name: local
        - containerPort: 8080
          name: updated
          protocol: TCP
`},

	{
		description: `Retain local protocol`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: HTTP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: HTTP
        - containerPort: 8080
          protocol: TCP
`},
}
