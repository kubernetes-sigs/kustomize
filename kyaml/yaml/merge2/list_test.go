// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},
	{description: `merge k8s deployment containers - prepend`,
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
      - name: foo1
      - name: foo2
      - name: foo3
      - name: foo0
      `,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListPrepend,
		},
	},
	{description: `merge k8s deployment containers - append`,
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},
	{description: `merge k8s deployment volumes - append`,
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},
	{description: `merge k8s deployment volumes - prepend`,
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
      - name: foo1
      - name: foo2
      - name: foo3
      - name: foo0
`,
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListPrepend,
		},
	},
	{description: `merge k8s deployment containers -- $patch directive`,
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
          - $patch: merge
`,
		dest: `
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
          - name: foo4
          - name: foo5
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
          - name: foo3
          - name: foo4
          - name: foo5
`,
	},
	{description: `replace k8s deployment containers -- $patch directive`,
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
          - $patch: replace
`,
		dest: `
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
          - name: foo4
          - name: foo5
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
          - name: foo3
`,
	},
	{description: `remove k8s deployment containers -- $patch directive`,
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
          - $patch: delete
`,
		dest: `
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
          - name: foo4
          - name: foo5
`,
		expected: `
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec: {}
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
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
		mergeOptions: yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		},
	},
}
