package nameref

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestNamerefFilter(t *testing.T) {
	testCases := map[string]struct {
		input         string
		candidates    string
		expected      string
		filter        Filter
		originalNames []string
	}{
		"simple scalar": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: NotSecret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", ""},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"sequence": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
seq:
- oldName1
- oldName2
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: NotSecret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName1", ""},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
seq:
- newName
- oldName2
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "seq"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"mapping": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: NotSecret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", ""},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: newName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "map"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"mapping with namespace": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: oldName
  namespace: oldNs
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
  namespace: oldNs
---
apiVersion: apps/v1
kind: NotSecret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", ""},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: newName
  namespace: oldNs
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "map"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"null value": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: null
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: NotSecret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", ""},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: null
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "map"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			factory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
			referrer, err := factory.FromBytes([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			tc.filter.Referrer = referrer

			resMapFactory := resmap.NewFactory(factory, nil)
			candidatesRes, err := factory.SliceFromBytesWithNames(
				tc.originalNames, []byte(tc.candidates))
			if err != nil {
				t.Fatal(err)
			}

			candidates := resMapFactory.FromResourceSlice(candidatesRes)
			tc.filter.ReferralCandidates = candidates

			if !assert.Equal(t,
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, tc.input, tc.filter))) {
				t.FailNow()
			}
		})
	}
}

func TestNamerefFilterUnhappy(t *testing.T) {
	testCases := map[string]struct {
		input         string
		candidates    string
		expected      string
		filter        Filter
		originalNames []string
	}{
		"multiple match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			expected:      "",
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"no name": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  notName: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			expected:      "",
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			factory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
			referrer, err := factory.FromBytes([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			tc.filter.Referrer = referrer

			resMapFactory := resmap.NewFactory(factory, nil)
			candidatesRes, err := factory.SliceFromBytesWithNames(
				tc.originalNames, []byte(tc.candidates))
			if err != nil {
				t.Fatal(err)
			}

			candidates := resMapFactory.FromResourceSlice(candidatesRes)
			tc.filter.ReferralCandidates = candidates

			_, err = filtertest_test.RunFilterE(t, tc.input, tc.filter)
			if err == nil {
				t.Fatalf("expect an error")
			}
			if tc.expected != "" && !assert.EqualError(t, err, tc.expected) {
				t.FailNow()
			}
		})
	}
}

func TestCandidatesWithDifferentPrefixSuffix(t *testing.T) {
	testCases := map[string]struct {
		input         string
		candidates    string
		expected      string
		filter        Filter
		originalNames []string
		prefix        []string
		suffix        []string
		inputPrefix   string
		inputSuffix   string
		err           bool
	}{
		"prefix match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix1", "prefix2"},
			suffix:        []string{"", "suffix2"},
			inputPrefix:   "prefix1",
			inputSuffix:   "",
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"suffix match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"", "prefix2"},
			suffix:        []string{"suffix1", "suffix2"},
			inputPrefix:   "",
			inputSuffix:   "suffix1",
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"prefix suffix both match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix1", "prefix2"},
			suffix:        []string{"suffix1", "suffix2"},
			inputPrefix:   "prefix1",
			inputSuffix:   "suffix1",
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"multiple match: both": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix", "prefix"},
			suffix:        []string{"suffix", "suffix"},
			inputPrefix:   "prefix",
			inputSuffix:   "suffix",
			expected:      "",
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: true,
		},
		"multiple match: only prefix": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix", "prefix"},
			suffix:        []string{"", ""},
			inputPrefix:   "prefix",
			inputSuffix:   "",
			expected:      "",
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: true,
		},
		"multiple match: only suffix": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"", ""},
			suffix:        []string{"suffix", "suffix"},
			inputPrefix:   "",
			inputSuffix:   "suffix",
			expected:      "",
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: true,
		},
		"no match: neither match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix1", "prefix2"},
			suffix:        []string{"suffix1", "suffix2"},
			inputPrefix:   "prefix",
			inputSuffix:   "suffix",
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"no match: prefix doesn't match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix1", "prefix2"},
			suffix:        []string{"suffix", "suffix"},
			inputPrefix:   "prefix",
			inputSuffix:   "suffix",
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"no match: suffix doesn't match": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName2
`,
			originalNames: []string{"oldName", "oldName"},
			prefix:        []string{"prefix", "prefix"},
			suffix:        []string{"suffix1", "suffix2"},
			inputPrefix:   "prefix",
			inputSuffix:   "suffix",
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			filter: Filter{
				FieldSpec: types.FieldSpec{Path: "ref/name"},
				Target: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			factory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
			referrer, err := factory.FromBytes([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if tc.inputPrefix != "" {
				referrer.AddNamePrefix(tc.inputPrefix)
			}
			if tc.inputSuffix != "" {
				referrer.AddNameSuffix(tc.inputSuffix)
			}
			tc.filter.Referrer = referrer

			resMapFactory := resmap.NewFactory(factory, nil)
			candidatesRes, err := factory.SliceFromBytesWithNames(
				tc.originalNames, []byte(tc.candidates))
			if err != nil {
				t.Fatal(err)
			}
			for i := range candidatesRes {
				if tc.prefix[i] != "" {
					candidatesRes[i].AddNamePrefix(tc.prefix[i])
				}
				if tc.suffix[i] != "" {
					candidatesRes[i].AddNameSuffix(tc.suffix[i])
				}
			}

			candidates := resMapFactory.FromResourceSlice(candidatesRes)
			tc.filter.ReferralCandidates = candidates

			if !tc.err {
				if !assert.Equal(t,
					strings.TrimSpace(tc.expected),
					strings.TrimSpace(
						filtertest_test.RunFilter(t, tc.input, tc.filter))) {
					t.FailNow()
				}
			} else {
				_, err := filtertest_test.RunFilterE(t, tc.input, tc.filter)
				if err == nil {
					t.Fatalf("an error is expected")
				}
			}
		})
	}
}
