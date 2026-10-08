// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"reflect"
	"strings"
	"testing"
)

func fixKustomizationPostUnmarshallingCheck(k, e *Kustomization) bool {
	return k.Kind == e.Kind &&
		k.APIVersion == e.APIVersion &&
		len(k.Resources) == len(e.Resources) &&
		k.Resources[0] == e.Resources[0] &&
		k.Bases == nil
}

func TestKustomization_CheckDeprecatedFields(t *testing.T) {
	tests := []struct {
		name string
		k    Kustomization
		want *[]string
	}{
		{
			name: "using_bases",
			k: Kustomization{
				Bases: []string{"base"},
			},
			want: &[]string{deprecatedBaseWarningMessage},
		},
		{
			name: "using_CommonLabels",
			k: Kustomization{
				CommonLabels: map[string]string{},
			},
			want: &[]string{deprecatedCommonLabelsWarningMessage},
		},
		{
			name: "using_ImageTags",
			k: Kustomization{
				ImageTags: []Image{},
			},
			want: &[]string{deprecatedImageTagsWarningMessage},
		},
		{
			name: "usingPatchesJson6902",
			k: Kustomization{
				PatchesJson6902: []Patch{},
			},
			want: &[]string{deprecatedPatchesJson6902Message},
		},
		{
			name: "usingPatchesStrategicMerge",
			k: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{},
			},
			want: &[]string{deprecatedPatchesStrategicMergeMessage},
		},
		{
			name: "usingVar",
			k: Kustomization{
				Vars: []Var{},
			},
			want: &[]string{deprecatedVarsMessage},
		},
		{
			name: "usingAll",
			k: Kustomization{
				Bases:                 []string{"base"},
				CommonLabels:          map[string]string{},
				ImageTags:             []Image{},
				PatchesJson6902:       []Patch{},
				PatchesStrategicMerge: []PatchStrategicMerge{},
				Vars:                  []Var{},
			},
			want: &[]string{
				deprecatedBaseWarningMessage,
				deprecatedCommonLabelsWarningMessage,
				deprecatedImageTagsWarningMessage,
				deprecatedPatchesJson6902Message,
				deprecatedPatchesStrategicMergeMessage,
				deprecatedVarsMessage,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.k
			if got := k.CheckDeprecatedFields(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Kustomization.CheckDeprecatedFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFixKustomizationPostUnmarshalling(t *testing.T) {
	var k Kustomization
	k.Bases = append(k.Bases, "foo")
	k.ConfigMapGenerator = []ConfigMapArgs{{GeneratorArgs{
		KvPairSources: KvPairSources{
			EnvSources: []string{"a", "b"},
			EnvSource:  "c",
		},
	}}}
	k.CommonLabels = map[string]string{
		"foo": "bar",
	}
	k.FixKustomization()

	expected := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       KustomizationKind,
			APIVersion: KustomizationVersion,
		},
		Resources: []string{"foo"},
		ConfigMapGenerator: []ConfigMapArgs{{GeneratorArgs{
			KvPairSources: KvPairSources{
				EnvSources: []string{"a", "b", "c"},
			},
		}}},
		CommonLabels: map[string]string{
			"foo": "bar",
		},
	}
	if !reflect.DeepEqual(k, expected) {
		t.Fatalf("unexpected output: %v", k)
	}
	if !fixKustomizationPostUnmarshallingCheck(&k, &expected) {
		t.Fatalf("unexpected output: %v", k)
	}
}

func TestFixKustomizationPostUnmarshalling_2(t *testing.T) {
	k := Kustomization{
		TypeMeta: TypeMeta{
			Kind: ComponentKind,
		},
	}
	k.Bases = append(k.Bases, "foo")
	k.FixKustomization()

	expected := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       ComponentKind,
			APIVersion: ComponentVersion,
		},
		Resources: []string{"foo"},
	}

	if !fixKustomizationPostUnmarshallingCheck(&k, &expected) {
		t.Fatalf("unexpected output: %v", k)
	}
}

func TestEnforceFields_InvalidKindAndVersion(t *testing.T) {
	k := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       "foo",
			APIVersion: "bar",
		},
	}

	errs := k.EnforceFields()
	if len(errs) != 2 {
		t.Fatalf("number of errors should be 2 but got: %v", errs)
	}
}

func TestEnforceFields_InvalidKind(t *testing.T) {
	k := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       "foo",
			APIVersion: KustomizationVersion,
		},
	}

	errs := k.EnforceFields()
	if len(errs) != 1 {
		t.Fatalf("number of errors should be 1 but got: %v", errs)
	}

	expected := "kind should be " + KustomizationKind + " or " + ComponentKind
	if errs[0] != expected {
		t.Fatalf("error should be %v but got: %v", expected, errs[0])
	}
}

func TestEnforceFields_InvalidVersion(t *testing.T) {
	k := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       KustomizationKind,
			APIVersion: "bar",
		},
	}

	errs := k.EnforceFields()
	if len(errs) != 1 {
		t.Fatalf("number of errors should be 1 but got: %v", errs)
	}

	expected := "apiVersion for " + k.Kind + " should be " + KustomizationVersion
	if errs[0] != expected {
		t.Fatalf("error should be %v but got: %v", expected, errs[0])
	}
}

func TestEnforceFields_ComponentKind(t *testing.T) {
	k := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       ComponentKind,
			APIVersion: "bar",
		},
	}

	errs := k.EnforceFields()
	if len(errs) != 1 {
		t.Fatalf("number of errors should be 1 but got: %v", errs)
	}

	expected := "apiVersion for " + k.Kind + " should be " + ComponentVersion
	if errs[0] != expected {
		t.Fatalf("error should be %v but got: %v", expected, errs[0])
	}
}

func TestEnforceFields(t *testing.T) {
	k := Kustomization{
		TypeMeta: TypeMeta{
			Kind:       KustomizationKind,
			APIVersion: KustomizationVersion,
		},
	}

	errs := k.EnforceFields()
	if len(errs) != 0 {
		t.Fatalf("number of errors should be 0 but got: %v", errs)
	}
}

func TestUnmarshal(t *testing.T) {
	y := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: kust
  namespace: default
  labels:
    foo: bar
  annotations:
    foo: bar
resources:
- foo
- bar
nameSuffix: dog
namePrefix: cat`)
	var k Kustomization
	err := k.Unmarshal(y)
	if err != nil {
		t.Fatal(err)
	}
	meta := ObjectMeta{
		Name:      "kust",
		Namespace: "default",
		Labels: map[string]string{
			"foo": "bar",
		},
		Annotations: map[string]string{
			"foo": "bar",
		},
	}
	if k.Kind != KustomizationKind || k.APIVersion != KustomizationVersion ||
		len(k.Resources) != 2 || k.NamePrefix != "cat" || k.NameSuffix != "dog" ||
		k.MetaData.Name != meta.Name || k.MetaData.Namespace != meta.Namespace ||
		k.MetaData.Labels["foo"] != meta.Labels["foo"] || k.MetaData.Annotations["foo"] != meta.Annotations["foo"] {
		t.Fatalf("wrong unmarshal result: %v", k)
	}
}

func TestUnmarshal_UnkownField(t *testing.T) {
	y := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
unknown: foo`)
	var k Kustomization
	err := k.Unmarshal(y)
	if err == nil {
		t.Fatalf("expect an error")
	}
	expect := "invalid Kustomization: json: unknown field \"unknown\""
	if err.Error() != expect {
		t.Fatalf("expect %v but got: %v", expect, err.Error())
	}
}

func TestUnmarshal_Failed(t *testing.T) {
	tests := []struct {
		name               string
		kustomizationYamls []byte
		errMsg             string
	}{
		{
			name: "invalid yaml",
			kustomizationYamls: []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
unknown`),
			errMsg: "invalid Kustomization: yaml: line 4: could not find expected ':'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var k Kustomization
			if err := k.Unmarshal(tt.kustomizationYamls); err == nil || err.Error() != tt.errMsg {
				t.Errorf("Kustomization.Unmarshal() error = %v, wantErr %v", err, tt.errMsg)
			}
		})
	}
}

func TestUnmarshalInlinePatchTypeError(t *testing.T) {
	for _, field := range []string{"patches", "patchesJson6902"} {
		for _, patch := range []struct {
			name  string
			value string
		}{
			{name: "sequence", value: "[{op: add, path: /metadata/labels, value: {app: test}}]"},
			{name: "mapping", value: "{apiVersion: v1, kind: ConfigMap, metadata: {name: test}}"},
		} {
			t.Run(field+"/"+patch.name, func(t *testing.T) {
				k := Kustomization{NamePrefix: "unchanged-"}
				err := k.Unmarshal([]byte(field + ":\n- patch: " + patch.value + "\n"))
				if err == nil {
					t.Fatal("expected an error for a non-string inline patch")
				}
				want := field + ".patch must be a string; use 'patch: |' for a multiline patch"
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want guidance %q", err, want)
				}
				if !reflect.DeepEqual(k, Kustomization{NamePrefix: "unchanged-"}) {
					t.Errorf("failed unmarshal modified the receiver: %#v", k)
				}
			})
		}
	}
}

func TestUnmarshalInlinePatchString(t *testing.T) {
	for _, field := range []string{"patches", "patchesJson6902"} {
		t.Run(field, func(t *testing.T) {
			var k Kustomization
			err := k.Unmarshal([]byte(field + ":\n- patch: |\n    - op: add\n      path: /metadata/labels\n      value: {app: test}\n"))
			if err != nil {
				t.Fatal(err)
			}
			patches := k.Patches
			if field == "patchesJson6902" {
				patches = k.PatchesJson6902
			}
			want := "- op: add\n  path: /metadata/labels\n  value: {app: test}\n"
			if len(patches) != 1 || patches[0].Patch != want {
				t.Errorf("patches = %#v, want one patch containing %q", patches, want)
			}
		})
	}
}

func TestUnmarshal_NonPatchTypeError(t *testing.T) {
	var k Kustomization
	err := k.Unmarshal([]byte("patches:\n- path: [patch.yaml]\n"))
	if err == nil {
		t.Fatal("expected an error for a non-string patch file path")
	}
	if strings.Contains(err.Error(), "patch: |") {
		t.Errorf("unrelated type error received inline patch guidance: %v", err)
	}
}

func TestKustomization_CheckEmpty(t *testing.T) {
	tests := []struct {
		name          string
		kustomization *Kustomization
		wantErr       bool
	}{
		{
			name:          "empty kustomization.yaml",
			kustomization: &Kustomization{},
			wantErr:       true,
		},
		{
			name: "empty kustomization.yaml",
			kustomization: &Kustomization{
				TypeMeta: TypeMeta{
					Kind:       KustomizationKind,
					APIVersion: KustomizationVersion,
				},
			},
			wantErr: true,
		},
		{
			name:          "non empty kustomization.yaml",
			kustomization: &Kustomization{Resources: []string{"res"}},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.kustomization
			k.FixKustomization()
			if err := k.CheckEmpty(); (err != nil) != tt.wantErr {
				t.Errorf("Kustomization.CheckEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
