// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var scalarTestCases = []testCase{
	{`replace scalar -- different value in dest`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
field: value0
`,
		`
kind: Deployment
field: value1
`,
	},

	{`replace scalar -- missing from dest`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
`,
		`
kind: Deployment
field: value1
`,
	},

	//
	// Test Case
	//
	{`keep scalar -- same value in src and dest`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
field: value1
`,
	},

	//
	// Test Case
	//
	{`keep scalar -- unspecified in src`,
		`
kind: Deployment
`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
field: value1
`,
	},

	//
	// Test Case
	//
	{`remove scalar -- null in src`,
		`
kind: Deployment
field: null
`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
`,
	},

	//
	// Test Case
	//
	{`remove scalar -- empty in src`,
		`
kind: Deployment
field: {}
`,
		`
kind: Deployment
field: value1
`,
		`
kind: Deployment
field: {}
`,
	},

	//
	// Test Case
	//
	{`remove scalar -- null in src, missing in dest`,
		`
kind: Deployment
field: null
`,
		`
kind: Deployment
`,
		`
kind: Deployment
`,
	},

	//
	// Test Case
	//
	{`merge an empty value`,
		`
kind: Deployment
field: {}
`,
		`
kind: Deployment
`,
		`
kind: Deployment
field: {}
`,
	},
}
