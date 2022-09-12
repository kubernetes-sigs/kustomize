// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestFix(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, []byte(`nameprefix: some-prefix-`))

	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.RunE(cmd, nil))

	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "apiVersion: ")
	assert.Contains(t, string(content), "kind: Kustomization")
}

func TestFixCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		files    map[string]string
		expected string
	}{
		{
			name: "FixOutdatedPatchesFieldTitle",
			input: `
patchesJson6902:
- path: patch1.yaml
  target:
    kind: Service
- path: patch2.yaml
  target:
    group: apps
    kind: Deployment
    version: v1
`,
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: patch1.yaml
  target:
    kind: Service
- path: patch2.yaml
  target:
    group: apps
    kind: Deployment
    version: v1
`,
		},
		{
			name: "TestRenameAndKeepOutdatedPatchesField",
			input: `
patchesJson6902:
- path: patch1.yaml
  target:
    kind: Deployment
patches:
- path: patch2.yaml
  target:
    kind: Deployment
- path: patch3.yaml
  target:
    kind: Service
`,
			expected: `
patches:
- path: patch2.yaml
  target:
    kind: Deployment
- path: patch3.yaml
  target:
    kind: Service
- path: patch1.yaml
  target:
    kind: Deployment
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		{
			name: "TestFixOutdatedPatchesStrategicMergeFieldTitle",
			input: `
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: nignx:latest
`,
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx
    spec:
      template:
        spec:
          containers:
            - name: nginx
              image: nignx:latest
`,
		},
		{
			name: "TestFixAndMergeOutdatedPatchesStrategicMergeFieldTitle",
			input: `
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            image: nignx:latest
patches:
- path: patch2.yaml
  target:
    kind: Deployment
`,
			expected: `
patches:
- path: patch2.yaml
  target:
    kind: Deployment
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx
    spec:
      template:
        spec:
          containers:
            - name: nginx
              image: nignx:latest
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
		{
			name: "TestFixOutdatedPatchesStrategicMergeToPathFieldTitle",
			input: `
patchesStrategicMerge:
- deploy.yaml
`,
			files: map[string]string{
				"deploy.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
        - name: nginx
          env:
            - name: CONFIG_FILE_PATH
              value: home.yaml`,
			},
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: deploy.yaml
`,
		},
		{
			name: "TestFixOutdatedPatchesStrategicMergeToPathFieldYMLTitle",
			input: `
patchesStrategicMerge:
- deploy.yml
`,
			files: map[string]string{
				"deploy.yml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: CONFIG_FILE_PATH
          value: home.yaml`,
			},
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: deploy.yml
`,
		},
		{
			name: "Test fix outdated patchesStrategicMerge from a file and one string literal",
			input: `
patchesStrategicMerge:
- deploy.yaml
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            env:
              - name: CONFIG_FILE_PATH
                value: home.yaml
`,
			files: map[string]string{
				"deploy.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: CONFIG_FILE_PATH
          value: home.yaml`,
			},
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: deploy.yaml
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx
    spec:
      template:
        spec:
          containers:
            - name: nginx
              env:
                - name: CONFIG_FILE_PATH
                  value: home.yaml
`,
		},
		{
			name: "Test fix outdated patchesStrategicMerge and patchesJson6902",
			input: `
patchesStrategicMerge:
- deploy.yaml
patchesJson6902:
- path: patch1.yaml
  target:
    kind: Deployment
`,
			files: map[string]string{
				"deploy.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: CONFIG_FILE_PATH
          value: home.yaml`,
			},
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- path: patch1.yaml
  target:
    kind: Deployment
- path: deploy.yaml
`,
		},
		{
			name: "TestFixOutdatedPatchesStrategicMergeFieldPatchEndOfYamlTitle",
			input: `
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
          - name: nginx
            env:
              - name: CONFIG_FILE_PATH
                value: home.yaml
`,
			expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
patches:
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx
    spec:
      template:
        spec:
          containers:
            - name: nginx
              env:
                - name: CONFIG_FILE_PATH
                  value: home.yaml
`,
		},
		{
			name: "TestFixOutdatedCommonLabels",
			input: `
commonLabels:
  foo: bar
labels:
- pairs:
    a: b
`,
			expected: `
labels:
- pairs:
    a: b
- includeSelectors: true
  pairs:
    foo: bar
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			testutils_test.WriteTestKustomizationWith(fSys, []byte(test.input))
			for filename, content := range test.files {
				require.NoError(t, fSys.WriteFile(filename, []byte(content)))
			}
			cmd := NewCmdFix(fSys, os.Stdout)
			require.NoError(t, cmd.RunE(cmd, nil))

			content, err := testutils_test.ReadTestKustomization(fSys)
			require.NoError(t, err)
			require.Contains(t, string(content), "apiVersion: ")
			require.Contains(t, string(content), "kind: Kustomization")

			require.Empty(t, cmp.Diff([]byte(test.expected), content))
		})
	}
}

func TestFixOutdatedCommonLabelsDuplicate(t *testing.T) {
	kustomizationContentWithOutdatedCommonLabels := []byte(`
commonLabels:
  foo: bar
labels:
- pairs:
    foo: baz
    a: b
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomizationContentWithOutdatedCommonLabels)
	cmd := NewCmdFix(fSys, os.Stdout)
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "label name 'foo' exists in both commonLabels and labels")
}
