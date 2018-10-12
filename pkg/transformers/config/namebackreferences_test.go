/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"reflect"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"testing"
)

func TestFoo(t *testing.T) {
	fsSlice1 := []FieldSpec{
		{
			Gvk: gvk.Gvk{
				Kind: "Pod",
			},
			Path:               "path/to/a/name",
			CreateIfNotPresent: false,
		},
		{
			Gvk: gvk.Gvk{
				Kind: "Deployment",
			},
			Path:               "another/path/to/some/name",
			CreateIfNotPresent: false,
		},
	}
	fsSlice2 := []FieldSpec{
		{
			Gvk: gvk.Gvk{
				Kind: "Job",
			},
			Path:               "morepath/to/name",
			CreateIfNotPresent: false,
		},
		{
			Gvk: gvk.Gvk{
				Kind: "StatefulSet",
			},
			Path:               "yet/another/path/to/a/name",
			CreateIfNotPresent: false,
		},
	}

	nbrsSlice1 := []NameBackReferences{
		{
			Gvk: gvk.Gvk{
				Kind: "ConfigMap",
			},
			FieldSpecs: fsSlice1,
		},
		{
			Gvk: gvk.Gvk{
				Kind: "Secret",
			},
			FieldSpecs: fsSlice2,
		},
	}
	nbrsSlice2 := []NameBackReferences{
		{
			Gvk: gvk.Gvk{
				Kind: "ConfigMap",
			},
			FieldSpecs: fsSlice1,
		},
		{
			Gvk: gvk.Gvk{
				Kind: "Secret",
			},
			FieldSpecs: fsSlice2,
		},
	}
	expected := []NameBackReferences{
		{
			Gvk: gvk.Gvk{
				Kind: "ConfigMap",
			},
			// Current behavior allows repeats of FieldSpec
			FieldSpecs: append(fsSlice1, fsSlice1...),
		},
		{
			Gvk: gvk.Gvk{
				Kind: "Secret",
			},
			FieldSpecs: append(fsSlice2, fsSlice2...),
		},
	}
	actual := mergeNameBackReferences(nbrsSlice1, nbrsSlice2)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected\n %v\n but got\n %v\n", expected, actual)
	}
}
