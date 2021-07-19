// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmap_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	. "sigs.k8s.io/kustomize/api/resmap"
	resmaptest_test "sigs.k8s.io/kustomize/api/testutils/resmaptest"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFromFile(t *testing.T) {
	resourceStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply2
---
# some comment
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply2
  namespace: test
---
`
	expected := resmaptest_test.NewRmBuilder(t, rf).
		Add(map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "dply1",
			}}).
		Add(map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "dply2",
			}}).
		Add(map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "dply2",
				"namespace": "test",
			}}).ResMap()
	expYaml, err := expected.AsYaml()
	assert.NoError(t, err)

	fSys := filesys.MakeFsInMemory()
	assert.NoError(t, fSys.WriteFile("deployment.yaml", []byte(resourceStr)))

	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	assert.NoError(t, err)

	m, err := rmF.FromFile(ldr, "deployment.yaml")
	assert.NoError(t, err)
	mYaml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, expYaml, mYaml)
}

func TestFromBytes(t *testing.T) {
	encoded := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`)
	expected := resmaptest_test.NewRmBuilder(t, rf).
		Add(map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm1",
			}}).
		Add(map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "cm2",
			}}).ResMap()
	expYaml, err := expected.AsYaml()
	assert.NoError(t, err)
	m, err := rmF.NewResMapFromBytes(encoded)
	assert.NoError(t, err)
	mYaml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, expYaml, mYaml)
}

func TestNewFromConfigMaps(t *testing.T) {
	type testCase struct {
		description string
		input       []types.ConfigMapArgs
		filepath    string
		content     string
		expected    ResMap
	}

	fSys := filesys.MakeFsInMemory()
	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}
	kvLdr := kv.NewLoader(ldr, valtest_test.MakeFakeValidator())
	testCases := []testCase{
		{
			description: "construct config map from env",
			input: []types.ConfigMapArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Name: "envConfigMap",
						KvPairSources: types.KvPairSources{
							EnvSources: []string{"app.env"},
						},
					},
				},
			},
			filepath: "app.env",
			content:  "DB_USERNAME=admin\nDB_PASSWORD=somepw",
			expected: resmaptest_test.NewRmBuilder(t, rf).Add(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "envConfigMap",
					},
					"data": map[string]interface{}{
						"DB_USERNAME": "admin",
						"DB_PASSWORD": "somepw",
					}}).ResMap(),
		},

		{
			description: "construct config map from file",
			input: []types.ConfigMapArgs{{
				GeneratorArgs: types.GeneratorArgs{
					Name: "fileConfigMap",
					KvPairSources: types.KvPairSources{
						FileSources: []string{"app-init.ini"},
					},
				},
			},
			},
			filepath: "app-init.ini",
			content:  "FOO=bar\nBAR=baz\n",
			expected: resmaptest_test.NewRmBuilder(t, rf).Add(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "fileConfigMap",
					},
					"data": map[string]interface{}{
						"app-init.ini": `FOO=bar
BAR=baz
`,
					},
				}).ResMap(),
		},
		{
			description: "construct config map from literal",
			input: []types.ConfigMapArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Name: "literalConfigMap",
						KvPairSources: types.KvPairSources{
							LiteralSources: []string{"a=x", "b=y", "c=\"Good Morning\"", "d=\"false\""},
						},
					},
				},
			},
			expected: resmaptest_test.NewRmBuilder(t, rf).Add(
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "literalConfigMap",
					},
					"data": map[string]interface{}{
						"a": "x",
						"b": "y",
						"c": "Good Morning",
						"d": "false",
					},
				}).ResMap(),
		},

		// TODO: add testcase for data coming from multiple sources like
		// files/literal/env etc.
	}
	for _, tc := range testCases {
		if tc.filepath != "" {
			if fErr := fSys.WriteFile(tc.filepath, []byte(tc.content)); fErr != nil {
				t.Fatalf("error adding file '%s': %v\n", tc.filepath, fErr)
			}
		}
		r, err := rmF.NewResMapFromConfigMapArgs(kvLdr, tc.input)
		assert.NoError(t, err, tc.description)
		r.RemoveBuildAnnotations()
		rYaml, err := r.AsYaml()
		assert.NoError(t, err, tc.description)
		tc.expected.RemoveBuildAnnotations()
		expYaml, err := tc.expected.AsYaml()
		assert.NoError(t, err, tc.description)
		assert.Equal(t, expYaml, rYaml)
	}
}

func TestNewResMapFromSecretArgs(t *testing.T) {
	secrets := []types.SecretArgs{
		{
			GeneratorArgs: types.GeneratorArgs{
				Name: "apple",
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{
						"DB_USERNAME=admin",
						"DB_PASSWORD=somepw",
					},
				},
			},
			Type: ifc.SecretTypeOpaque,
		},
	}
	fSys := filesys.MakeFsInMemory()
	fSys.Mkdir(filesys.SelfDir)

	actual, err := rmF.NewResMapFromSecretArgs(
		kv.NewLoader(
			loader.NewFileLoaderAtRoot(fSys),
			valtest_test.MakeFakeValidator()), secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actual.RemoveBuildAnnotations()
	actYaml, err := actual.AsYaml()
	assert.NoError(t, err)

	expected := resmaptest_test.NewRmBuilder(t, rf).Add(
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name": "apple",
			},
			"type": ifc.SecretTypeOpaque,
			"data": map[string]interface{}{
				"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
				"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
			},
		}).ResMap()
	expYaml, err := expected.AsYaml()
	assert.NoError(t, err)

	assert.Equal(t, string(expYaml), string(actYaml))
}

func TestFromRNodeSlice(t *testing.T) {
	type testcase struct {
		input    string
		expected ResMap
	}
	testcases := map[string]testcase{
		"no resource": {
			input:    "---",
			expected: resmaptest_test.NewRmBuilder(t, rf).ResMap(),
		},
		"single resource": {
			input: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: namespace-reader
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - watch
  - list
      `,
			expected: resmaptest_test.NewRmBuilder(t, rf).Add(
				map[string]interface{}{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRole",
					"metadata": map[string]interface{}{
						"name": "namespace-reader",
					},
					"rules": []interface{}{
						map[string]interface{}{
							"apiGroups": []interface{}{
								"",
							},
							"resources": []interface{}{
								"namespaces",
							},
							"verbs": []interface{}{
								"get",
								"watch",
								"list",
							},
						},
					},
				}).ResMap(),
		},
		"local config": {
			// local config should be ignored
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.kubernetes.io/local-config: 'true'
`,
			expected: resmaptest_test.NewRmBuilder(t, rf).ResMap(),
		},
	}
	for name := range testcases {
		tc := testcases[name]
		t.Run(name, func(t *testing.T) {
			rm, err := rmF.NewResMapFromRNodeSlice(
				[]*yaml.RNode{yaml.MustParse(tc.input)})
			if err != nil {
				t.Fatalf("unexpected error in test case [%s]: %v", name, err)
			}
			if err = tc.expected.ErrorIfNotEqualLists(rm); err != nil {
				t.Fatalf("error in test case [%s]: %s", name, err)
			}
		})
	}
}
