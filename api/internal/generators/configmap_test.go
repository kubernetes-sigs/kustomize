// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	. "sigs.k8s.io/kustomize/api/internal/generators"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

var binaryHello = []byte{
	0xff, // non-utf8
	0x68, // h
	0x65, // e
	0x6c, // l
	0x6c, // l
	0x6f, // o
}

func manyHellos(count int) (result []byte) {
	for i := 0; i < count; i++ {
		result = append(result, binaryHello...)
	}
	return
}

func TestMakeConfigMap(t *testing.T) {
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
  DB_PASSWORD: qwerty
  DB_USERNAME: admin
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
				out: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fileConfigMap2
data:
  app-init.ini: |
    FOO=bar
    BAR=baz
binaryData:
  app.bin: |
    /2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbG
    xv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hl
    bGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv/2
    hlbGxv/2hlbGxv/2hlbGxv/2hlbGxv
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
		manyHellos(30))
	kvLdr := kv.NewLoader(
		loader.NewFileLoaderAtRoot(fSys),
		valtest_test.MakeFakeValidator())

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := MakeConfigMap(kvLdr, &tc.args)
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
