// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestCreateSetterCommand(t *testing.T) {
	var tests = []struct {
		name              string
		input             string
		args              []string
		schema            string
		out               string
		inputOpenAPI      string
		expectedOpenAPI   string
		expectedResources string
		err               string
	}{
		{
			name: "add replicas",
			args: []string{"replicas", "3", "--description", "hello world", "--set-by", "me"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
 `,
		},

		{
			name: "add replicas no match",
			args: []string{"replicas", "3", "--description", "hello world", "--set-by", "me"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  foo: 2
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  foo: 2
 `,
		},
		{
			name: "error if substitution with same name exists",
			args: []string{"my-image", "3", "--description", "hello world", "--set-by", "me"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.substitutions.my-image:
      x-k8s-cli:
        substitution:
          name: my-image
          pattern: something/${my-image-setter}::${my-tag-setter}/nginxotherthing
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
 `,
			err: "substitution with name my-image already exists, substitution and setter can't have same name",
		},

		{
			name:   "add replicas with schema",
			args:   []string{"replicas", "3", "--description", "hello world", "--set-by", "me"},
			schema: `{"maximum": 10, "type": "integer"}`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      maximum: 10
      type: integer
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
 `,
		},

		{
			name: "add replicas not enough arguments",
			args: []string{"replicas", "--description", "hello world", "--set-by", "me"},
			err:  `setter name and value must be provided, value can either be an argument or can be passed as a flag --value`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		},

		{
			name:   "list values with schema",
			args:   []string{"list", "--description", "hello world", "--set-by", "me", "--type", "array", "--field", "spec.list"},
			schema: `{"maxItems": 3, "type": "array", "items": {"type": "string"}}`,
			input: `
apiVersion: example.com/v1beta1
kind: Example1
spec:
  list:
  - "a"
  - "b"
  - "c"
---
apiVersion: example.com/v1beta1
kind: Example2
spec:
  list:
  - "a"
  - "b"
  - "c"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: myspace
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: sidecar
        image: nginx:1.7.9
      - name: nginx
        image: otherspace/nginx:1.7.9
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
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
          value: ""
          listValues:
          - a
          - b
          - c
          setBy: me
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example1
spec:
  list: # {"$openapi":"list"}
  - "a"
  - "b"
  - "c"
---
apiVersion: example.com/v1beta1
kind: Example2
spec:
  list: # {"$openapi":"list"}
  - "a"
  - "b"
  - "c"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: myspace
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: sidecar
        image: nginx:1.7.9
      - name: nginx
        image: otherspace/nginx:1.7.9
 `,
		},

		{
			name:   "error list path with different values",
			args:   []string{"list", "--description", "hello world", "--set-by", "me", "--type", "array", "--field", "spec.list"},
			schema: `{"maxItems": 3, "type": "array", "items": {"type": "string"}}`,
			input: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list:
  - "a"
  - "b"
  - "c"
---
apiVersion: example.com/v1beta1
kind: Example
spec:
  list:
  - "c"
  - "d"
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			err: `setters can only be created for fields with same values, encountered different ` +
				`array values for specified field path: [c d], [a b c]`,
		},

		{
			name:   "list values error if field not set",
			args:   []string{"list", "a", "--description", "hello world", "--set-by", "me", "--type", "array"},
			schema: `{"maxItems": 3, "type": "array", "items": {"type": "string"}}`,
			input: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list:
  - "a"
  - "b"
  - "c"
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
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
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$openapi":"list"}
  - "a"
  - "b"
  - "c"
 `,
			err: `field flag must be set for array type setters`,
		},
		{
			name: "add replicas with value set by flag",
			args: []string{"replicas", "--value", "3", "--description", "hello world", "--set-by", "me"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
 `,
		},
		{
			name: "add setter with . in the name",
			args: []string{"foo.bar", "3"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.foo.bar:
      x-k8s-cli:
        setter:
          name: foo.bar
          value: "3"
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"foo.bar"}
 `,
		},

		{
			name: "create setter v1",
			args: []string{"hubsetter", "my-hub"},
			inputOpenAPI: `
 `,
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: my-hub
 `,
			expectedOpenAPI: `
 `,
			expectedResources: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: my-hub # {"type":"","x-kustomize":{"setter":{"name":"hubsetter","value":"my-hub"}}}
 `,
		},

		{
			name: "create partial setter v1",
			args: []string{"regionsetter", "us-east1-d", "--partial"},
			inputOpenAPI: `
 `,
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name" # {"type":"","x-kustomize":{"partialSetters":[{"name":"projectsetter","value":"project-id"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: my-hub # {"type":"","x-kustomize":{"setter":{"name":"hubsetter","value":"my-hub"}}}
 `,
			expectedOpenAPI: `
 `,
			expectedResources: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name" # {"type":"","x-kustomize":{"partialSetters":[{"name":"projectsetter","value":"project-id"},{"name":"regionsetter","value":"us-east1-d"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: my-hub # {"type":"","x-kustomize":{"setter":{"name":"hubsetter","value":"my-hub"}}}
 `,
		},
		{
			name:   "provide different type values in schema and with flag",
			args:   []string{"replicas", "3", "--description", "hello world", "--type", "string"},
			schema: `{"type": "integer"}`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			err: `type provided in type flag (string) and in schema (integer) doesn't match`,
		},
		{
			name:   "invalid json in schema",
			args:   []string{"replicas", "3"},
			schema: `{"foo": bar`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			err: "unable to parse schema: invalid character 'b' looking for beginning of value",
		},
		{
			name:   "unknown fields in schema are dropped",
			args:   []string{"replicas", "3"},
			schema: `{"foo": "bar"}`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
 `,
		},
		{
			name:   "unknown types are not accepted",
			args:   []string{"replicas", "3"},
			schema: `{"type": "int"}`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			err: `invalid schema: type "int" is not supported. Must be one of: object, array, string, integer, number, boolean, file, null`,
		},
		{
			name:   "unknown types are not accepted in nested structures",
			args:   []string{"replicas", "3", "--field", "replicas"},
			schema: `{"maxItems": 3, "type": "array", "items": {"type": "foo"}}`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			err: `invalid schema: type "foo" is not supported. Must be one of: object, array, string, integer, number, boolean, file, null`,
		},
		{
			name: "unknown types are not accepted in --type flag",
			args: []string{"replicas", "3", "--type", "bar"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			err: `invalid schema: type "bar" is not supported. Must be one of: object, array, string, integer, number, boolean, file, null`,
		},
		{
			name:   "unknown properties in schema are dropped",
			args:   []string{"replicas", "3", "--type", "integer"},
			schema: `{"maximum": 3, "unknown": 42}`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      maximum: 3
      type: integer
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
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

			err = ioutil.WriteFile(f.Name(), []byte(test.inputOpenAPI), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if test.schema != "" {
				sch, err := ioutil.TempFile("", "schema.json")
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				defer os.Remove(sch.Name())

				err = ioutil.WriteFile(sch.Name(), []byte(test.schema), 0600)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				test.args = append(test.args, "--schema-path", sch.Name())
			}

			old := ext.GetOpenAPIFile
			defer func() { ext.GetOpenAPIFile = old }()
			ext.GetOpenAPIFile = func(args []string) (s string, err error) {
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

			runner := commands.NewCreateSetterRunner("")
			out := &bytes.Buffer{}
			runner.Command.SetOut(out)
			runner.Command.SetArgs(append([]string{r.Name()}, test.args...))
			err = runner.Command.Execute()
			if test.err != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				} else {
					assert.Equal(t, err.Error(), test.err)
					return
				}
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t, test.out, out.String()) {
				t.FailNow()
			}

			actualResources, err := ioutil.ReadFile(r.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(test.expectedResources),
				strings.TrimSpace(string(actualResources))) {
				t.FailNow()
			}

			actualOpenAPI, err := ioutil.ReadFile(f.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(test.expectedOpenAPI),
				strings.TrimSpace(string(actualOpenAPI))) {
				t.FailNow()
			}
		})
	}
}
