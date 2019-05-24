// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configmapandsecret

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kv"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
)

func TestKeyValuesFromFileSources(t *testing.T) {
	tests := []struct {
		description string
		sources     []string
		expected    []kv.Pair
	}{
		{
			description: "create kvs from file sources",
			sources:     []string{"files/app-init.ini"},
			expected: []kv.Pair{
				{
					Key:   "app-init.ini",
					Value: "FOO=bar",
				},
			},
		},
	}

	fSys := fs.MakeFakeFS()
	fSys.WriteFile("/files/app-init.ini", []byte("FOO=bar"))
	bf := NewFactory(loader.NewFileLoaderAtRoot(fSys), nil)
	for _, tc := range tests {
		kvs, err := bf.keyValuesFromFileSources(tc.sources)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(kvs, tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, kvs, tc.expected)
		}
	}
}
