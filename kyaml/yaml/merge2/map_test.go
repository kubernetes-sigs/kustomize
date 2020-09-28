// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var mapTestCases = []testCase{

	{description: `strategic merge patch delete 1`,
		source: `
kind: Deployment
$patch: delete
`,
		dest: `
kind: Deployment
spec:
  foo: bar1
`,
		expected: ``,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	{description: `strategic merge patch delete 2`,
		source: `
kind: Deployment
spec:
  $patch: delete
`,
		dest: `
kind: Deployment
spec:
  foo: bar
  color: red
`,
		expected: `
kind: Deployment
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	{description: `strategic merge patch delete 3`,
		source: `
kind: Deployment
spec:
  metadata:
    name: wut
  template:
    $patch: delete
`,
		dest: `
kind: Deployment
spec:
  metadata:
    name: wut
  template:
    spec:
      containers:
      - name: foo
      - name: bar
`,
		expected: `
kind: Deployment
spec:
  metadata:
    name: wut
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	{description: `strategic merge patch delete 4`,
		source: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
$patch: delete
`,
		dest: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`,
		expected: `
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

	{description: `strategic merge patch replace 1`,
		source: `
kind: Deployment
spec:
  metal: heavy
  $patch: replace
  veggie: carrot
`,
		dest: `
kind: Deployment
spec:
  river: nile
  color: red
`,
		expected: `
kind: Deployment
spec:
  metal: heavy
  veggie: carrot
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},

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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
  baz: buz
  foo: bar1
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},
}
