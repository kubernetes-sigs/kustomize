package types_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/resid"
	. "sigs.k8s.io/kustomize/api/types"
)

func TestSelectorRegexMatchGvk(t *testing.T) {
	testcases := []struct {
		S        Selector
		G        resid.Gvk
		Expected bool
	}{
		{
			S: Selector{
				Gvk: resid.Gvk{
					Group:   "group",
					Version: "version",
					Kind:    "kind",
				},
			},
			G: resid.Gvk{
				Group:   "group",
				Version: "version",
				Kind:    "kind",
			},
			Expected: true,
		},
		{
			S: Selector{
				Gvk: resid.Gvk{
					Group:   "group",
					Version: "",
					Kind:    "",
				},
			},
			G: resid.Gvk{
				Group:   "group",
				Version: "version",
				Kind:    "kind",
			},
			Expected: true,
		},
		{
			S: Selector{
				Gvk: resid.Gvk{
					Group:   "group",
					Version: "version",
					Kind:    "kind",
				},
			},
			G: resid.Gvk{
				Group:   "group",
				Version: "version",
				Kind:    "",
			},
			Expected: false,
		},
		{
			S: Selector{
				Gvk: resid.Gvk{
					Group:   "group",
					Version: "version",
					Kind:    "kind",
				},
			},
			G: resid.Gvk{
				Group:   "group",
				Version: "version",
				Kind:    "kind2",
			},
			Expected: false,
		},
		{
			S: Selector{
				Gvk: resid.Gvk{
					Group:   "g.*",
					Version: "\\d+",
					Kind:    ".{4}",
				},
			},
			G: resid.Gvk{
				Group:   "group",
				Version: "123",
				Kind:    "abcd",
			},
			Expected: true,
		},
		{
			S: Selector{
				Gvk: resid.Gvk{
					Group:   "g.*",
					Version: "\\d+",
					Kind:    ".{4}",
				},
			},
			G: resid.Gvk{
				Group:   "group",
				Version: "123",
				Kind:    "abc",
			},
			Expected: false,
		},
	}

	for _, tc := range testcases {
		sr, err := NewSelectorRegex(&tc.S)
		if err != nil {
			t.Fatal(err)
		}
		if sr.MatchGvk(tc.G) != tc.Expected {
			t.Fatalf("unexpected result for selector gvk %s and gvk %s",
				tc.S.Gvk.String(), tc.G.String())
		}
	}
}

func TestSelectorRegexMatchName(t *testing.T) {
	testcases := []struct {
		S        Selector
		Name     string
		Expected bool
	}{
		{
			S: Selector{
				Name:      "foo",
				Namespace: "bar",
			},
			Name:     "foo",
			Expected: true,
		},
		{
			S: Selector{
				Name:      "foo",
				Namespace: "bar",
			},
			Name:     "bar",
			Expected: false,
		},
		{
			S: Selector{
				Name: "f.*",
			},
			Name:     "foo",
			Expected: true,
		},
		{
			S: Selector{
				Name: "b.*",
			},
			Name:     "foo",
			Expected: false,
		},
	}
	for _, tc := range testcases {
		sr, err := NewSelectorRegex(&tc.S)
		if err != nil {
			t.Fatal(err)
		}
		if sr.MatchName(tc.Name) != tc.Expected {
			t.Fatalf("unexpected result for selector name %s and name %s",
				tc.S.Name, tc.Name)
		}
	}
}

func TestSelectorRegexMatchNamespace(t *testing.T) {
	testcases := []struct {
		S         Selector
		Namespace string
		Expected  bool
	}{
		{
			S: Selector{
				Name:      "bar",
				Namespace: "foo",
			},
			Namespace: "foo",
			Expected:  true,
		},
		{
			S: Selector{
				Name:      "foo",
				Namespace: "bar",
			},
			Namespace: "foo",
			Expected:  false,
		},
		{
			S: Selector{
				Namespace: "f.*",
			},
			Namespace: "foo",
			Expected:  true,
		},
		{
			S: Selector{
				Namespace: "b.*",
			},
			Namespace: "foo",
			Expected:  false,
		},
	}
	for _, tc := range testcases {
		sr, err := NewSelectorRegex(&tc.S)
		if err != nil {
			t.Fatal(err)
		}
		if sr.MatchNamespace(tc.Namespace) != tc.Expected {
			t.Fatalf("unexpected result for selector namespace %s and namespace %s",
				tc.S.Namespace, tc.Namespace)
		}
	}
}
