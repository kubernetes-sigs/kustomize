// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
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
				ImageTags:             []Image{},
				PatchesJson6902:       []Patch{},
				PatchesStrategicMerge: []PatchStrategicMerge{},
				Vars:                  []Var{},
			},
			want: &[]string{
				deprecatedBaseWarningMessage,
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

func TestFixKustomizationPreMarshalling_SplitPatches(t *testing.T) {
	testCases := map[string]struct {
		kustomization         Kustomization
		files                 map[string]string
		expectedKustomization Kustomization
		expectedFiles         map[string]string
	}{
		"no split needed": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml"},
			},
			files: map[string]string{
				"secret.patch.yaml": `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Path: "secret.patch.yaml"}},
			},
			expectedFiles: map[string]string{
				"secret.patch.yaml": `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
			},
		},

		"no split needed (inline)": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{`
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`},
			},
			files: map[string]string{},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Patch: `
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`}},
			},
			expectedFiles: map[string]string{},
		},

		"remove unnecessary document separators": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml"},
			},
			files: map[string]string{
				"secret.patch.yaml": `
--- # Some comment
---
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
---
---
`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Path: "secret.patch.yaml"}},
			},
			expectedFiles: map[string]string{
				"secret.patch.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
			},
		},

		"remove unnecessary document separators (inline)": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{`
--- # Some comment
---
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
---
---
`},
			},
			files: map[string]string{},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Patch: `apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`}},
			},
			expectedFiles: map[string]string{},
		},

		"split into two patches": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml"},
			},
			files: map[string]string{
				"secret.patch.yaml": `
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Path: "secret.patch-1.yaml"}, {Path: "secret.patch-2.yaml"}},
			},
			expectedFiles: map[string]string{
				"secret.patch-1.yaml": `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
				"secret.patch-2.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
			},
		},

		"split into two patches (inline)": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{`
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`},
			},
			files: map[string]string{},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Patch: `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`}, {Patch: `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`}},
			},
			expectedFiles: map[string]string{},
		},

		"split should not affect existing patch": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml"},
				Patches:               []Patch{{Path: "existing.patch.yaml"}},
			},
			files: map[string]string{
				"secret.patch.yaml": `
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"existing.patch.yaml": `
apiVersion: v1
kind: Foo
metadata:
  name: foo1
spec: something
`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Path: "existing.patch.yaml"}, {Path: "secret.patch-1.yaml"}, {Path: "secret.patch-2.yaml"}},
			},
			expectedFiles: map[string]string{
				"secret.patch-1.yaml": `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
				"secret.patch-2.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"existing.patch.yaml": `
apiVersion: v1
kind: Foo
metadata:
  name: foo1
spec: something
`,
			},
		},

		"split into two patches handle filename collision": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml"},
			},
			files: map[string]string{
				"secret.patch.yaml": `
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"secret.patch-1.yaml":       `# I'm here to cause a filename collision`,
				"secret.patch-2.yaml":       `# I'm here to cause a filename collision`,
				"secret.patch-2-2.yaml":     `# I'm here to cause a filename collision`,
				"secret.patch-2-2-2.yaml":   `# I'm here to cause a filename collision`,
				"secret.patch-2-2-2-2.yaml": `# I'm here to cause a filename collision`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Path: "secret.patch-1-1.yaml"}, {Path: "secret.patch-2-2-2-2-2.yaml"}},
			},
			expectedFiles: map[string]string{
				"secret.patch-1-1.yaml": `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
				"secret.patch-2-2-2-2-2.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"secret.patch-1.yaml":       `# I'm here to cause a filename collision`,
				"secret.patch-2.yaml":       `# I'm here to cause a filename collision`,
				"secret.patch-2-2.yaml":     `# I'm here to cause a filename collision`,
				"secret.patch-2-2-2.yaml":   `# I'm here to cause a filename collision`,
				"secret.patch-2-2-2-2.yaml": `# I'm here to cause a filename collision`,
			},
		},

		"split into two patches and handle unnecessary document separators": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml"},
			},
			files: map[string]string{
				"secret.patch.yaml": `
---
---
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
--- # Something here
---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
---

`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{{Path: "secret.patch-1.yaml"}, {Path: "secret.patch-2.yaml"}},
			},
			expectedFiles: map[string]string{
				"secret.patch-1.yaml": `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
				"secret.patch-2.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
			},
		},

		"split multiple patches into multiple patches": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml", "foo.patch.yaml", "bar.patch.yaml"},
			},
			files: map[string]string{
				"secret.patch.yaml": `
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"foo.patch.yaml": `
apiVersion: v1
kind: Foo
metadata:
  name: foo1
spec: something
---
apiVersion: v1
kind: Foo
metadata:
  name: foo2
spec: something
`,
				"bar.patch.yaml": `
apiVersion: v1
kind: Bar
metadata:
  name: bar1
spec: something
---
apiVersion: v1
kind: Bar
metadata:
  name: bar2
spec: something
`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{
					{Path: "secret.patch-1.yaml"},
					{Path: "secret.patch-2.yaml"},
					{Path: "foo.patch-1.yaml"},
					{Path: "foo.patch-2.yaml"},
					{Path: "bar.patch-1.yaml"},
					{Path: "bar.patch-2.yaml"},
				},
			},
			expectedFiles: map[string]string{
				"secret.patch-1.yaml": `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
				"secret.patch-2.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"foo.patch-1.yaml": `apiVersion: v1
kind: Foo
metadata:
  name: foo1
spec: something
`,
				"foo.patch-2.yaml": `apiVersion: v1
kind: Foo
metadata:
  name: foo2
spec: something
`,
				"bar.patch-1.yaml": `apiVersion: v1
kind: Bar
metadata:
  name: bar1
spec: something
`,
				"bar.patch-2.yaml": `apiVersion: v1
kind: Bar
metadata:
  name: bar2
spec: something
`,
			},
		},

		"split multiple patches into multiple patches (mix of files and inline)": {
			kustomization: Kustomization{
				PatchesStrategicMerge: []PatchStrategicMerge{"secret.patch.yaml", "foo.patch.yaml", `
--- # Some comment
apiVersion: v1
kind: Bar
metadata:
  name: bar1
spec: something
---
---   
apiVersion: v1
kind: Bar
metadata:
  name: bar2
spec: something
---
`},
			},
			files: map[string]string{
				"secret.patch.yaml": `
# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword

---
apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"foo.patch.yaml": `
apiVersion: v1
kind: Foo
metadata:
  name: foo1
spec: something
---
apiVersion: v1
kind: Foo
metadata:
  name: foo2
spec: something
`,
			},
			expectedKustomization: Kustomization{
				Patches: []Patch{
					{Path: "secret.patch-1.yaml"},
					{Path: "secret.patch-2.yaml"},
					{Path: "foo.patch-1.yaml"},
					{Path: "foo.patch-2.yaml"},
					{Patch: `apiVersion: v1
kind: Bar
metadata:
  name: bar1
spec: something
`},
					{Patch: `apiVersion: v1
kind: Bar
metadata:
  name: bar2
spec: something
`},
				},
			},
			expectedFiles: map[string]string{
				"secret.patch-1.yaml": `# secret.patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret1
type: Opaque
stringData:
  admin.password: newpassword
`,
				"secret.patch-2.yaml": `apiVersion: v1
kind: Secret
metadata:
  name: secret2
type: Opaque
stringData:
  admin.user: newuser
`,
				"foo.patch-1.yaml": `apiVersion: v1
kind: Foo
metadata:
  name: foo1
spec: something
`,
				"foo.patch-2.yaml": `apiVersion: v1
kind: Foo
metadata:
  name: foo2
spec: something
`,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			for filename, content := range tc.files {
				require.NoError(t, fSys.WriteFile(filename, []byte(content)))
			}

			err := tc.kustomization.FixKustomizationPreMarshalling(fSys)
			if err != nil {
				t.Fatal(err)
			}

			require.Emptyf(t, cmp.Diff(tc.expectedKustomization, tc.kustomization), "kustomization mismatch")

			for filename, expectedContent := range tc.expectedFiles {
				actualContent, err := fSys.ReadFile(filename)
				if err != nil {
					t.Fatalf("failed to read expected file %s: %v", filename, err)
				}
				require.Emptyf(t, cmp.Diff(expectedContent, string(actualContent)), "file %s content mismatch", filename)
			}

			for filename := range tc.files {
				if _, ok := tc.expectedFiles[filename]; ok {
					continue
				}
				if fSys.Exists(filename) {
					t.Errorf("expected file %s to have been deleted", filename)
				}
			}
		})
	}
}
