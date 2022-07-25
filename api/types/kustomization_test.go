// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"reflect"
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
				PatchesJson6902:       []Patch{},
				PatchesStrategicMerge: []PatchStrategicMerge{},
				Vars:                  []Var{},
			},
			want: &[]string{
				deprecatedBaseWarningMessage,
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
	k.FixKustomizationPostUnmarshalling()

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
	k.FixKustomizationPostUnmarshalling()

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
	expect := "json: unknown field \"unknown\""
	if err.Error() != expect {
		t.Fatalf("expect %v but got: %v", expect, err.Error())
	}
}

func TestUnmarshal_InvalidYaml(t *testing.T) {
	y := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
unknown`)
	var k Kustomization
	err := k.Unmarshal(y)
	if err == nil {
		t.Fatalf("expect an error")
	}
}
