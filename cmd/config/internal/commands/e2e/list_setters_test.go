// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestListSetters(t *testing.T) {
	tests := []test{
		{
			name: "set",
			args: []string{"cfg", "list-setters", "."},
			files: map[string]string{
				"deployment.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3  # {"$openapi":"replicas"}
`,
				"Krmfile": `
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
`,
			},
			expectedStdOut: `
./
    NAME     VALUE   SET BY   DESCRIPTION   COUNT   REQUIRED   IS SET  
  replicas   3                              1       No         No      
`,
		},
	}
	runTests(t, tests)
}
