package patchstrategicmerge

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter(t *testing.T) {
	testCases := map[string]struct {
		input    string
		patch    *yaml.RNode
		expected string
	}{
		"simple patch": {
			input: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
`,
			patch: yaml.MustParse(`
metadata:
  name: yourDeploy
`),
			expected: `
apiVersion: apps/v1
metadata:
  name: yourDeploy
kind: Deployment
`,
		},
		"container patch": {
			input: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo1
      - name: foo2
      - name: foo3
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo0
`),
			expected: `
apiVersion: apps/v1
metadata:
  name: myDeploy
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
		},
		"volumes patch": {
			input: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      volumes:
      - name: foo1
      - name: foo2
      - name: foo3
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      volumes:
      - name: foo0
`),
			expected: `
apiVersion: apps/v1
metadata:
  name: myDeploy
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
		},
		"nested patch": {
			input: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  containers:
  - name: nginx
    args:
    - abc
`,
			patch: yaml.MustParse(`
spec:
  containers:
  - name: nginx
    args:
    - def
`),
			expected: `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  containers:
  - name: nginx
    args:
    - def
`,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			f := Filter{
				Patch: tc.patch,
			}
			if !assert.Equal(t,
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(
					filtertest.RunFilter(t, tc.input, f))) {
				t.FailNow()
			}
		})
	}
}
