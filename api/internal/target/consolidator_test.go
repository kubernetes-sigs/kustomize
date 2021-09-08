package target

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func TestTransformerSorter_sortTransformers_Success(t *testing.T) {
	tests := []struct {
		name         string
		transformers []types.Transformer
		ordering     []resid.ResId
		expected     []types.Transformer
	}{
		{
			name: "fully specified reorder",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "third",
					},
					TypeMeta: types.TypeMeta{
						APIVersion: "org1.example.co/v1alpha1",
						Kind:       "Labeller",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Name: "second",
					Gvk: resid.Gvk{
						Kind:    "Annotator",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Name: "third",
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
			},
			expected: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "third",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
		},
		{
			name: "reorder with names only",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "third",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{Name: "first"},
				{Name: "second"},
				{Name: "third"},
			},
			expected: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "third",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
		},
		{
			name: "reorder with ambiguity resolved by kind",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind: "Labeller",
					},
				},
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind: "Annotator",
					},
				},
				{
					Name: "second",
				},
			},
			expected: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
		},
		{
			name: "reorder with ambiguity resolved by apiVersion",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Name: "first",
					Gvk: resid.Gvk{
						Group:   "org2.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Name: "second",
				},
			},
			expected: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sorter := transformerSorter{ordering: tt.ordering}
			result, err := sorter.sortTransformers(tt.transformers)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformerSorter_sortTransformers_Error(t *testing.T) {
	tests := []struct {
		name         string
		transformers []types.Transformer
		ordering     []resid.ResId
		errorMsg     string
	}{
		{
			name: "name is required",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "third",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Annotator",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Gvk: resid.Gvk{
						Kind:    "Annotator",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Version: "org2.example.co/v1alpha1",
					},
				},
			},
			errorMsg: "failed to sort transformers: transformer identifier \"org1.example.co_v1alpha1_Labeller|~X|~N\" must include a name",
		},
		{
			name: "ordering cannot contain duplicates - names only",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{Name: "second"},
				{Name: "first"},
				{Name: "second"},
			},
			errorMsg: "failed to sort transformers: list contains multiple \"~G_~V_~K|~X|second\" transformers",
		},
		{
			name: "ordering cannot contain duplicates - GKN",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Group:   "org1.example.co",
						Version: "v1alpha1"},
				},
			},
			errorMsg: "failed to sort transformers: list contains multiple \"org1.example.co_v1alpha1_Labeller|~X|first\" transformers",
		},
		{
			name: "ambiguous ordering - GK match and GVN match",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{

						Kind: "Labeller",
					},
				},
				{
					Name: "first",
					Gvk: resid.Gvk{
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
			},
			errorMsg: "unable to find position for transformer \"org1.example.co_v1alpha1_Labeller|~X|first\": multiple entries matched",
		},
		{
			name: "ambiguous ordering - GVKN match and KN match",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{
					Name: "first",
					Gvk: resid.Gvk{

						Kind: "Labeller",
					},
				},
			},
			errorMsg: "unable to find position for transformer \"org1.example.co_v1alpha1_Labeller|~X|first\": multiple entries matched",
		},
		{
			name: "ambiguous ordering - GVN match and name match",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{

						Group:   "org1.example.co",
						Version: "v1alpha1",
					},
				},
				{Name: "first"},
			},
			errorMsg: "unable to find position for transformer \"org1.example.co_v1alpha1_Labeller|~X|first\": multiple entries matched",
		},
		{
			name: "transformer order must be complete",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org2.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{

						Kind: "Labeller",
					},
				},
			},
			errorMsg: "transformer order list contains too few entries",
		},
		{
			name: "transformer order cannot have extra entries",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "second",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{Name: "first"},
				{Name: "second"},
				{Name: "third"},
			},
			errorMsg: "transformer order list contains too many entries",
		},
		{
			name: "unable to find transformer by name and kind",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind: "Annotator",
					},
				},
			},
			errorMsg: "unable to find position for transformer \"org1.example.co_v1alpha1_Labeller|~X|first\": no match found",
		},
		{
			name: "unable to find transformer by name, kind and GV",
			transformers: []types.Transformer{
				{
					MetaData: types.ObjectMeta{
						Name: "first",
					},
					TypeMeta: types.TypeMeta{
						Kind:       "Labeller",
						APIVersion: "org1.example.co/v1alpha1",
					},
				},
			},
			ordering: []resid.ResId{
				{
					Name: "first",
					Gvk: resid.Gvk{
						Kind:    "Labeller",
						Version: "org3.example.co/v1alpha1",
					},
				},
			},
			errorMsg: "unable to find position for transformer \"org1.example.co_v1alpha1_Labeller|~X|first\": no match found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sorter := transformerSorter{ordering: tt.ordering}
			_, err := sorter.sortTransformers(tt.transformers)
			require.Error(t, err)
			require.EqualError(t, err, tt.errorMsg)
		})
	}
}
