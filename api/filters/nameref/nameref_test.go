package nameref

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestNamerefFilter(t *testing.T) {
	testCases := map[string]struct {
		referrerOriginal string
		candidates       string
		referrerFinal    string
		filter           Filter
		originalNames    []string
	}{
		"simple scalar": {
			referrerOriginal: `
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
			originalNames: []string{"oldName", "newName2"},
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"sequence": {
			referrerOriginal: `
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
			originalNames: []string{"oldName1", "newName2"},
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
seq:
- newName
- oldName2
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "seq"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"mapping": {
			referrerOriginal: `
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
			originalNames: []string{"oldName", "newName2"},
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: newName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "map"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"mapping with namespace": {
			referrerOriginal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
  namespace: someNs
map:
  name: oldName
  namespace: someNs
`,
			candidates: `
apiVersion: apps/v1
kind: Secret
metadata:
  name: newName
  namespace: someNs
---
apiVersion: apps/v1
kind: NotSecret
metadata:
  name: newName2
---
apiVersion: apps/v1
kind: Secret
metadata:
  name: thirdName
`,
			originalNames: []string{"oldName", "oldName", "oldName"},
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
  namespace: someNs
map:
  name: newName
  namespace: someNs
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "map"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"null value": {
			referrerOriginal: `
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
			originalNames: []string{"oldName", "newName2"},
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
map:
  name: null
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "map"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			factory := provider.NewDefaultDepProvider().GetResourceFactory()
			referrer, err := factory.FromBytes([]byte(tc.referrerOriginal))
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

			result := filtertest_test.RunFilter(t, tc.referrerOriginal, tc.filter)
			if !assert.Equal(t,
				strings.TrimSpace(tc.referrerFinal),
				strings.TrimSpace(result)) {
				t.FailNow()
			}
		})
	}
}

func TestNamerefFilterUnhappy(t *testing.T) {
	testCases := map[string]struct {
		referrerOriginal string
		candidates       string
		referrerFinal    string
		filter           Filter
		originalNames    []string
	}{
		"multiple match": {
			referrerOriginal: `
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
			referrerFinal: "",
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
		"no name": {
			referrerOriginal: `
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
			referrerFinal: "",
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			factory := provider.NewDefaultDepProvider().GetResourceFactory()
			referrer, err := factory.FromBytes([]byte(tc.referrerOriginal))
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

			_, err = filtertest_test.RunFilterE(t, tc.referrerOriginal, tc.filter)
			if err == nil {
				t.Fatalf("expect an error")
			}
			if tc.referrerFinal != "" && !assert.EqualError(t, err, tc.referrerFinal) {
				t.FailNow()
			}
		})
	}
}

func TestCandidatesWithDifferentPrefixSuffix(t *testing.T) {
	testCases := map[string]struct {
		referrerOriginal string
		candidates       string
		referrerFinal    string
		filter           Filter
		originalNames    []string
		prefix           []string
		suffix           []string
		inputPrefix      string
		inputSuffix      string
		err              bool
	}{
		"prefix match": {
			referrerOriginal: `
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
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"suffix match": {
			referrerOriginal: `
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
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"prefix suffix both match": {
			referrerOriginal: `
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
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: newName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"multiple match: both": {
			referrerOriginal: `
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
			referrerFinal: "",
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: true,
		},
		"multiple match: only prefix": {
			referrerOriginal: `
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
			referrerFinal: "",
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: true,
		},
		"multiple match: only suffix": {
			referrerOriginal: `
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
			referrerFinal: "",
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: true,
		},
		"no match: neither match": {
			referrerOriginal: `
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
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"no match: prefix doesn't match": {
			referrerOriginal: `
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
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    "Secret",
				},
			},
			err: false,
		},
		"no match: suffix doesn't match": {
			referrerOriginal: `
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
			referrerFinal: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
ref:
  name: oldName
`,
			filter: Filter{
				NameFieldToUpdate: types.FieldSpec{Path: "ref/name"},
				ReferralTarget: resid.Gvk{
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
			factory := provider.NewDefaultDepProvider().GetResourceFactory()
			referrer, err := factory.FromBytes([]byte(tc.referrerOriginal))
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
					strings.TrimSpace(tc.referrerFinal),
					strings.TrimSpace(
						filtertest_test.RunFilter(
							t, tc.referrerOriginal, tc.filter))) {
					t.FailNow()
				}
			} else {
				_, err := filtertest_test.RunFilterE(
					t, tc.referrerOriginal, tc.filter)
				if err == nil {
					t.Fatalf("an error is expected")
				}
			}
		})
	}
}
