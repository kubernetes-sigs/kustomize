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

func TestMakeSecret(t *testing.T) {
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
				out: `apiVersion: v1
kind: Secret
metadata:
  name: envSecret
type: Opaque
data:
  DB_PASSWORD: cXdlcnR5
  DB_USERNAME: YWRtaW4=
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
				out: `apiVersion: v1
kind: Secret
metadata:
  name: fileSecret1
type: Opaque
data:
  app-init.ini: Rk9PPWJhcgpCQVI9YmF6Cg==
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
				out: `apiVersion: v1
kind: Secret
metadata:
  name: fileSecret2
type: Opaque
data:
  app-init.ini: Rk9PPWJhcgpCQVI9YmF6Cg==
  app.bin: //0=
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
				out: `apiVersion: v1
kind: Secret
metadata:
  name: literalSecret1
  labels:
    foo: 'bar'
type: Opaque
data:
  a: eA==
  b: eQ==
  c: SGVsbG8gV29ybGQ=
  d: dHJ1ZQ==
`,
			},
		},
		"construct secret with type": {
			args: types.SecretArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "literalSecret1",
					KvPairSources: types.KvPairSources{
						LiteralSources: []string{"a=x"},
					},
					Options: &types.GeneratorOptions{
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
				Type: "foobar",
			},
			exp: expected{
				out: `apiVersion: v1
kind: Secret
metadata:
  name: literalSecret1
  labels:
    foo: 'bar'
type: foobar
data:
  a: eA==
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
type: Opaque
data:
  a: eA==
  b: eQ==
  c: SGVsbG8gV29ybGQ=
  d: dHJ1ZQ==
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
			rn, err := MakeSecret(kvLdr, &tc.args)
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
