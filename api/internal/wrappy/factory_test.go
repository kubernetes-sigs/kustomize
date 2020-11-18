// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestMakeConfigMap(t *testing.T) {
	factory := &WNodeFactory{}
	type expected struct {
		out    string
		errMsg string
	}

	testCases := map[string]struct {
		args types.ConfigMapArgs
		exp  expected
	}{
		"construct config map from env": {
			args: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "envConfigMap",
					KvPairSources: types.KvPairSources{
						EnvSources: []string{
							filepath.Join("configmap", "app.env"),
						},
					},
				},
			},
			exp: expected{
				out: `apiVersion: v1
kind: ConfigMap
metadata:
  name: envConfigMap
data:
  DB_USERNAME: admin
  DB_PASSWORD: qwerty
`,
			},
		},
		"construct config map from text file": {
			args: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "fileConfigMap1",
					KvPairSources: types.KvPairSources{
						FileSources: []string{
							filepath.Join("configmap", "app-init.ini"),
						},
					},
				},
			},
			exp: expected{
				out: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fileConfigMap1
data:
  app-init.ini: |
    FOO=bar
    BAR=baz
`,
			},
		},
		"construct config map from text and binary file": {
			args: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "fileConfigMap2",
					KvPairSources: types.KvPairSources{
						FileSources: []string{
							filepath.Join("configmap", "app-init.ini"),
							filepath.Join("configmap", "app.bin"),
						},
					},
				},
			},
			exp: expected{
				errMsg: "configMap generate error: key 'app.bin' appears " +
					"to have non-utf8 data; binaryData field not yet supported",
				out: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fileConfigMap2
data:
  app-init.ini: |
    FOO=bar
    BAR=baz
`,
			},
		},
		"construct config map from literal": {
			args: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "literalConfigMap1",
					KvPairSources: types.KvPairSources{
						LiteralSources: []string{"a=x", "b=y", "c=\"Hello World\"", "d='true'"},
					},
					Options: &types.GeneratorOptions{
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			exp: expected{
				out: `apiVersion: v1
kind: ConfigMap
metadata:
  name: literalConfigMap1
  labels:
    foo: 'bar'
data:
  a: x
  b: y
  c: Hello World
  d: "true"
`,
			},
		},
		"construct config map from literal with GeneratorOptions in ConfigMapArgs": {
			args: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "literalConfigMap2",
					KvPairSources: types.KvPairSources{
						LiteralSources: []string{"a=x", "b=y", "c=\"Hello World\"", "d='true'"},
					},
					Options: &types.GeneratorOptions{
						Labels: map[string]string{
							"veggie": "celery",
							"dog":    "beagle",
							"cat":    "annoying",
						},
						Annotations: map[string]string{
							"river": "Missouri",
							"city":  "Iowa City",
						},
					},
				},
			},
			exp: expected{
				out: `apiVersion: v1
kind: ConfigMap
metadata:
  name: literalConfigMap2
  labels:
    cat: 'annoying'
    dog: 'beagle'
    veggie: 'celery'
  annotations:
    city: 'Iowa City'
    river: 'Missouri'
data:
  a: x
  b: y
  c: Hello World
  d: "true"
`,
			},
		},
	}
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(
		filesys.RootedPath("configmap", "app.env"),
		[]byte("DB_USERNAME=admin\nDB_PASSWORD=qwerty\n"))
	fSys.WriteFile(
		filesys.RootedPath("configmap", "app-init.ini"),
		[]byte("FOO=bar\nBAR=baz\n"))
	fSys.WriteFile(
		filesys.RootedPath("configmap", "app.bin"),
		[]byte{0xff, 0xfd})
	kvLdr := kv.NewLoader(
		loader.NewFileLoaderAtRoot(fSys),
		valtest_test.MakeFakeValidator())

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := factory.makeConfigMap(kvLdr, &tc.args)
			if err != nil {
				if !assert.EqualError(t, err, tc.exp.errMsg) {
					t.FailNow()
				}
				return
			}
			if tc.exp.errMsg != "" {
				t.Fatalf("%s: should return error '%s'", n, tc.exp.errMsg)
			}
			output := rn.MustString()
			if !assert.Equal(t, tc.exp.out, output) {
				t.FailNow()
			}
		})
	}
}

func TestMakeSecret(t *testing.T) {
	factory := &WNodeFactory{}
	type expected struct {
		out    string
		errMsg string
	}

	testCases := map[string]struct {
		args types.SecretArgs
		exp  expected
	}{
		"construct secret from env": {
			args: types.SecretArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "envSecret",
					KvPairSources: types.KvPairSources{
						EnvSources: []string{
							filepath.Join("secret", "app.env"),
						},
					},
				},
			},
			exp: expected{
				errMsg: "TODO(WNodeFactory): finish implementation of makeSecret",
				out: `apiVersion: v1
kind: Secret
metadata:
  name: envSecret
data:
  DB_USERNAME: admin
  DB_PASSWORD: qwerty
`,
			},
		},
		"construct secret from text file": {
			args: types.SecretArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "fileSecret1",
					KvPairSources: types.KvPairSources{
						FileSources: []string{
							filepath.Join("secret", "app-init.ini"),
						},
					},
				},
			},
			exp: expected{
				errMsg: "TODO(WNodeFactory): finish implementation of makeSecret",
				out: `apiVersion: v1
kind: Secret
metadata:
  name: fileSecret1
data:
  app-init.ini: |
    FOO=bar
    BAR=baz
`,
			},
		},
		"construct secret from text and binary file": {
			args: types.SecretArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "fileSecret2",
					KvPairSources: types.KvPairSources{
						FileSources: []string{
							filepath.Join("secret", "app-init.ini"),
							filepath.Join("secret", "app.bin"),
						},
					},
				},
			},
			exp: expected{
				errMsg: "TODO(WNodeFactory): finish implementation of makeSecret",
				out: `apiVersion: v1
kind: Secret
metadata:
  name: fileSecret2
data:
  app-init.ini: |
    FOO=bar
    BAR=baz
`,
			},
		},
		"construct secret from literal": {
			args: types.SecretArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "literalSecret1",
					KvPairSources: types.KvPairSources{
						LiteralSources: []string{"a=x", "b=y", "c=\"Hello World\"", "d='true'"},
					},
					Options: &types.GeneratorOptions{
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			exp: expected{
				errMsg: "TODO(WNodeFactory): finish implementation of makeSecret",
				out: `apiVersion: v1
kind: Secret
metadata:
  name: literalSecret1
  labels:
    foo: 'bar'
data:
  a: x
  b: y
  c: Hello World
  d: "true"
`,
			},
		},
		"construct secret from literal with GeneratorOptions in SecretArgs": {
			args: types.SecretArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "literalSecret2",
					KvPairSources: types.KvPairSources{
						LiteralSources: []string{"a=x", "b=y", "c=\"Hello World\"", "d='true'"},
					},
					Options: &types.GeneratorOptions{
						Labels: map[string]string{
							"veggie": "celery",
							"dog":    "beagle",
							"cat":    "annoying",
						},
						Annotations: map[string]string{
							"river": "Missouri",
							"city":  "Iowa City",
						},
					},
				},
			},
			exp: expected{
				errMsg: "TODO(WNodeFactory): finish implementation of makeSecret",
				out: `apiVersion: v1
kind: Secret
metadata:
  name: literalSecret2
  labels:
    cat: 'annoying'
    dog: 'beagle'
    veggie: 'celery'
  annotations:
    city: 'Iowa City'
    river: 'Missouri'
data:
  a: x
  b: y
  c: Hello World
  d: "true"
`,
			},
		},
	}
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile(
		filesys.RootedPath("secret", "app.env"),
		[]byte("DB_USERNAME=admin\nDB_PASSWORD=qwerty\n"))
	fSys.WriteFile(
		filesys.RootedPath("secret", "app-init.ini"),
		[]byte("FOO=bar\nBAR=baz\n"))
	fSys.WriteFile(
		filesys.RootedPath("secret", "app.bin"),
		[]byte{0xff, 0xfd})
	kvLdr := kv.NewLoader(
		loader.NewFileLoaderAtRoot(fSys),
		valtest_test.MakeFakeValidator())

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := factory.makeSecret(kvLdr, &tc.args)
			if err != nil {
				if !assert.EqualError(t, err, tc.exp.errMsg) {
					t.FailNow()
				}
				return
			}
			if tc.exp.errMsg != "" {
				t.Fatalf("%s: should return error '%s'", n, tc.exp.errMsg)
			}
			output := rn.MustString()
			if !assert.Equal(t, tc.exp.out, output) {
				t.FailNow()
			}
		})
	}
}

func TestSliceFromBytes(t *testing.T) {
	factory := &WNodeFactory{}
	testConfigMap :=
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "winnie",
			},
		}
	testConfigMapList :=
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMapList",
			"items": []interface{}{
				testConfigMap,
				testConfigMap,
			},
		}

	type expected struct {
		out   []map[string]interface{}
		isErr bool
	}

	testCases := map[string]struct {
		input []byte
		exp   expected
	}{
		"garbage": {
			input: []byte("garbageIn: garbageOut"),
			exp: expected{
				isErr: true,
			},
		},
		"noBytes": {
			input: []byte{},
			exp: expected{
				out: []map[string]interface{}{},
			},
		},
		"goodJson": {
			input: []byte(`
{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap},
			},
		},
		"goodYaml1": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap},
			},
		},
		"goodYaml2": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap, testConfigMap},
			},
		},
		"localConfigYaml": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie-skip
  annotations:
    # this annotation causes the Resource to be ignored by kustomize
    config.kubernetes.io/local-config: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMap},
			},
		},
		"garbageInOneOfTwoObjects": {
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
WOOOOOOOOOOOOOOOOOOOOOOOOT:  woot
`),
			exp: expected{
				isErr: true,
			},
		},
		"emptyObjects": {
			input: []byte(`
---
#a comment

---

`),
			exp: expected{
				out: []map[string]interface{}{},
			},
		},
		"Missing .metadata.name in object": {
			input: []byte(`
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    foo: bar
`),
			exp: expected{
				isErr: true,
			},
		},
		"nil value in list": {
			input: []byte(`
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: kube100-site
	labels:
	  app: web
testList:
- testA
-
`),
			exp: expected{
				isErr: true,
			},
		},
		"List": {
			input: []byte(`
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{
					testConfigMap,
					testConfigMap},
			},
		},
		"ConfigMapList": {
			input: []byte(`
apiVersion: v1
kind: ConfigMapList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
`),
			exp: expected{
				out: []map[string]interface{}{testConfigMapList},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rs, err := factory.SliceFromBytes(tc.input)
			if tc.exp.isErr && err == nil {
				t.Fatalf("%v: should return error", n)
			}
			if !tc.exp.isErr && err != nil {
				t.Fatalf("%v: unexpected error: %s", n, err)
			}
			if len(tc.exp.out) != len(rs) {
				fmt.Printf("%s: \nexpected:%v\nactual: %v\n",
					n, tc.exp.out, rs)
				t.Fatalf("%s: length mismatch; expected %d, actual %d",
					n, len(tc.exp.out), len(rs))
			}
			for i := range rs {
				if !reflect.DeepEqual(tc.exp.out[i], rs[i].Map()) {
					t.Fatalf("%s: Got: %v\nexpected:%v",
						n, rs[i].Map(), tc.exp.out[i])
				}
			}
		})
	}
}
