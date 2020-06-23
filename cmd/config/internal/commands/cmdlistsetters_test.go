// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestListSettersCommand(t *testing.T) {
	var tests = []struct {
		name     string
		openapi  string
		input    string
		args     []string
		expected string
	}{
		{
			name: "list-replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
          required: true
      description: "hello world"
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expected: `    NAME     VALUE   SET BY   DESCRIPTION   COUNT   REQUIRED  
  replicas   3       me       hello world   1       Yes       
`,
		},
		{
			name: "list-multiple",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: "hello world 1"
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me1
    io.k8s.cli.setters.image:
      description: "hello world 2"
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
          setBy: me2
    io.k8s.cli.setters.tag:
      description: "hello world 3"
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
          setBy: me3
          required: true
          isSet: false
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx # {"$ref": "#/definitions/io.k8s.cli.setters.image"}
 `,
			expected: `    NAME     VALUE   SET BY    DESCRIPTION    COUNT   REQUIRED  
  image      nginx   me2      hello world 2   2       No        
  replicas   3       me1      hello world 1   1       No        
  tag        1.7.9   me3      hello world 3   1       Yes       
  SUBSTITUTION    PATTERN    REFERENCES   
  image          IMAGE:TAG   [image,tag]  
`,
		},
		{
			name: "list-multiple-resources",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: "hello world 1"
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me1
    io.k8s.cli.setters.image:
      description: "hello world 2"
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
          setBy: me2
    io.k8s.cli.setters.tag:
      description: "hello world 3"
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
          setBy: me3
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx # {"$ref": "#/definitions/io.k8s.cli.setters.image"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx
`,
			expected: `    NAME     VALUE   SET BY    DESCRIPTION    COUNT   REQUIRED  
  image      nginx   me2      hello world 2   3       No        
  replicas   3       me1      hello world 1   2       No        
  tag        1.7.9   me3      hello world 3   2       No        
  SUBSTITUTION    PATTERN    REFERENCES   
  image          IMAGE:TAG   [image,tag]  
`,
		},
		{
			name: "list-name",
			args: []string{"image"},
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: "hello world 1"
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me1
    io.k8s.cli.setters.image:
      description: "hello world 2"
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
          setBy: me2
          required: true
    io.k8s.cli.setters.tag:
      description: "hello world 3"
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
          setBy: me3
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx # {"$ref": "#/definitions/io.k8s.cli.setters.image"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx
`,
			expected: `  NAME    VALUE   SET BY    DESCRIPTION    COUNT   REQUIRED  
  image   nginx   me2      hello world 2   3       Yes       
  SUBSTITUTION    PATTERN    REFERENCES   
  image          IMAGE:TAG   [image,tag]  
`,
		},

		{
			name: "list array setter",
			openapi: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      items:
        type: string
      maxItems: 3
      type: array
      description: hello world
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - a
          - b
          - c
          setBy: me
          required: true
 `,
			input: `
apiVersion: example.com/v1beta1
kind: Example
metadata:
  annotations:
    foo: bar
spec:
  list: # {"$openapi":"list"}
  - "a"
  - "b"
  - "c"
`,
			expected: `  NAME    VALUE    SET BY   DESCRIPTION   COUNT   REQUIRED  
  list   [a,b,c]   me       hello world   1       Yes       
`,
		},

		{
			name: "nested substitution",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: something/nginx::1.7.9/nginxotherthing # {"$openapi":"my-nested-subst"}
      - name: sidecar
        image: nginx::1.7.9 # {"$openapi":"my-image-subst"}
 `,
			openapi: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image-setter:
      x-k8s-cli:
        setter:
          name: my-image-setter
          value: nginx
    io.k8s.cli.setters.my-tag-setter:
      x-k8s-cli:
        setter:
          name: my-tag-setter
          value: 1.7.9
          required: true
          isSet: true
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: ${my-image-setter}::${my-tag-setter}
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
    io.k8s.cli.substitutions.my-nested-subst:
      x-k8s-cli:
        substitution:
          name: my-nested-subst
          pattern: something/${my-image-subst}/${my-other-setter}
          values:
          - marker: ${my-image-subst}
            ref: '#/definitions/io.k8s.cli.substitutions.my-image-subst'
          - marker: ${my-other-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-other-setter'
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
 `,
			expected: `       NAME              VALUE        SET BY   DESCRIPTION   COUNT   REQUIRED  
  my-image-setter   nginx                                    2       No        
  my-other-setter   nginxotherthing                          1       No        
  my-tag-setter     1.7.9                                    2       Yes       
   SUBSTITUTION                        PATTERN                                  REFERENCES             
  my-image-subst    ${my-image-setter}::${my-tag-setter}             [my-image-setter,my-tag-setter]   
  my-nested-subst   something/${my-image-subst}/${my-other-setter}   [my-image-subst,my-other-setter]  
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()

			f, err := ioutil.TempFile("", "k8s-cli-")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(f.Name())
			old := ext.GetOpenAPIFile
			defer func() { ext.GetOpenAPIFile = old }()
			ext.GetOpenAPIFile = func(args []string) (s string, err error) {
				err = ioutil.WriteFile(f.Name(), []byte(test.openapi), 0600)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				return f.Name(), nil
			}

			r, err := ioutil.TempFile("", "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(r.Name())
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			runner := commands.NewListSettersRunner("")
			actual := &bytes.Buffer{}
			runner.Command.SetOut(actual)
			runner.Command.SetArgs(append([]string{r.Name()}, test.args...))
			err = runner.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t, test.expected, actual.String()) {
				t.FailNow()
			}
		})
	}
}
