// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fixsetters

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixSettersV1(t *testing.T) {
	var tests = []struct {
		name            string
		input           string
		err             string
		dryRun          bool
		openAPIFile     string
		expectedOutput  string
		expectedOpenAPI string
		needFix         bool
		createdSetters  []string
		createdSubst    []string
		failedSetters   map[string]error
		failedSubst     map[string]error
	}{
		{
			name: "upgrade-delete-partial-setters",
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  cluster: "someproj/someclus" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"project","value":"someproj"},{"name":"cluster","value":"someclus"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profile","value":"asm"}}}
  cluster: "someproj/someclus" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"project","value":"someproj"},{"name":"cluster","value":"someclus"}]}}
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization`,

			needFix:        true,
			createdSetters: []string{"cluster", "profile", "project"},
			createdSubst:   []string{"project-cluster-54235872"},
			failedSetters:  map[string]error{},
			failedSubst:    map[string]error{},

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  cluster: "someproj/someclus" # {"$openapi":"project-cluster-54235872"}
spec:
  profile: asm # {"$openapi":"profile"}
  cluster: "someproj/someclus" # {"$openapi":"project-cluster-54235872"}
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.cluster:
      type: string
      x-k8s-cli:
        setter:
          name: cluster
          value: someclus
    io.k8s.cli.setters.profile:
      type: string
      x-k8s-cli:
        setter:
          name: profile
          value: asm
    io.k8s.cli.setters.project:
      type: string
      x-k8s-cli:
        setter:
          name: project
          value: someproj
    io.k8s.cli.substitutions.project-cluster-54235872:
      x-k8s-cli:
        substitution:
          name: project-cluster-54235872
          pattern: ${project}/${cluster}
          values:
          - marker: ${project}
            ref: '#/definitions/io.k8s.cli.setters.project'
          - marker: ${cluster}
            ref: '#/definitions/io.k8s.cli.setters.cluster'
`,
		},

		{
			name:   "upgrade-delete-partial-setters-dryRun",
			dryRun: true,
			input: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  cluster: "someproj/someclus" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"project","value":"someproj"},{"name":"cluster","value":"someclus"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profile","value":"asm"}}}
  cluster: "someproj/someclus" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"project","value":"someproj"},{"name":"cluster","value":"someclus"}]}}
`,
			needFix:        true,
			createdSetters: []string{"cluster", "profile", "project"},
			createdSubst:   []string{"project-cluster-54235872"},
			failedSetters:  map[string]error{},
			failedSubst:    map[string]error{},

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
metadata:
  cluster: "someproj/someclus" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"project","value":"someproj"},{"name":"cluster","value":"someclus"}]}}
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profile","value":"asm"}}}
  cluster: "someproj/someclus" # {"type":"string","x-kustomize":{"partialSetters":[{"name":"project","value":"someproj"},{"name":"cluster","value":"someclus"}]}}
`,
		},

		{
			name: "partial-setters-same-value",
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
spec:
  profile: asm # {"type":"string","x-kustomize":{"setter":{"name":"profile","value":"asm"}}}
  team: asm # {"type":"string","x-kustomize":{"setter":{"name":"team","value":"asm"}}}
  profile-team: asm/asm # {"type":"string","x-kustomize":{"partialSetters":[{"name":"profile","value":"asm"},{"name":"team","value":"asm"}]}}
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization`,

			needFix:        true,
			createdSetters: []string{"profile", "team"},
			createdSubst:   []string{"profile-team-1851878264"},
			failedSetters:  map[string]error{},
			failedSubst:    map[string]error{},

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
spec:
  profile: asm # {"$openapi":"profile"}
  team: asm # {"$openapi":"team"}
  profile-team: asm/asm # {"$openapi":"profile-team-1851878264"}
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.profile:
      type: string
      x-k8s-cli:
        setter:
          name: profile
          value: asm
    io.k8s.cli.setters.team:
      type: string
      x-k8s-cli:
        setter:
          name: team
          value: asm
    io.k8s.cli.substitutions.profile-team-1851878264:
      x-k8s-cli:
        substitution:
          name: profile-team-1851878264
          pattern: ${profile}/${team}
          values:
          - marker: ${profile}
            ref: '#/definitions/io.k8s.cli.setters.profile'
          - marker: ${team}
            ref: '#/definitions/io.k8s.cli.setters.team'
`,
		},

		{
			name: "partial-setters-suffix-subst",
			input: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
spec:
  profile: asm-profile # {"type":"string","x-kustomize":{"partialSetters":[{"name":"asm","value":"asm"}]}}
  team: asm-team # {"type":"string","x-kustomize":{"partialSetters":[{"name":"asm","value":"asm"}]}}
 `,

			openAPIFile: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization`,

			needFix:        true,
			createdSetters: []string{"asm"},
			createdSubst:   []string{"asm-3472570278", "asm-3647054792"},
			failedSetters:  map[string]error{},
			failedSubst:    map[string]error{},

			expectedOutput: `apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
spec:
  profile: asm-profile # {"$openapi":"asm-3647054792"}
  team: asm-team # {"$openapi":"asm-3472570278"}
`,

			expectedOpenAPI: `apiVersion: kustomization.dev/v1alpha1
kind: Kustomization
openAPI:
  definitions:
    io.k8s.cli.setters.asm:
      type: string
      x-k8s-cli:
        setter:
          name: asm
          value: asm
    io.k8s.cli.substitutions.asm-3472570278:
      x-k8s-cli:
        substitution:
          name: asm-3472570278
          pattern: ${asm}-team
          values:
          - marker: ${asm}
            ref: '#/definitions/io.k8s.cli.setters.asm'
    io.k8s.cli.substitutions.asm-3647054792:
      x-k8s-cli:
        substitution:
          name: asm-3647054792
          pattern: ${asm}-profile
          values:
          - marker: ${asm}
            ref: '#/definitions/io.k8s.cli.setters.asm'
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

			needFix:        true,
			createdSetters: []string{"profilesetter"},
			failedSetters:  map[string]error{},
			failedSubst:    map[string]error{},

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

			needFix:        true,
			createdSetters: []string{"profilesetter"},
			failedSetters:  map[string]error{},
			failedSubst:    map[string]error{},
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
			failedSetters: map[string]error{},
			failedSubst:   map[string]error{},

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
			sf := SetterFixer{
				PkgPath:     dir,
				DryRun:      test.dryRun,
				OpenAPIPath: filepath.Join(dir, "Krmfile"),
			}

			sfr, err := sf.FixV1Setters()
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
			assert.Equal(t, test.expectedOutput, string(actualOutput))

			if test.expectedOpenAPI != "" {
				actualOpenAPI, err := ioutil.ReadFile(filepath.Join(dir, openAPIFileName))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				assert.Equal(t, test.expectedOpenAPI, string(actualOpenAPI))
			}
			assert.Equal(t, test.needFix, sfr.NeedFix)
			assert.Equal(t, test.createdSetters, sfr.CreatedSetters)
			assert.Equal(t, test.createdSubst, sfr.CreatedSubst)
			assert.Equal(t, test.failedSubst, sfr.FailedSubst)
			assert.Equal(t, test.failedSetters, sfr.FailedSetters)
		})
	}
}
