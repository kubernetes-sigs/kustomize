// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

var listTestCases = []testCase{
	{description: `strategic merge patch delete 1`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
        $patch: delete
      - name: foo2
      - name: foo3
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
      - name: foo2
      - name: foo3
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo2
      - name: foo3
`,
	},
	{description: `strategic merge patch delete 2`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
      - name: foo2
      - name: foo3
        $patch: delete
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
      - name: foo2
      - name: foo3
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
      - name: foo2
`,
	},
	{description: `merge k8s deployment containers`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
      - name: foo2
      - name: foo3
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo0
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo0
      - name: foo1
      - name: foo2
      - name: foo3
`,
	},
	{description: `merge k8s deployment volumes`,
		source: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      volumes:
      - name: foo1
      - name: foo2
      - name: foo3
`,
		dest: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      volumes:
      - name: foo0
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      volumes:
      - name: foo0
      - name: foo1
      - name: foo2
      - name: foo3
`,
	},

	{description: `replace List -- different value in dest`,
		source: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		dest: `
kind: Deployment
items:
- 0
- 1
`,
		expected: `
kind: Deployment
items:
- 1
- 2
- 3
`,
	},

	{description: `replace List -- missing from dest`,
		source: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		dest: `
kind: Deployment
`,
		expected: `
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
	{description: `keep List -- same value in src and dest`,
		source: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		dest: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		expected: `
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
	{description: `keep List -- unspecified in src`,
		source: `
kind: Deployment
`,
		dest: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		expected: `
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
	{description: `remove List -- null in src`,
		source: `
kind: Deployment
items: null
`,
		dest: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		expected: `
kind: Deployment
`,
	},

	//
	// Test Case
	//
	{description: `remove list -- empty in src`,
		source: `
kind: Deployment
items: []
`,
		dest: `
kind: Deployment
items:
- 1
- 2
- 3
`,
		expected: `
kind: Deployment
items: []
`,
	},
}
