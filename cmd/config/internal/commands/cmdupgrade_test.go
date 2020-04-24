// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

func TestUpgradeCommand(t *testing.T) {
	var tests = []struct {
		name            string
		input           string
		err             string
		openAPIFile     string
		expectedOutput  string
		expectedOpenAPI string
	}{
		{
			name: "upgrade-delete-partial-setters",
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"gcloud.core.project","value":"project-id"},{"name":"cluster-name","value":"cluster-name"},{"name":"gcloud.compute.zone","value":"us-east1-d"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub:
  - --gcr.io/asm-testing # {"type":"string","x-kustomize":{"setter":{"name":"hubsetter","value":"--gcr.io/asm-testing"}}}
  - --gcr.io/asm-testing2
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization`,

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"$openapi":"profilesetter"}
  hub:
  - --gcr.io/asm-testing # {"$openapi":"hubsetter"}
  - --gcr.io/asm-testing2
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.cluster-name:
      type: string
      x-k8s-cli:
        setter:
          name: cluster-name
          value: cluster-name
    io.k8s.cli.setters.gcloud.compute.zone:
      type: string
      x-k8s-cli:
        setter:
          name: gcloud.compute.zone
          value: us-east1-d
    io.k8s.cli.setters.gcloud.core.project:
      type: string
      x-k8s-cli:
        setter:
          name: gcloud.core.project
          value: project-id
    io.k8s.cli.setters.hubsetter:
      type: string
      x-k8s-cli:
        setter:
          name: hubsetter
          value: --gcr.io/asm-testing
    io.k8s.cli.setters.profilesetter:
      type: string
      x-k8s-cli:
        setter:
          name: profilesetter
          value: asm
`,
		},
		{
			name: "upgrade-with-both-versions",

			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: gcr.io/asm-testing # {"$openapi":"hubsetter"}
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.hubsetter:
      type: string
      x-k8s-cli:
        setter:
          name: hubsetter
          value: gcr.io/asm-testing`,

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"$openapi":"profilesetter"}
  hub: gcr.io/asm-testing # {"$openapi":"hubsetter"}
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.hubsetter:
      type: string
      x-k8s-cli:
        setter:
          name: hubsetter
          value: gcr.io/asm-testing
    io.k8s.cli.setters.profilesetter:
      type: string
      x-k8s-cli:
        setter:
          name: profilesetter
          value: asm
`,
		},

		{
			name: "setter-already-exists",

			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: asm # {"$openapi":"profilesetter"}
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.profilesetter:
      type: string
      x-k8s-cli:
        setter:
          name: profilesetter
          value: asm
`,

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"$openapi":"profilesetter"}
  hub: asm # {"$openapi":"profilesetter"}
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.profilesetter:
      type: string
      x-k8s-cli:
        setter:
          name: profilesetter
          value: asm
`,
		},
		{
			name: "do-not-delete-latest setters",

			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"$openapi":"profilesetter"}
  hub: gcr.io/asm-testing
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.profilesetter:
      type: string
      x-k8s-cli:
        setter:
          name: profilesetter
          value: asm
`,

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"$openapi":"profilesetter"}
  hub: gcr.io/asm-testing
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.profilesetter:
      type: string
      x-k8s-cli:
        setter:
          name: profilesetter
          value: asm
`,
		},
		{
			name: "no-openAPI-file-error",
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  clusterName: "project-id/us-east1-d/cluster-name"
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profilesetter","value":"asm"}}}
  hub: gcr.io/asm-testing
 `,

			err: "Krmfile:",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			openAPIFileName := "Krmfile"

			dir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			defer os.RemoveAll(dir)

			err = ioutil.WriteFile(filepath.Join(dir, "deploy.yaml"), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if test.openAPIFile != "" {
				err = ioutil.WriteFile(filepath.Join(dir, openAPIFileName), []byte(test.openAPIFile), 0600)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			}

			runner := commands.NewUpgradeRunner("")
			runner.Command.SetArgs([]string{dir})
			err = runner.Command.Execute()
			if test.err == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			} else {
				if !assert.Contains(t, err.Error(), test.err) {
					t.FailNow()
				}
				return
			}

			actualOutput, err := ioutil.ReadFile(filepath.Join(dir, "deploy.yaml"))
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			actualOpenAPI, err := ioutil.ReadFile(filepath.Join(dir, openAPIFileName))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.Equal(t, test.expectedOutput, string(actualOutput))
			assert.Equal(t, test.expectedOpenAPI, string(actualOpenAPI))
		})
	}
}
