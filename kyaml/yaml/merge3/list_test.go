// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var listTestCases = []testCase{
	// List Field Test Cases

	//
	// Test Case
	//
	{description: `Replace list`,
		origin: `
list:
- 1
- 2
- 3`,
		update: `
list:
- 2
- 3
- 4`,
		local: `
list:
- 1
- 2
- 3`,
		expected: `
list:
- 2
- 3
- 4`},

	//
	// Test Case
	//
	{description: `Add an updated list`,
		origin: `
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3
`,
		update: `
apiVersion: apps/v1
list: # new value
- 2
- 3
- 4
`,
		local: `
apiVersion: apps/v1`,
		expected: `
apiVersion: apps/v1
list:
- 2
- 3
- 4
`},

	//
	// Test Case
	//
	{description: `Add keep an omitted field`,
		origin: `
apiVersion: apps/v1
kind: Deployment`,
		update: `
apiVersion: apps/v1
kind: StatefulSet`,
		local: `
apiVersion: apps/v1
list: # not present in sources
- 2
- 3
- 4
`,
		expected: `
apiVersion: apps/v1
list: # not present in sources
- 2
- 3
- 4
kind: StatefulSet
`},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{description: `Change an updated field`,
		origin: `
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3`,
		update: `
apiVersion: apps/v1
list: # new value
- 2
- 3
- 4`,
		local: `
apiVersion: apps/v1
list: # conflicting value
- a
- b
- c`,
		expected: `
apiVersion: apps/v1
list: # conflicting value
- 2
- 3
- 4
`},

	//
	// Test Case
	//
	{description: `Ignore a field -- set`,
		origin: `
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3
`,
		update: `
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3`,
		local: `
apiVersion: apps/v1
list:
- 2
- 3
- 4
`,
		expected: `
apiVersion: apps/v1
list:
- 2
- 3
- 4
`},

	//
	// Test Case
	//
	{description: `Ignore a field -- empty`,
		origin: `
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3`,
		update: `
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3`,
		local: `
apiVersion: apps/v1
`,
		expected: `
apiVersion: apps/v1
`},

	//
	// Test Case
	//
	{description: `Explicitly clear a field`,
		origin: `
apiVersion: apps/v1`,
		update: `
apiVersion: apps/v1
list: null # clear`,
		local: `
apiVersion: apps/v1
list: # value to clear
- 1
- 2
- 3`,
		expected: `
apiVersion: apps/v1`},

	//
	// Test Case
	//
	{description: `Implicitly clear a field`,
		origin: `
apiVersion: apps/v1
list: # clear value
- 1
- 2
- 3`,
		update: `
apiVersion: apps/v1`,
		local: `
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3`,
		expected: `
apiVersion: apps/v1`},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{description: `Implicitly clear a changed field`,
		origin: `
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3`,
		update: `
apiVersion: apps/v1`,
		local: `
apiVersion: apps/v1
list: # old value
- a
- b
- c`,
		expected: `
apiVersion: apps/v1`},
}
