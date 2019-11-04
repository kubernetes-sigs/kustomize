// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var listTestCases = []testCase{
	{`replace List -- different value in dest`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
items:
- 0
- 1
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
	},

	{`replace List -- missing from dest`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
	},

	//
	// Test Case
	//
	{`keep List -- same value in src and dest`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
	},

	//
	// Test Case
	//
	{`keep List -- unspecified in src`,
		`
kind: Deployment
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
	},

	//
	// Test Case
	//
	{`remove List -- null in src`,
		`
kind: Deployment
items: null
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
`,
	},

	//
	// Test Case
	//
	{`remove list -- empty in src`,
		`
kind: Deployment
items: {}
`,
		`
kind: Deployment
items:
- 1
- 2
- 3
`,
		`
kind: Deployment
items: {}
`,
	},
}
