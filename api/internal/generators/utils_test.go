// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestParseFileSource(t *testing.T) {
	tests := map[string]*struct {
		Input    string
		Error    string
		Key      string
		Filename string
	}{
		"filename only": {
			Input:    "./path/myfile",
			Key:      "myfile",
			Filename: "./path/myfile",
		},
		"key and filename": {
			Input:    "newName.ini=oldName",
			Key:      "newName.ini",
			Filename: "oldName",
		},
		"multiple =": {
			Input: "newName.ini==oldName",
			Error: `source "newName.ini==oldName" key name or file path contains '='`,
		},
		"missing key": {
			Input: "=myfile",
			Error: `missing key name for file path "myfile" in source "=myfile"`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			key, file, err := ParseFileSource(test.Input)
			if test.Error != "" {
				require.EqualError(t, err, test.Error)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.Key, key)
				require.Equal(t, test.Filename, file)
			}
		})
	}
}

func TestMakeValidateDataMapOrder(t *testing.T) {
	tests := []struct {
		name        string
		pairs       []types.Pair
		wantKeys    []string
		wantValues  map[string]string
		wantErr     bool
		validatorFn func() ifc.Validator
	}{
		{
			name: "single pair",
			pairs: []types.Pair{
				{Key: "key", Value: "value"},
			},
			wantKeys:   []string{"key"},
			wantValues: map[string]string{"key": "value"},
		},
		{
			name: "multiple pairs preserve order",
			pairs: []types.Pair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
				{Key: "C", Value: "3"},
			},
			wantKeys:   []string{"A", "B", "C"},
			wantValues: map[string]string{"A": "1", "B": "2", "C": "3"},
		},
		{
			name: "multiple pairs out of order - loader order wins",
			pairs: []types.Pair{
				{Key: "third", Value: "3"},
				{Key: "first", Value: "1"},
				{Key: "second", Value: "2"},
			},
			wantKeys:   []string{"third", "first", "second"},
			wantValues: map[string]string{"third": "3", "first": "1", "second": "2"},
		},
		{
			name: "duplicate key returns error",
			pairs: []types.Pair{
				{Key: "dup", Value: "1"},
				{Key: "dup", Value: "2"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &fakeKvLoader{
				Pairs: tt.pairs,
				Val: func() ifc.Validator {
					if tt.validatorFn != nil {
						return tt.validatorFn()
					}
					return valtest_test.MakeFakeValidator()
				}(),
			}

			args := types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "envConfigMap",
					KvPairSources: types.KvPairSources{
						EnvSources: []string{"app.env"},
					},
				},
			}

			res, err := makeValidatedDataMap(loader, args.Name, args.KvPairSources)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Exactly(t, tt.wantKeys, res.keys)
			require.Exactly(t, tt.wantValues, res.values)
		})
	}
}

type fakeKvLoader struct {
	Pairs []types.Pair
	Val   ifc.Validator
}

func (f *fakeKvLoader) Load(_ types.KvPairSources) ([]types.Pair, error) {
	return f.Pairs, nil
}

func (f *fakeKvLoader) Validator() ifc.Validator {
	if f.Val != nil {
		return f.Val
	}
	return valtest_test.MakeFakeValidator()
}
