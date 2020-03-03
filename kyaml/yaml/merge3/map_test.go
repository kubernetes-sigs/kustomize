// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var mapTestCases = []testCase{
	//
	// Test Case
	//
	{description: `Add the annotations map field`,
		origin: `
kind: Deployment`,
		update: `
kind: Deployment
metadata:
  annotations:
    d: e # add these annotations
`,
		local: `
kind: Deployment`,
		expected: `
kind: Deployment
metadata:
  annotations:
    d: e # add these annotations`},

	//
	// Test Case
	//
	{description: `Add an annotation to the field`,
		origin: `
kind: Deployment
metadata:
  annotations:
    a: b`,
		update: `
kind: Deployment
metadata:
  annotations:
    a: b
    d: e  # add these annotations`,
		local: `
kind: Deployment
metadata:
  annotations:
    g: h  # keep these annotations`,
		expected: `
kind: Deployment
metadata:
  annotations:
    g: h # keep these annotations
    d: e # add these annotations`},

	//
	// Test Case
	//
	{description: `Add an annotation to the field, field missing from dest`,
		origin: `
kind: Deployment
metadata:
  annotations:
    a: b # ignored because unchanged`,
		update: `
kind: Deployment
metadata:
  annotations:
    a: b # ignore because unchanged
    d: e`,
		local: `
kind: Deployment`,
		expected: `
kind: Deployment
metadata:
  annotations:
    d: e`},

	//
	// Test Case
	//
	{description: `Update an annotation on the field, field messing rom the dest`,
		origin: `
kind: Deployment
metadata:
  annotations:
    a: b
    d: c`,
		update: `
kind: Deployment
metadata:
  annotations:
    a: b
    d: e  # set these annotations`,
		local: `
kind: Deployment
metadata:
  annotations:
    g: h  # keep these annotations`,
		expected: `
kind: Deployment
metadata:
  annotations:
    g: h # keep these annotations
    d: e # set these annotations`},

	//
	// Test Case
	//
	{description: `Add an annotation to the field, field missing from dest`,
		origin: `
kind: Deployment
metadata:
  annotations:
    a: b # ignored because unchanged`,
		update: `
kind: Deployment
metadata:
  annotations:
    a: b # ignore because unchanged
    d: e`,
		local: `
kind: Deployment`,
		expected: `
kind: Deployment
metadata:
  annotations:
    d: e`},

	//
	// Test Case
	//
	{description: `Remove an annotation`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    a: b`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations: {}`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    c: d
    a: b`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    c: d`},

	//
	// Test Case
	//
	// TODO(#36) support ~annotations~: {} deletion
	{description: `Specify a field as empty that isn't present in the source`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations: null`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    a: b`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo`},

	//
	// Test Case
	//
	{description: `Remove an annotation`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    a: b`,
		update: `
apiVersion: apps/v1
kind: Deployment`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    c: d
    a: b`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    c: d`},

	//
	// Test Case
	//
	{description: `Remove annotations field`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    a: b`,
		update: `
apiVersion: apps/v1
kind: Deployment`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`},

	//
	// Test Case
	//
	{description: `Remove annotations field, but keep in dest`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    a: b`,
		update: `
apiVersion: apps/v1
kind: Deployment`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    foo: bar # keep this annotation even though the parent field was removed`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    foo: bar # keep this annotation even though the parent field was removed`},

	//
	// Test Case
	//
	{description: `Remove annotations, but they are already empty`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    a: b
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations: {}
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations: {}
`},
}
