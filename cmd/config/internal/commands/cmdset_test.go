// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestSetCommand(t *testing.T) {
	var tests = []struct {
		name              string
		inputOpenAPI      string
		input             string
		args              []string
		out               string
		expectedOpenAPI   string
		expectedResources string
		errMsg            string
	}{
		{
			name: "set replicas",
			args: []string{"replicas", "4", "--description", "hi there", "--set-by", "pw"},
			out:  "set 1 field(s)\n",
			inputOpenAPI: `
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
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hi there
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
          setBy: pw
          isSet: true
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4 # {"$openapi":"replicas"}
 `,
		},
		{
			name: "validate length of argument",
			args: []string{"--description", "hi there", "--set-by", "pw"},
			inputOpenAPI: `
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
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
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
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			errMsg: "requires at least 2 arg(s), only received 1",
		},

		{
			name: "set replicas no description",
			args: []string{"replicas", "4"},
			out:  "set 1 field(s)\n",
			inputOpenAPI: `
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
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
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
          value: "4"
          isSet: true
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},

		{
			name: "set image with value",
			args: []string{"tag", "1.8.1"},
			out:  "set 1 field(s)\n",
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image-setter:
      x-k8s-cli:
        setter:
          name: image-setter
          value: "nginx"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image-setter'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
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
        image: nginx:1.7.9 # {"$openapi":"image"}
      - name: sidecar
        image: sidecar:1.7.9
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image-setter:
      x-k8s-cli:
        setter:
          name: image-setter
          value: "nginx"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "1.8.1"
          isSet: true
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image-setter'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
			expectedResources: `
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
        image: nginx:1.8.1 # {"$openapi":"image"}
      - name: sidecar
        image: sidecar:1.7.9
`,
		},

		{
			name: "validate openAPI number",
			args: []string{"replicas", "four"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      type: number
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      type: number
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
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			errMsg: "replicas in body must be of type number",
		},

		{
			name: "validate openAPI string maxLength",
			args: []string{"name", "wordpress"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      type: string
      maxLength: 5
      description: hello world
      x-k8s-cli:
        setter:
          name: name
          value: nginx
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx # {"$ref":"#/definitions/io.k8s.cli.setters.name"}
spec:
  replicas: 3
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      type: string
      maxLength: 5
      description: hello world
      x-k8s-cli:
        setter:
          name: name
          value: nginx
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx # {"$ref":"#/definitions/io.k8s.cli.setters.name"}
spec:
  replicas: 3
 `,
			errMsg: "name in body should be at most 5 chars long",
		},

		{
			name: "validate substitution",
			args: []string{"tag", "1.8.1"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
    io.k8s.cli.setters.tag:
      type: string
      minLength: 6
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
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
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
    io.k8s.cli.setters.tag:
      type: string
      minLength: 6
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
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
			expectedResources: `
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
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
`,
			errMsg: "tag in body should be at least 6 chars long",
		},

		{
			name: "validate openAPI list values",
			args: []string{"list", "10", "hi", "true"},
			inputOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			input: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			expectedOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			errMsg: `list in body must be of type integer: "string"
list in body must be of type integer: "boolean"
list in body should have at most 2 items`,
		},

		{
			name: "set replicas with value set by flag",
			args: []string{"replicas", "--values", "4", "--description", "hi there"},
			out:  "set 1 field(s)\n",
			inputOpenAPI: `
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
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hi there
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
          isSet: true
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},

		{
			name: "validate values set from either flag or arg",
			args: []string{"replicas", "4", "--values", "4", "--description", "hi there"},
			inputOpenAPI: `
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
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
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
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			errMsg: `value should set either from flag or arg`,
		},

		{
			name: "openAPI list values set by flag success",
			args: []string{"list", "--values", "10", "--values", "11"},
			out:  "set 1 field(s)\n",
			inputOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			input: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			expectedOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - "10"
          - "11"
          isSet: true
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - "10"
  - "11"
 `,
		},

		{
			name: "validate openAPI list values set by flag error",
			args: []string{"list", "--values", "10", "--values", "hi", "--values", "true"},
			inputOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			input: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			expectedOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			errMsg: `list in body must be of type integer: "string"
list in body must be of type integer: "boolean"
list in body should have at most 2 items`,
		},
		{
			name: "nested substitution",
			args: []string{"my-image-setter", "ubuntu"},
			out:  "set 2 field(s)\n",
			inputOpenAPI: `
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
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
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
 `,
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
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image-setter:
      x-k8s-cli:
        setter:
          name: my-image-setter
          value: ubuntu
          isSet: true
    io.k8s.cli.setters.my-tag-setter:
      x-k8s-cli:
        setter:
          name: my-tag-setter
          value: 1.7.9
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
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
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
 `,
			expectedResources: `
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
        image: something/ubuntu::1.7.9/nginxotherthing # {"$openapi":"my-nested-subst"}
      - name: sidecar
        image: ubuntu::1.7.9 # {"$openapi":"my-image-subst"}
`,
		},
		{
			name: "nested cyclic substitution",
			args: []string{"my-image-setter", "ubuntu"},
			inputOpenAPI: `
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
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: ${my-nested-subst}::${my-tag-setter}
          values:
          - marker: ${my-nested-subst}
            ref: '#/definitions/io.k8s.cli.substitutions.my-nested-subst'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
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
 `,
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
			expectedOpenAPI: `
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
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: ${my-nested-subst}::${my-tag-setter}
          values:
          - marker: ${my-nested-subst}
            ref: '#/definitions/io.k8s.cli.substitutions.my-nested-subst'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
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
 `,
			expectedResources: `
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
			errMsg: "cyclic substitution detected with name my-nested-subst",
		},

		{
			name: "set v1 setter asm",
			args: []string{"profilesetter", "my-asm"},
			out:  "set 1 fields\n",
			inputOpenAPI: `
 `,
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub:
  - --gcr.io/asm-testing
  - --gcr.io/asm-testing2
 `,
			expectedOpenAPI: `
 `,
			expectedResources: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: my-asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"my-asm"}}}
  hub:
  - --gcr.io/asm-testing
  - --gcr.io/asm-testing2
 `,
		},

		{
			name: "set v1 partial setter",
			args: []string{"gcloud.core.project", "my-project"},
			out:  "set 1 fields\n",
			inputOpenAPI: `
 `,
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"gcloud.core.project","value":"project-id"},{"name":"cluster-name","value":"cluster-name"},{"name":"gcloud.compute.zone","value":"us-east1-d"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub:
  - --gcr.io/asm-testing
  - --gcr.io/asm-testing2
 `,
			expectedOpenAPI: `
 `,
			expectedResources: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "my-project/us-east1-d/cluster-name" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"gcloud.core.project","value":"my-project"},{"name":"cluster-name","value":"cluster-name"},{"name":"gcloud.compute.zone","value":"us-east1-d"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub:
  - --gcr.io/asm-testing
  - --gcr.io/asm-testing2
 `,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()

			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.RemoveAll(baseDir)

			f := filepath.Join(baseDir, "Krmfile")
			err = ioutil.WriteFile(f, []byte(test.inputOpenAPI), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			r, err := ioutil.TempFile(baseDir, "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(r.Name())
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			runner := commands.NewSetRunner("")
			out := &bytes.Buffer{}
			runner.Command.SetOut(out)
			runner.Command.SetArgs(append([]string{baseDir}, test.args...))
			err = runner.Command.Execute()
			if test.errMsg != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errMsg) {
					t.FailNow()
				}
			}

			if test.errMsg == "" && !assert.NoError(t, err) {
				t.FailNow()
			}

			if test.errMsg == "" && !assert.Contains(t, out.String(), strings.TrimSpace(test.out)) {
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

			actualOpenAPI, err := ioutil.ReadFile(f)
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

func TestSetSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "set-recurse-subpackages",
			dataset: "dataset-with-setters",
			args:    []string{"namespace", "otherspace", "-R"},
			expected: `${baseDir}/mysql/
set 1 field(s) of setter "namespace" to value "otherspace"

${baseDir}/mysql/nosetters/
setter "namespace" is not found

${baseDir}/mysql/storage/
set 1 field(s) of setter "namespace" to value "otherspace"
`,
		},
		{
			name:        "set-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-with-setters",
			packagePath: "mysql",
			args:        []string{"namespace", "otherspace"},
			expected: `${baseDir}/mysql/
set 1 field(s) of setter "namespace" to value "otherspace"
`,
		},
		{
			name:        "set-nested-pkg-no-recurse-subpackages",
			dataset:     "dataset-with-setters",
			packagePath: "mysql/storage",
			args:        []string{"namespace", "otherspace"},
			expected: `${baseDir}/mysql/storage/
set 1 field(s) of setter "namespace" to value "otherspace"
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()
			sourceDir := filepath.Join("test", "testdata", test.dataset)
			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			copyutil.CopyDir(sourceDir, baseDir)
			defer os.RemoveAll(baseDir)
			runner := commands.NewSetRunner("")
			actual := &bytes.Buffer{}
			runner.Command.SetOut(actual)
			runner.Command.SetArgs(append([]string{filepath.Join(baseDir, test.packagePath)}, test.args...))
			err = runner.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// normalize path format for windows
			actualNormalized := strings.Replace(
				strings.Replace(actual.String(), "\\", "/", -1),
				"//", "/", -1)

			expected := strings.Replace(test.expected, "${baseDir}", baseDir, -1)
			expectedNormalized := strings.Replace(
				strings.Replace(expected, "\\", "/", -1),
				"//", "/", -1)
			if !assert.Equal(t, strings.TrimSpace(expectedNormalized), strings.TrimSpace(actualNormalized)) {
				t.FailNow()
			}
		})
	}
}
