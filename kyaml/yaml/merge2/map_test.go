// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var mapTestCases = []testCase{
	{description: `merge Map -- update field in dest`,
		source: `
kind: Deployment
spec:
  foo: bar1
`,
		dest: `
kind: Deployment
spec:
  foo: bar0
  baz: buz
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{description: `merge Map -- add field to dest`,
		source: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		dest: `
kind: Deployment
spec:
  foo: bar0
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{description: `merge Map -- add list, empty in dest`,
		source: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		dest: `
kind: Deployment
spec: {}
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{description: `merge Map -- add list, missing from dest`,
		source: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		dest: `
kind: Deployment
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{description: `merge Map -- add Map first`,
		source: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		dest: `
kind: Deployment
spec:
  foo: bar1
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{description: `merge Map -- add Map second`,
		source: `
kind: Deployment
spec:
  baz: buz
  foo: bar1
`,
		dest: `
kind: Deployment
spec:
  foo: bar1
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	//
	// Test Case
	//
	{description: `keep map -- map missing from src`,
		source: `
kind: Deployment
`,
		dest: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	//
	// Test Case
	//
	{description: `keep map -- empty list in src`,
		source: `
kind: Deployment
items: {}
`,
		dest: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		expected: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
items: {}
`,
	},

	//
	// Test Case
	//
	{description: `remove Map -- null in src`,
		source: `
kind: Deployment
spec: null
`,
		dest: `
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		expected: `
kind: Deployment
`,
	},
}
