// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var elementTestCases = []testCase{
	{`merge Element -- keep field in dest`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v0
  command: ['run.sh']
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command:
  - run.sh
`,
	},

	{`merge Element -- add field to dest`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v0
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command:
  - run.sh
`,
	},

	{`merge Element -- add list, empty in dest`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
		`
kind: Deployment
items: {}
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command:
  - run.sh
`,
	},

	{`merge Element -- add list, missing from dest`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command: ['run.sh']
`,
		`
kind: Deployment
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
  command:
  - run.sh
`,
	},

	{`merge Element -- add Element first`,
		`
kind: Deployment
items:
- name: bar
  image: bar:v1
  command: ['run2.sh']
- name: foo
  image: foo:v1
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v0
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
	},

	{`merge Element -- add Element second`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v0
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
	},

	//
	// Test Case
	//
	{`keep list -- list missing from src`,
		`
kind: Deployment
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
	},

	//
	// Test Case
	//
	{`keep Element -- element missing in src`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v0
- name: bar
  image: bar:v1
  command: ['run2.sh']
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
	},

	//
	// Test Case
	//
	{`keep element -- empty list in src`,
		`
kind: Deployment
items: {}
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
	},

	//
	// Test Case
	//
	{`remove Element -- null in src`,
		`
kind: Deployment
items: null
`,
		`
kind: Deployment
items:
- name: foo
  image: foo:v1
- name: bar
  image: bar:v1
  command:
  - run2.sh
`,
		`
kind: Deployment
`,
	},
}
