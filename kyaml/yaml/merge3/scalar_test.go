// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var scalarTestCases = []testCase{
	// Scalar Field Test Cases
	//
	// Test Case
	//
	{`Set and updated a field`,
		`kind: Deployment`,
		`kind: StatefulSet`,
		`kind: Deployment`,
		`kind: StatefulSet`, nil},

	{`Add an updated field`,
		`
apiVersion: apps/v1
kind: Deployment # old value`,
		`
apiVersion: apps/v1
kind: StatefulSet # new value`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
kind: StatefulSet # new value`, nil},

	{`Add keep an omitted field`,
		`
apiVersion: apps/v1
kind: Deployment`,
		`
apiVersion: apps/v1
kind: StatefulSet`,
		`
apiVersion: apps/v1
spec: foo # field not present in source
`,
		`
apiVersion: apps/v1
spec: foo # field not present in source
kind: StatefulSet
`, nil},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{`Change an updated field`,
		`
apiVersion: apps/v1
kind: Deployment # old value`,
		`
apiVersion: apps/v1
kind: StatefulSet # new value`,
		`
apiVersion: apps/v1
kind: Service # conflicting value`,
		`
apiVersion: apps/v1
kind: StatefulSet # new value`, nil},

	{`Ignore a field`,
		`
apiVersion: apps/v1
kind: Deployment # ignore this field`,
		`
apiVersion: apps/v1
kind: Deployment # ignore this field`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1`, nil},

	{`Explicitly clear a field`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
kind: null # clear this value`,
		`
apiVersion: apps/v1
kind: Deployment # value to be cleared`,
		`
apiVersion: apps/v1`, nil},

	{`Implicitly clear a field`,
		`
apiVersion: apps/v1
kind: Deployment # clear this field`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
kind: Deployment # clear this field`,
		`
apiVersion: apps/v1`, nil},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{`Implicitly clear a changed field`,
		`
apiVersion: apps/v1
kind: Deployment`,
		`
apiVersion: apps/v1`,
		`
apiVersion: apps/v1
kind: StatefulSet`,
		`
apiVersion: apps/v1`, nil},

	//
	// Test Case
	//
	{`Merge an empty scalar value`,
		`
apiVersion: apps/v1
`,
		`
apiVersion: apps/v1
kind: {}
`,
		`
apiVersion: apps/v1
`,
		`
apiVersion: apps/v1
kind: {}
`, nil},
}
