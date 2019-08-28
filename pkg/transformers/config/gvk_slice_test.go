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
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
)

func Test_gvkSlice_merge(t *testing.T) {
	type args struct {
		incoming gvkSlice
	}
	tests := []struct {
		name       string
		s          gvkSlice
		args       args
		wantResult gvkSlice
	}{
		{
			name: "merge nil and nil",
			s:    nil,
			args: args{
				nil,
			},
			wantResult: nil,
		},
		{
			name: "merge into nil",
			s:    nil,
			args: args{
				gvkSlice{gvk.Gvk{Kind: "kind-a"}},
			},
			wantResult: gvkSlice{gvk.Gvk{Kind: "kind-a"}},
		},
		{
			name: "merge from nil",
			s:    gvkSlice{gvk.Gvk{Kind: "kind-a"}},
			args: args{
				nil,
			},
			wantResult: gvkSlice{gvk.Gvk{Kind: "kind-a"}},
		},
		{
			name: "merge uniques",
			s:    gvkSlice{gvk.Gvk{Kind: "kind-a"}},
			args: args{
				gvkSlice{gvk.Gvk{Kind: "kind-b"}},
			},
			wantResult: gvkSlice{gvk.Gvk{Kind: "kind-a"}, gvk.Gvk{Kind: "kind-b"}},
		},
		{
			name: "merge overlapping",
			s:    gvkSlice{gvk.Gvk{Kind: "kind-a"}},
			args: args{
				gvkSlice{gvk.Gvk{Kind: "kind-a"}},
			},
			wantResult: gvkSlice{gvk.Gvk{Kind: "kind-a"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.s.merge(tt.args.incoming); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("gvkSlice.merge() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
