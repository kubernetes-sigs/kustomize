package nameref

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func SeqFilter(node *yaml.RNode) (*yaml.RNode, error) {
	if node.YNode().Value == "aaa" {
		node.YNode().SetString("ccc")
	}
	return node, nil
}

func TestApplyFilterToSeq(t *testing.T) {
	fltr := yaml.FilterFunc(SeqFilter)

	testCases := map[string]struct {
		input  string
		expect string
	}{
		"replace in seq": {
			input: `
- aaa
- bbb`,
			expect: `
- ccc
- bbb`,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			node, err := yaml.Parse(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			err = applyFilterToSeq(fltr, node)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t,
				strings.TrimSpace(tc.expect),
				strings.TrimSpace(node.MustString())) {
				t.Fatalf("expect:\n%s\nactual:\n%s",
					strings.TrimSpace(tc.expect),
					strings.TrimSpace(node.MustString()))
			}
		})
	}
}

func TestApplyFilterToSeqUnhappy(t *testing.T) {
	fltr := yaml.FilterFunc(SeqFilter)

	testCases := map[string]struct {
		input string
	}{
		"replace in seq": {
			input: `
aaa`,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			node, err := yaml.Parse(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			err = applyFilterToSeq(fltr, node)
			if err == nil {
				t.Fatalf("expect an error")
			}
		})
	}
}
