// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var scalarTestCases = []testCase{
	// Scalar Field Test Cases
	//
	// Test Case
	//
	{
		description: `Set and updated a field`,
		origin:      `kind: Deployment`,
		update:      `kind: StatefulSet`,
		local:       `kind: Deployment`,
		expected:    `kind: StatefulSet`},

	{
		description: `Add an updated field`,
		origin: `
apiVersion: apps/v1
kind: Deployment # old value`,
		update: `
apiVersion: apps/v1
kind: StatefulSet # new value`,
		local: `
apiVersion: apps/v1`,
		expected: `
apiVersion: apps/v1
kind: StatefulSet # new value`},

	//
	// Test Case
	//
	{
		description: `Ensure comments are added`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'dest'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'updated'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas"}`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'original'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'updated'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas"}
`},

	//
	// Test Case
	//
	{
		description: `Ensure comments are updated`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'dest'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas"}`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'updated'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas_new"}`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'original'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas"}
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'updated'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas_new"}
`},

	{
		description: `Ensure deleted comments are not updated`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'dest'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas"}`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'updated'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 3 # {"$openapi":"replicas"}`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'original'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 4
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/merge-source: 'updated'
    config.kubernetes.io/path: 'temp.yaml'
spec:
  replicas: 4
`},

	{
		description: `Add keep an omitted field`,
		origin: `
apiVersion: apps/v1
kind: Deployment`,
		update: `
apiVersion: apps/v1
kind: StatefulSet`,
		local: `
apiVersion: apps/v1
spec: foo # field not present in source
`,
		expected: `
apiVersion: apps/v1
spec: foo # field not present in source
kind: StatefulSet
`},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{
		description: `Change an updated field`,
		origin: `
apiVersion: apps/v1
kind: Deployment # old value`,
		update: `
apiVersion: apps/v1
kind: StatefulSet # new value`,
		local: `
apiVersion: apps/v1
kind: Service # conflicting value`,
		expected: `
apiVersion: apps/v1
kind: StatefulSet # new value`},

	{
		description: `Ignore a field`,
		origin: `
apiVersion: apps/v1
kind: Deployment # ignore this field`,
		update: `
apiVersion: apps/v1
kind: Deployment # ignore this field`,
		local: `
apiVersion: apps/v1`,
		expected: `
apiVersion: apps/v1`},

	{
		description: `Explicitly clear a field`,
		origin: `
apiVersion: apps/v1`,
		update: `
apiVersion: apps/v1
kind: null # clear this value`,
		local: `
apiVersion: apps/v1
kind: Deployment # value to be cleared`,
		expected: `
apiVersion: apps/v1`},

	{description: `Implicitly clear a field`,
		origin: `
apiVersion: apps/v1
kind: Deployment # clear this field`,
		update: `
apiVersion: apps/v1`,
		local: `
apiVersion: apps/v1
kind: Deployment # clear this field`,
		expected: `
apiVersion: apps/v1`},

	//
	// Test Case
	//
	// TODO(#36): consider making this an error
	{
		description: `Implicitly clear a changed field`,
		origin: `
apiVersion: apps/v1
kind: Deployment`,
		update: `
apiVersion: apps/v1`,
		local: `
apiVersion: apps/v1
kind: StatefulSet`,
		expected: `
apiVersion: apps/v1`},

	//
	// Test Case
	//
	{
		description: `Merge an empty scalar value`,
		origin: `
apiVersion: apps/v1
`,
		update: `
apiVersion: apps/v1
kind: {}
`,
		local: `
apiVersion: apps/v1
`,
		expected: `
apiVersion: apps/v1
kind: {}
`},
}
