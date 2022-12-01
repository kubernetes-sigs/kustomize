// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var scalarTestCases = []testCase{
	{description: `replace scalar -- different value in dest`,
		source: `
kind: Deployment
field: value1
`,
		dest: `
kind: Deployment
field: value0
`,
		expected: `
kind: Deployment
field: value1
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	{description: `replace scalar -- missing from dest`,
		source: `
kind: Deployment
field: value1
`,
		dest: `
kind: Deployment
`,
		expected: `
kind: Deployment
field: value1
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `keep scalar -- same value in src and dest`,
		source: `
kind: Deployment
field: value1
`,
		dest: `
kind: Deployment
field: value1
`,
		expected: `
kind: Deployment
field: value1
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `keep scalar -- unspecified in src`,
		source: `
kind: Deployment
`,
		dest: `
kind: Deployment
field: value1
`,
		expected: `
kind: Deployment
field: value1
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `remove scalar -- null in src`,
		source: `
kind: Deployment
field: null
`,
		dest: `
kind: Deployment
field: value1
`,
		expected: `
kind: Deployment
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `remove scalar -- empty in src`,
		source: `
kind: Deployment
field: null
`,
		dest: `
kind: Deployment
field: value1
`,
		expected: `
kind: Deployment
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `remove scalar -- null in src, missing in dest`,
		source: `
kind: Deployment
field: null
`,
		dest: `
kind: Deployment
`,
		expected: `
kind: Deployment
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `remove scalar -- null in src, empty in dest`,
		source: `
kind: Deployment
field: null
`,
		dest: `
kind: Deployment
field:
`,
		expected: `
kind: Deployment
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `remove scalar -- null in src, null in dest`,
		source: `
kind: Deployment
field: null
`,
		dest: `
kind: Deployment
field: null
`,
		expected: `
kind: Deployment
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `keep scalar -- missing in src, null in dest`,
		source: `
kind: Deployment
`,
		dest: `
kind: Deployment
field: null
`,
		expected: `
kind: Deployment
field: null
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	//
	// Test Case
	//
	{description: `merge an empty value`,
		source: `
kind: Deployment
field: {}
`,
		dest: `
kind: Deployment
`,
		expected: `
kind: Deployment
field: {}
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},
}
