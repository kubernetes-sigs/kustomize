// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var mapTestCases = []testCase{
	{`merge Map -- update field in dest`,
		`
kind: Deployment
spec:
  foo: bar1
`,
		`
kind: Deployment
spec:
  foo: bar0
  baz: buz
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{`merge Map -- add field to dest`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
kind: Deployment
spec:
  foo: bar0
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{`merge Map -- add list, empty in dest`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
kind: Deployment
spec: {}
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{`merge Map -- add list, missing from dest`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
kind: Deployment
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{`merge Map -- add Map first`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
kind: Deployment
spec:
  foo: bar1
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	{`merge Map -- add Map second`,
		`
kind: Deployment
spec:
  baz: buz
  foo: bar1
`,
		`
kind: Deployment
spec:
  foo: bar1
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	//
	// Test Case
	//
	{`keep map -- map missing from src`,
		`
kind: Deployment
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
	},

	//
	// Test Case
	//
	{`keep map -- empty list in src`,
		`
kind: Deployment
items: {}
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
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
	{`remove Map -- null in src`,
		`
kind: Deployment
spec: null
`,
		`
kind: Deployment
spec:
  foo: bar1
  baz: buz
`,
		`
kind: Deployment
`,
	},
}
