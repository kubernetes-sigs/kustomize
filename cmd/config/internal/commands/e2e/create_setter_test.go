// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"
)

func TestCreateSetter(t *testing.T) {
	tests := []test{
		{
			name: "create_setter",
			args: []string{"cfg", "create-setter", ".", "replicas", "3"},
			files: map[string]string{
				"deployment.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
`,
				"Krmfile": `
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
`,
			},
			expectedFiles: map[string]string{
				"deployment.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3 # {"$openapi":"replicas"}
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
		},
	}
	runTests(t, tests)
}
