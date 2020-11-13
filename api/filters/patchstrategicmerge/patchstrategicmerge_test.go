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
      - name: foo0
      - name: foo1
      - name: foo2
      - name: foo3
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
      - name: foo0
      - name: foo1
      - name: foo2
      - name: foo3
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
		"remove mapping - directive": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
        $patch: delete   
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers: []
`,
		},
		"replace mapping - directive": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      $patch: replace
      containers:
      - name: new
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: new
`,
		},
		"merge mapping - directive": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test1
        $patch: merge   
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test1
`,
		},
		"remove list - directive": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - whatever
      - $patch: delete
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec: {}
`,
		},
		"replace list - directive": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: replace
        image: replace
      - $patch: replace   
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: replace
        image: replace
`,
		},
		"merge list - directive": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test
        image: test
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test2
        image: test2
      - $patch: merge   
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - name: test2
        image: test2
      - name: test
        image: test
`,
		},
		"list map keys - add a port, no names": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
         protocol: TCP
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
         protocol: UDP
       - containerPort: 80
         protocol: UDP
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
        - containerPort: 80
          protocol: UDP
        - containerPort: 8080
          protocol: TCP
`,
		},
		"list map keys - add name to port": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
         protocol: UDP
       - containerPort: 8080
         protocol: TCP
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
         protocol: UDP
         name: UDP-name-patch
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: UDP-name-patch
        - containerPort: 8080
          protocol: TCP
`,
		},
		"list map keys - replace port name": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
         protocol: UDP
         name: UDP-name-original
       - containerPort: 8080
         protocol: TCP
         name: TCP-name-original
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
         protocol: UDP
         name: UDP-name-patch
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: UDP-name-patch
        - containerPort: 8080
          protocol: TCP
          name: TCP-name-original
`,
		},
		"list map keys - add a port, no protocol": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 8080
`,
			patch: yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: test-deployment
spec:
 template:
   spec:
     containers:
     - image: test-image
       name: test-deployment
       ports:
       - containerPort: 80
`),
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
        - containerPort: 8080
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
