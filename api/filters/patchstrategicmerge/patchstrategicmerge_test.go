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
//		"simple patch": {
//			input: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//`,
//			patch: yaml.MustParse(`
//metadata:
//  name: yourDeploy
//`),
//			expected: `
//apiVersion: apps/v1
//metadata:
//  name: yourDeploy
//kind: Deployment
//`,
//		},
//		"container patch": {
//			input: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  template:
//    spec:
//      containers:
//      - name: foo1
//      - name: foo2
//      - name: foo3
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  template:
//    spec:
//      containers:
//      - name: foo0
//`),
//			expected: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  template:
//    spec:
//      containers:
//      - name: foo0
//      - name: foo1
//      - name: foo2
//      - name: foo3
//`,
//		},
//		"volumes patch": {
//			input: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  template:
//    spec:
//      volumes:
//      - name: foo1
//      - name: foo2
//      - name: foo3
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  template:
//    spec:
//      volumes:
//      - name: foo0
//`),
//			expected: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  template:
//    spec:
//      volumes:
//      - name: foo0
//      - name: foo1
//      - name: foo2
//      - name: foo3
//`,
//		},
//		"nested patch": {
//			input: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  containers:
//  - name: nginx
//    args:
//    - abc
//`,
//			patch: yaml.MustParse(`
//spec:
//  containers:
//  - name: nginx
//    args:
//    - def
//`),
//			expected: `
//apiVersion: apps/v1
//metadata:
//  name: myDeploy
//kind: Deployment
//spec:
//  containers:
//  - name: nginx
//    args:
//    - def
//`,
//		},
//		"remove mapping - directive": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//        $patch: delete
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers: []
//`,
//		},
//		"replace mapping - directive": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      $patch: replace
//      containers:
//      - name: new
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: new
//`,
//		},
//		"merge mapping - directive": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test1
//        $patch: merge
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test1
//`,
//		},
//		"remove list - directive": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - whatever
//      - $patch: delete
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec: {}
//`,
//		},
//		"replace list - directive": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: replace
//        image: replace
//      - $patch: replace
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: replace
//        image: replace
//`,
//		},
//		"merge list - directive": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test
//        image: test
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test2
//        image: test2
//      - $patch: merge
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: myDeploy
//spec:
//  template:
//    spec:
//      containers:
//      - name: test2
//        image: test2
//      - name: test
//        image: test
//`,
//		},
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
         name: potato
       - containerPort: 8090
         protocol: other
         name: pot
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
         protocol: UDP
         name: roger
       - containerPort: 8080
         protocol: UDP
         name: potato
`),
			expected: `

`,
		},
//		"list map keys - port aliasing": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          name: disappearing-act
//          protocol: TCP
//        - containerPort: 8080
//          name: take-over-the-world
//          protocol: TCP
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//  labels:
//    test-transformer: did-my-job
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//  labels:
//    test-transformer: did-my-job
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          name: take-over-the-world
//          protocol: TCP
//`,
//		},
//		"list map keys - add port name": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: test-merge-keys
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: test-merge-keys
//`,
//		},
//		"list map keys - add a port, both named": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: tcp-port
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: UDP
//          name: udp-port
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: tcp-port
//        - containerPort: 8080
//          protocol: UDP
//          name: udp-port
//`,
//		},
//		"list map keys - change port name": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: original-port-name
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: patch-port-name
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          protocol: TCP
//          name: patch-port-name
//`,
//		},
//		"list map keys - change port name, no protocol": {
//			input: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          name: original-port-name
//`,
//			patch: yaml.MustParse(`
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          name: patch-port-name
//`),
//			expected: `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
//  name: test-deployment
//spec:
//  template:
//    spec:
//      containers:
//      - image: test-image
//        name: test-deployment
//        ports:
//        - containerPort: 8080
//          name: patch-port-name
//`,
//		},
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
