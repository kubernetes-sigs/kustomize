// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var elementTestCases = []testCase{
	//
	// Test Case
	//
	{`Add an element to an existing list`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:1
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:1
- name: baz
  image: baz:2
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:1
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:1
- image: baz:2
  name: baz
  
`, nil},

	//
	// Test Case
	//
	{`Add an element to a non-existing list`,
		`
kind: Deployment`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
`,
		`
kind: Deployment
containers:
- image: foo:bar
  name: foo
`, nil},

	{`Add an element to a non-existing list, existing in dest`,
		`
kind: Deployment`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers:
- name: baz
  image: baz:bar
`,
		`
kind: Deployment
containers:
- name: baz
  image: baz:bar
- image: foo:bar
  name: foo
`, nil},

	//
	// Test Case
	// TODO(pwittrock): Figure out if there is something better we can do here
	// This element is missing from the destination -- only the new fields are added
	{`Add a field to the element, element missing from dest`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command:
  - run.sh
`,
		`
kind: Deployment
`,
		`
kind: Deployment
containers:
- command:
  - run.sh
  name: foo
`, nil},

	//
	// Test Case
	//
	{`Update a field on the elem, element missing from the dest`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command:
  - run.sh
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: 
  - run2.sh
`,
		`
kind: Deployment
`,
		`
kind: Deployment
containers:
- command:
  - run2.sh
  name: foo
`, nil},

	//
	// Test Case
	//
	{`Update a field on the elem, element present in the dest`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`, nil},

	//
	// Test Case
	//
	{`Add a field on the elem, element present in the dest`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`, nil},

	//
	// Test Case
	//
	{`Add a field on the elem, element and field present in the dest`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run2.sh']
`, nil},

	//
	// Test Case
	//
	{`Ignore an element`,
		`
kind: Deployment
containers: {}
`,
		`
kind: Deployment
containers: {}
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`, nil},

	//
	// Test Case
	//
	{`Leave deleted`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
`,
		`
kind: Deployment
`,
		`
kind: Deployment
`, nil},

	//
	// Test Case
	//
	{`Remove an element -- matching`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
`, nil},

	//
	// Test Case
	//
	{`Remove an element -- field missing from update`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		`
kind: Deployment
`, nil},

	//
	// Test Case
	//
	{`Remove an element -- element missing`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
- name: baz
  image: baz:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
- name: baz
  image: baz:bar
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`, nil},

	//
	// Test Case
	//
	{`Remove an element -- empty containers`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
containers: {}
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		`
kind: Deployment
`, nil},

	//
	// Test Case
	//
	{`Remove an element -- missing list field`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
`,
		`
kind: Deployment
`,
		`
kind: Deployment
containers:
- name: foo
  image: foo:bar
  command: ['run.sh']
`,
		`
kind: Deployment
`, nil},
}
