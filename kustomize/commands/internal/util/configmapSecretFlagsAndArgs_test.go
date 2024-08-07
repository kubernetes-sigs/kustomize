// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestValidateAdd(t *testing.T) {
	fa := ConfigMapSecretFlagsAndArgs{}

	testCases := []struct {
		name    string
		args    []string
		wantErr func(require.TestingT, error, ...interface{})
	}{
		{
			"validateAdd with no arguments",
			[]string{},
			require.Error,
		},
		{
			"validateAdd with more than one name",
			[]string{"name", "othername"},
			require.Error,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.wantErr(t, fa.ValidateAdd(tc.args))
		})
	}
}

func TestValidateSet(t *testing.T) {
	testCases := []struct {
		name    string
		args    []string
		fa      ConfigMapSecretFlagsAndArgs
		wantErr func(require.TestingT, error, ...interface{})
	}{
		{
			// must have one single name
			name:    "fails with no arguments",
			args:    []string{},
			fa:      ConfigMapSecretFlagsAndArgs{},
			wantErr: require.Error,
		},
		{
			// must have one single name
			name:    "fails with more than one name",
			args:    []string{"testname", "testname2"},
			fa:      ConfigMapSecretFlagsAndArgs{},
			wantErr: require.Error,
		},
		{
			// must have at least --from-literal or --new-namespace
			name: "succeeds with --from-literal",
			args: []string{"test-configmap"},
			fa: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{"key1=value1"},
			},
			wantErr: require.NoError,
		},
		{
			// must have at least --from-literal or --new-namespace
			name: "succeeds with --new-namespace",
			args: []string{"test-configmap"},
			fa: ConfigMapSecretFlagsAndArgs{
				NewNamespace: "new-namespace",
			},
			wantErr: require.NoError,
		},
		{
			// must have at least --from-literal or --new-namespace
			name: "succeeds with --new-namespace and --from-literal",
			args: []string{"test-configmap"},
			fa: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{"key1=value1"},
				NewNamespace:   "new-namespace",
			},
			wantErr: require.NoError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.wantErr(t, tc.fa.ValidateSet(tc.args))
		})
	}
}

func TestDataConfigValidation_Flags(t *testing.T) {
	tests := []struct {
		name    string
		fa      ConfigMapSecretFlagsAndArgs
		wantErr func(require.TestingT, error, ...interface{})
	}{
		{
			name: "env-file-source and literal are both set",
			fa: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{"one", "two"},
				EnvFileSource:  "three",
			},
			wantErr: require.Error,
		},
		{
			name: "env-file-source and from-file are both set",
			fa: ConfigMapSecretFlagsAndArgs{
				FileSources:   []string{"one", "two"},
				EnvFileSource: "three",
			},
			wantErr: require.Error,
		},
		{
			name:    "we don't have any option set",
			fa:      ConfigMapSecretFlagsAndArgs{},
			wantErr: require.Error,
		},
		{
			name: "we have from-file and literal ",
			fa: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{"one", "two"},
				FileSources:    []string{"three", "four"},
			},
			wantErr: require.NoError,
		},
		{
			name: "correct behavior",
			fa: ConfigMapSecretFlagsAndArgs{
				EnvFileSource: "foo",
				Behavior:      "merge",
			},
			wantErr: require.NoError,
		},
		{
			name: "incorrect behavior",
			fa: ConfigMapSecretFlagsAndArgs{
				EnvFileSource: "foo",
				Behavior:      "merge-unknown",
			},
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.wantErr(t, test.fa.ValidateAdd([]string{"name"}))
		})
	}
}

func TestExpandFileSource(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	_, err := fSys.Create("dir/fa1")
	require.NoError(t, err)
	_, err = fSys.Create("dir/fa2")
	require.NoError(t, err)
	_, err = fSys.Create("dir/readme")
	require.NoError(t, err)
	fa := ConfigMapSecretFlagsAndArgs{
		FileSources: []string{"dir/fa*"},
	}
	err = fa.ExpandFileSource(fSys)
	require.NoError(t, err)
	expected := []string{
		"dir/fa1",
		"dir/fa2",
	}
	if !reflect.DeepEqual(fa.FileSources, expected) {
		t.Fatalf("FileSources is not correctly expanded: %v", fa.FileSources)
	}
}

func TestExpandFileSourceWithKey(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	_, err := fSys.Create("dir/faaaaaaaaaabbbbbbbbbccccccccccccccccc")
	require.NoError(t, err)
	_, err = fSys.Create("dir/foobar")
	require.NoError(t, err)
	_, err = fSys.Create("dir/simplebar")
	require.NoError(t, err)
	_, err = fSys.Create("dir/readme")
	require.NoError(t, err)
	fa := ConfigMapSecretFlagsAndArgs{
		FileSources: []string{"foo-key=dir/fa*", "bar-key=dir/foobar", "dir/simplebar"},
	}
	err = fa.ExpandFileSource(fSys)
	require.NoError(t, err)
	expected := []string{
		"foo-key=dir/faaaaaaaaaabbbbbbbbbccccccccccccccccc",
		"bar-key=dir/foobar",
		"dir/simplebar",
	}
	if !reflect.DeepEqual(fa.FileSources, expected) {
		t.Fatalf("FileSources is not correctly expanded: %v", fa.FileSources)
	}
}

func TestExpandFileSourceWithKeyAndError(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	_, err := fSys.Create("dir/fa1")
	require.NoError(t, err)
	_, err = fSys.Create("dir/fa2")
	require.NoError(t, err)
	_, err = fSys.Create("dir/readme")
	require.NoError(t, err)
	fa := ConfigMapSecretFlagsAndArgs{
		FileSources: []string{"foo-key=dir/fa*"},
	}
	err = fa.ExpandFileSource(fSys)
	if err == nil {
		t.Fatalf("FileSources should not be correctly expanded: %v", fa.FileSources)
	}
}

func TestUpdateLiteralSources(t *testing.T) {
	testCases := []struct {
		name         string
		args         *types.GeneratorArgs
		flags        ConfigMapSecretFlagsAndArgs
		expectedArgs *types.GeneratorArgs
		wantErr      func(require.TestingT, error, ...interface{})
	}{
		{
			name: "fails when key doesn't exist",
			args: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key1=val1", "otherkey=value"},
				},
			},
			flags: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{
					"key2=value",
				},
			},
			expectedArgs: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key1=val1", "otherkey=value"},
				},
			},
			wantErr: require.Error,
		},
		{
			name: "updates correctly an existing key",
			args: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1", "otherkey=value"},
				},
			},
			flags: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{
					"key2=value",
				},
			},
			expectedArgs: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=value", "otherkey=value"},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "fails when format for literal sources is incorrect in flags",
			args: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1", "otherkey=value"},
				},
			},
			flags: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{
					"key2",
				},
			},
			expectedArgs: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1", "otherkey=value"},
				},
			},
			wantErr: require.Error,
		},
		{ // unlikely to happen
			name: "fails when format for literal sources is incorrect in existing args",
			args: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2", "otherkey=value"},
				},
			},
			flags: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{
					"key2=val2",
				},
			},
			expectedArgs: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2", "otherkey=value"},
				},
			},
			wantErr: require.Error,
		},
		{
			name: "fails when literal sources from flags contain more than one =",
			args: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1", "otherkey=value"},
				},
			},
			flags: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{
					"key2=val2=val3",
				},
			},
			expectedArgs: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1", "otherkey=value"},
				},
			},
			wantErr: require.Error,
		},
		{ // unlikely to happen
			name: "fails when literal sources contain more than one = in existing args",
			args: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1=val3", "otherkey=value"},
				},
			},
			flags: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{
					"key2=val2",
				},
			},
			expectedArgs: &types.GeneratorArgs{
				KvPairSources: types.KvPairSources{
					LiteralSources: []string{"key2=val1=val3", "otherkey=value"},
				},
			},
			wantErr: require.Error,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.wantErr(t, UpdateLiteralSources(tc.args, tc.flags))
			require.ElementsMatch(t, tc.expectedArgs.LiteralSources, tc.args.LiteralSources)
		})
	}
}
