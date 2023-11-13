// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestByteWriter(t *testing.T) {
	type testCase struct {
		name           string
		err            string
		items          []string
		functionConfig string
		results        string
		expectedOutput string
		instance       ByteWriter
	}

	testCases := []testCase{
		//
		// Test Case
		//
		{
			name: "wrap_resource_list",
			instance: ByteWriter{
				Sort:               true,
				WrappingKind:       ResourceListKind,
				WrappingAPIVersion: ResourceListAPIVersion,
			},
			items: []string{
				`a: b #first`,
				`c: d # second`,
			},
			functionConfig: `
e: f
g:
  h:
  - i # has a list
  - j`,
			expectedOutput: `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- a: b #first
- c: d # second
functionConfig:
  e: f
  g:
    h:
    - i # has a list
    - j
`,
		},

		//
		// Test Case
		//
		{
			name: "multiple_items",
			items: []string{
				`c: d # second`,
				`e: f
g:
  h:
  # has a list
  - i : [i1, i2] # line comment
  # has a list 2
  - j : j1
`,
				`a: b #first`,
			},
			expectedOutput: `
c: d # second
---
e: f
g:
  h:
  # has a list
  - i: [i1, i2] # line comment
  # has a list 2
  - j: j1
---
a: b #first
`,
		},

		//
		// Test Case
		//
		{
			name: "handle_comments",
			items: []string{
				`# comment 0
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
  namespace: my-space
  labels:
    env: dev
    foo: bar
spec:
  # comment 1
  replicas: 3
  selector:
    # comment 2
    matchLabels: # comment 3
      # comment 4
      app: nginx # comment 5
  template:
    metadata:
      labels:
        app: nginx
    spec:
      # comment 6
      containers:
        # comment 7
        - name: nginx
          image: nginx:1.14.2 # comment 8
          ports:
            # comment 9
            - containerPort: 80 # comment 10
`,
				`apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  ports:
    # comment 1
    - name: etcd-server-ssl
      port: 2380
    # comment 2
    - name: etcd-client-ssl
      port: 2379
`,
				`apiVersion: constraints.gatekeeper.sh/v1beta1
kind: EnforceFoo
metadata:
  name: enforce-foo
spec:
  parameters:
    naming_rules:
      - kind: Folder
        patterns:
          # comment 1
          - ^(dev|prod|staging|qa|shared)$
`,
			},
			expectedOutput: `# comment 0
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
  namespace: my-space
  labels:
    env: dev
    foo: bar
spec:
  # comment 1
  replicas: 3
  selector:
    # comment 2
    matchLabels: # comment 3
      # comment 4
      app: nginx # comment 5
  template:
    metadata:
      labels:
        app: nginx
    spec:
      # comment 6
      containers:
      # comment 7
      - name: nginx
        image: nginx:1.14.2 # comment 8
        ports:
        # comment 9
        - containerPort: 80 # comment 10
---
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  ports:
  # comment 1
  - name: etcd-server-ssl
    port: 2380
  # comment 2
  - name: etcd-client-ssl
    port: 2379
---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: EnforceFoo
metadata:
  name: enforce-foo
spec:
  parameters:
    naming_rules:
    - kind: Folder
      patterns:
      # comment 1
      - ^(dev|prod|staging|qa|shared)$
`,
		},

		//
		// Test Case
		//
		{
			name:     "sort_keep_annotation",
			instance: ByteWriter{Sort: true, KeepReaderAnnotations: true},
			items: []string{
				`a: b #first
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
`,
				`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/b_test.yaml"
`,
				`c: d # second
metadata:
  annotations:
    internal.config.kubernetes.io/index: 1
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
`,
			},

			expectedOutput: `a: b #first
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    internal.config.kubernetes.io/index: 1
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
---
e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/b_test.yaml"
`,
		},

		//
		// Test Case
		//
		{
			name:     "sort_partial_annotations",
			instance: ByteWriter{Sort: true},
			items: []string{
				`a: b #first
metadata:
  annotations:
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
`,
				`c: d # second
metadata:
  annotations:
    internal.config.kubernetes.io/index: 1
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
`,
				`e: f
g:
  h:
  - i # has a list
  - j
`,
			},

			expectedOutput: `e: f
g:
  h:
  - i # has a list
  - j
---
a: b #first
metadata:
  annotations:
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
---
c: d # second
metadata:
  annotations:
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
`,
		},

		//
		// Test Case
		//
		{
			name:     "keep_annotation_seqindent",
			instance: ByteWriter{KeepReaderAnnotations: true},
			items: []string{
				`a: b #first
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
    internal.config.kubernetes.io/index: "compact"
`,
				`e: f
g:
  h:
  - i # has a list
  - j
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/b_test.yaml"
    internal.config.kubernetes.io/seqindent: "wide"
`,
				`c: d # second
metadata:
  annotations:
    internal.config.kubernetes.io/index: 1
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
    internal.config.kubernetes.io/seqindent: "compact"
`,
			},

			expectedOutput: `a: b #first
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
    internal.config.kubernetes.io/index: "compact"
---
e: f
g:
  h:
    - i # has a list
    - j
metadata:
  annotations:
    internal.config.kubernetes.io/index: 0
    internal.config.kubernetes.io/path: "a/b/b_test.yaml"
    internal.config.kubernetes.io/seqindent: "wide"
---
c: d # second
metadata:
  annotations:
    internal.config.kubernetes.io/index: 1
    internal.config.kubernetes.io/path: "a/b/a_test.yaml"
    internal.config.kubernetes.io/seqindent: "compact"
`,
		},

		//
		// Test Case
		//
		{
			name: "encode_valid_json",
			items: []string{
				`{
  "a": "a long string that would certainly see a newline introduced by the YAML marshaller abcd123",
  metadata: {
    annotations: {
      internal.config.kubernetes.io/path: test.json
    }
  }
}`,
			},

			expectedOutput: `{
  "a": "a long string that would certainly see a newline introduced by the YAML marshaller abcd123",
  "metadata": {
    "annotations": {
      "internal.config.kubernetes.io/path": "test.json"
    }
  }
}`,
		},

		//
		// Test Case
		//
		{
			name: "encode_valid_json_remove_seqindent_annotation",
			items: []string{
				`{
  "a": "a long string that would certainly see a newline introduced by the YAML marshaller abcd123",
  metadata: {
    annotations: {
      "internal.config.kubernetes.io/seqindent": "compact",
      "internal.config.kubernetes.io/index": "0",
      "internal.config.kubernetes.io/path": "test.json"
    }
  }
}`,
			},

			expectedOutput: `{
  "a": "a long string that would certainly see a newline introduced by the YAML marshaller abcd123",
  "metadata": {
    "annotations": {
      "internal.config.kubernetes.io/path": "test.json"
    }
  }
}`,
		},

		//
		// Test Case
		//
		{
			name: "encode_unformatted_valid_json",
			items: []string{
				`{ "a": "b", metadata: { annotations: { internal.config.kubernetes.io/path: test.json } } }`,
			},

			expectedOutput: `{
  "a": "b",
  "metadata": {
    "annotations": {
      "internal.config.kubernetes.io/path": "test.json"
    }
  }
}`,
		},

		//
		// Test Case
		//
		{
			name: "encode_wrapped_json_as_yaml",
			instance: ByteWriter{
				Sort:               true,
				WrappingKind:       ResourceListKind,
				WrappingAPIVersion: ResourceListAPIVersion,
			},
			items: []string{
				`{
  "a": "b",
  "metadata": {
    "annotations": {
      "internal.config.kubernetes.io/path": "test.json"
    }
  }
}`,
			},

			expectedOutput: `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- {"a": "b", "metadata": {"annotations": {"internal.config.kubernetes.io/path": "test.json"}}}
`,
		},

		//
		// Test Case
		//
		{
			name: "encode_multi_doc_json_as_yaml",
			items: []string{
				`{
  "a": "b",
  "metadata": {
    "annotations": {
      "internal.config.kubernetes.io/path": "test-1.json"
    }
  }
}`,
				`{
  "c": "d",
  "metadata": {
    "annotations": {
      "internal.config.kubernetes.io/path": "test-2.json"
    }
  }
}`,
			},

			expectedOutput: `
{"a": "b", "metadata": {"annotations": {"internal.config.kubernetes.io/path": "test-1.json"}}}
---
{"c": "d", "metadata": {"annotations": {"internal.config.kubernetes.io/path": "test-2.json"}}}
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			actual := &bytes.Buffer{}
			w := tc.instance
			w.Writer = actual

			if tc.functionConfig != "" {
				w.FunctionConfig = yaml.MustParse(tc.functionConfig)
			}

			if tc.results != "" {
				w.Results = yaml.MustParse(tc.results)
			}

			var items []*yaml.RNode
			for i := range tc.items {
				items = append(items, yaml.MustParse(tc.items[i]))
			}
			err := w.Write(items)

			if tc.err != "" {
				if !assert.EqualError(t, err, tc.err) {
					t.FailNow()
				}
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(actual.String())) {
				t.FailNow()
			}
		})
	}
}
