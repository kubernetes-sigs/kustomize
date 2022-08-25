// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filtersutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestTrackableSetter_SetScalarIfEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input *yaml.RNode
		value string
		want  string
	}{
		{
			name:  "sets null values",
			input: yaml.MakeNullNode(),
			value: "foo",
			want:  "foo",
		},
		{
			name:  "sets empty values",
			input: yaml.NewScalarRNode(""),
			value: "foo",
			want:  "foo",
		},
		{
			name:  "does not overwrite values",
			input: yaml.NewStringRNode("a"),
			value: "foo",
			want:  "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wasSet := false
			s := (&filtersutil.TrackableSetter{}).WithMutationTracker(func(_, _, _ string, _ *yaml.RNode) {
				wasSet = true
			})
			wantSet := tt.value == tt.want
			fn := s.SetScalarIfEmpty(tt.value)
			require.NoError(t, fn(tt.input))
			assert.Equal(t, tt.want, yaml.GetValue(tt.input))
			assert.Equal(t, wantSet, wasSet, "tracker invoked even though value was not changed")
		})
	}
}

func TestTrackableSetter_SetEntryIfEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input *yaml.RNode
		key   string
		value string
		want  string
	}{
		{
			name:  "sets empty values",
			input: yaml.NewMapRNode(&map[string]string{"setMe": ""}),
			key:   "setMe",
			value: "foo",
			want:  "foo",
		},
		{
			name:  "sets missing keys",
			input: yaml.NewMapRNode(&map[string]string{}),
			key:   "setMe",
			value: "foo",
			want:  "foo",
		},
		{
			name:  "does not overwrite values",
			input: yaml.NewMapRNode(&map[string]string{"existing": "original"}),
			key:   "existing",
			value: "foo",
			want:  "original",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wasSet := false
			s := (&filtersutil.TrackableSetter{}).WithMutationTracker(func(_, _, _ string, _ *yaml.RNode) {
				wasSet = true
			})
			wantSet := tt.value == tt.want
			fn := s.SetEntryIfEmpty(tt.key, tt.value, "")
			require.NoError(t, fn(tt.input))
			assert.Equal(t, tt.want, yaml.GetValue(tt.input.Field(tt.key).Value))
			assert.Equal(t, wantSet, wasSet, "tracker invoked even though value was not changed")
		})
	}
}

func TestTrackableSetter_SetEntryIfEmpty_BadInputNodeKind(t *testing.T) {
	fn := filtersutil.TrackableSetter{}.SetEntryIfEmpty("foo", "false", yaml.NodeTagBool)
	rn := yaml.NewListRNode("nope")
	rn.AppendToFieldPath("dummy", "path")
	assert.EqualError(t, fn(rn), `wrong node kind: expected MappingNode but got SequenceNode: node contents:
- nope
`)
}
