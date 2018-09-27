package transformerconfig

import (
	"testing"

	"reflect"
	"sigs.k8s.io/kustomize/pkg/gvk"
)

func TestAddNameReferencePathConfigs(t *testing.T) {
	cfg := MakeEmptyTransformerConfig()

	pathConfig := ReferencePathConfig{
		Gvk: gvk.Gvk{
			Kind: "KindA",
		},
		PathConfigs: []PathConfig{
			{
				Gvk: gvk.Gvk{
					Kind: "KindB",
				},
				Path:               "path/to/a/field",
				CreateIfNotPresent: false,
			},
		},
	}

	cfg.AddNamereferencePathConfig(pathConfig)
	if len(cfg.NameReference) != 1 {
		t.Fatal("failed to add namerefence pathconfig")
	}
}

func TestAddPathConfigs(t *testing.T) {
	cfg := MakeEmptyTransformerConfig()

	pathConfig := PathConfig{
		Gvk:                gvk.Gvk{Group: "GroupA", Kind: "KindB"},
		Path:               "path/to/a/field",
		CreateIfNotPresent: true,
	}

	cfg.AddPrefixPathConfig(pathConfig)
	if len(cfg.NamePrefix) != 1 {
		t.Fatalf("failed to add nameprefix pathconfig")
	}
	cfg.AddLabelPathConfig(pathConfig)
	if len(cfg.CommonLabels) != 1 {
		t.Fatalf("failed to add nameprefix pathconfig")
	}
	cfg.AddAnnotationPathConfig(pathConfig)
	if len(cfg.CommonAnnotations) != 1 {
		t.Fatalf("failed to add nameprefix pathconfig")
	}
}

func TestMerge(t *testing.T) {
	nameReference := []ReferencePathConfig{
		{
			Gvk: gvk.Gvk{
				Kind: "KindA",
			},
			PathConfigs: []PathConfig{
				{
					Gvk: gvk.Gvk{
						Kind: "KindB",
					},
					Path:               "path/to/a/field",
					CreateIfNotPresent: false,
				},
			},
		},
		{
			Gvk: gvk.Gvk{
				Kind: "KindA",
			},
			PathConfigs: []PathConfig{
				{
					Gvk: gvk.Gvk{
						Kind: "KindC",
					},
					Path:               "path/to/a/field",
					CreateIfNotPresent: false,
				},
			},
		},
	}
	pathConfigs := []PathConfig{
		{
			Gvk:                gvk.Gvk{Group: "GroupA", Kind: "KindB"},
			Path:               "path/to/a/field",
			CreateIfNotPresent: true,
		},
		{
			Gvk:                gvk.Gvk{Group: "GroupA", Kind: "KindC"},
			Path:               "path/to/a/field",
			CreateIfNotPresent: true,
		},
	}
	cfga := MakeEmptyTransformerConfig()
	cfga.AddNamereferencePathConfig(nameReference[0])
	cfga.AddPrefixPathConfig(pathConfigs[0])

	cfgb := MakeEmptyTransformerConfig()
	cfgb.AddNamereferencePathConfig(nameReference[1])
	cfgb.AddPrefixPathConfig(pathConfigs[1])

	actual := cfga.Merge(cfgb)

	if len(actual.NamePrefix) != 2 {
		t.Fatal("merge failed for namePrefix pathconfig")
	}

	if len(actual.NameReference) != 1 {
		t.Fatal("merge failed for namereference pathconfig")
	}

	expected := MakeEmptyTransformerConfig()
	expected.AddNamereferencePathConfig(nameReference[0])
	expected.AddNamereferencePathConfig(nameReference[1])
	expected.AddPrefixPathConfig(pathConfigs[0])
	expected.AddPrefixPathConfig(pathConfigs[1])

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected: %v\n but got: %v\n", expected, actual)
	}
}
