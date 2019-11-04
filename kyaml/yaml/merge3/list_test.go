// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var listTestCases = []testCase{
	// List Field Test Cases

	//
	// Test Case
	//
	{`Replace list`,
		`
list:
- 1
- 2
- 3`,
		`
list:
- 2
- 3
- 4`,
		`
list:
- 1
- 2
- 3`,
		`
list:
- 2
- 3
- 4`, nil},

	//
	// Test Case
	//
	{`Add an updated list`,
		`
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3
`,
		`
apiVersion: apps/v1
list: # new value
- 2
- 3
- 4
`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
list:
- 2
- 3
- 4
`, nil},

	//
	// Test Case
	//
	{`Add keep an omitted field`,
		`
apiVersion: apps/v1
kind: Deployment`,
		`
apiVersion: apps/v1
kind: StatefulSet`,
		`
apiVersion: apps/v1
list: # not present in sources
- 2
- 3
- 4
`,
		`
apiVersion: apps/v1
list: # not present in sources
- 2
- 3
- 4
kind: StatefulSet
`, nil},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{`Change an updated field`,
		`
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3`,
		`
apiVersion: apps/v1
list: # new value
- 2
- 3
- 4`,
		`
apiVersion: apps/v1
list: # conflicting value
- a
- b
- c`,
		`
apiVersion: apps/v1
list: # conflicting value
- 2
- 3
- 4
`, nil},

	//
	// Test Case
	//
	{`Ignore a field -- set`,
		`
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3
`,
		`
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3`, `
apiVersion: apps/v1
list:
- 2
- 3
- 4
`, `
apiVersion: apps/v1
list:
- 2
- 3
- 4
`, nil},

	//
	// Test Case
	//
	{`Ignore a field -- empty`,
		`
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3`,
		`
apiVersion: apps/v1
list: # ignore value
- 1
- 2
- 3`,
		`
apiVersion: apps/v1
`,
		`
apiVersion: apps/v1
`, nil},

	//
	// Test Case
	//
	{`Explicitly clear a field`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
list: null # clear`,
		`
apiVersion: apps/v1
list: # value to clear
- 1
- 2
- 3`,
		`
apiVersion: apps/v1`, nil},

	//
	// Test Case
	//
	{`Implicitly clear a field`,
		`
apiVersion: apps/v1
list: # clear value
- 1
- 2
- 3`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3`,
		`
apiVersion: apps/v1`, nil},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{`Implicitly clear a changed field`,
		`
apiVersion: apps/v1
list: # old value
- 1
- 2
- 3`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
list: # old value
- a
- b
- c`,
		`
apiVersion: apps/v1`, nil},
}
